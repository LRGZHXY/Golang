package net

import (
	"common/logs"
	"encoding/json"
	"errors"
	"fmt"
	"framework/game"
	"framework/protocol"
	"framework/remote"
	"github.com/gorilla/websocket"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	websocketUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { //允许跨域
			return true
		},
		ReadBufferSize:  1024, //读缓冲区大小
		WriteBufferSize: 1024,
	}
)

type CheckOriginHandler func(r *http.Request) bool
type Manager struct {
	sync.RWMutex       //读写锁
	websocketUpgrader  *websocket.Upgrader
	ServerId           string
	CheckOriginHandler CheckOriginHandler
	clients            map[string]Connection
	ClientReadChan     chan *MsgPack
	handlers           map[protocol.PackageType]EventHandler
	ConnectorHandlers  LogicHandler
	RemoteReadChan     chan []byte
	RemoteCli          remote.Client
}

type HandlerFunc func(session *Session, body []byte) (any, error)
type LogicHandler map[string]HandlerFunc
type EventHandler func(packet *protocol.Packet, c Connection) error

func (m *Manager) Run(addr string) {
	go m.clientReadChanHandler()
	go m.remoteReadChanHandler()
	http.HandleFunc("/", m.serveWS)
	//设置不同的消息处理器
	m.setupEventHandlers()
	logs.Fatal("connector listen serve err:%v", http.ListenAndServe(addr, nil))
}

// serveWS 接收HTTP请求、升级为WebSocket连接
func (m *Manager) serveWS(w http.ResponseWriter, r *http.Request) {
	//升级为websocket连接
	if m.websocketUpgrader == nil {
		m.websocketUpgrader = &websocketUpgrader
	}
	wsConn, err := m.websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		logs.Error("websocket upgrade err:%v", err)
		return
	}
	client := NewWsConnection(wsConn, m) //创建websocket连接
	m.addClient(client)
	client.Run()
}

// addClient 将websocket客户端添加到clients映射中
func (m *Manager) addClient(client *WsConnection) {
	m.Lock() //加锁确保多个并发连接不会同时修改clients映射
	defer m.Unlock()
	m.clients[client.Cid] = client
}

func (m *Manager) removeClient(wc *WsConnection) {
	for cid, c := range m.clients {
		if cid == wc.Cid {
			c.Close()
			delete(m.clients, cid)
		}
	}
}

func (m *Manager) clientReadChanHandler() {
	for {
		select {
		case body, ok := <-m.ClientReadChan: //读取客户端发来的消息
			if ok {
				m.decodeClientPack(body)
			}
		}
	}
}

func (m *Manager) decodeClientPack(body *MsgPack) {
	//解析协议
	packet, err := protocol.Decode(body.Body) //解码
	if err != nil {
		logs.Error("decode message err:%v", err)
		return
	}
	if err := m.routeEvent(packet, body.Cid); err != nil {
		logs.Error("routeEvent err:%v", err)
	}
}

func (m *Manager) Close() {
	for cid, v := range m.clients {
		v.Close()
		delete(m.clients, cid)
	}
}

// routeEvent 根据packet.type来做不同的处理
func (m *Manager) routeEvent(packet *protocol.Packet, cid string) error {
	conn, ok := m.clients[cid]
	if ok {
		handler, ok := m.handlers[packet.Type]
		if ok {
			return handler(packet, conn)
		} else {
			return errors.New("no packetType found")
		}
	}
	return errors.New("no client found")
}

func (m *Manager) setupEventHandlers() {
	m.handlers[protocol.Handshake] = m.HandshakeHandler
	m.handlers[protocol.HandshakeAck] = m.HandshakeAckHandler
	m.handlers[protocol.Heartbeat] = m.HeartbeatHandler
	m.handlers[protocol.Data] = m.MessageHandler
	m.handlers[protocol.Kick] = m.KickHandler
}

// HandshakeHandler 处理握手请求
func (m *Manager) HandshakeHandler(packet *protocol.Packet, c Connection) error {
	res := protocol.HandshakeResponse{
		Code: 200,
		Sys: protocol.Sys{
			Heartbeat: 3, //3秒发一次心跳
		},
	}
	data, _ := json.Marshal(res)                   //将握手响应结构体序列化为JSON字节流
	buf, err := protocol.Encode(packet.Type, data) //编码
	if err != nil {
		logs.Error("encode packet err:%v", err)
		return err
	}
	return c.SendMessage(buf) //把握手响应数据发送给客户端
}

// HandshakeAckHandler 处理握手确认
func (m *Manager) HandshakeAckHandler(packet *protocol.Packet, c Connection) error {
	logs.Info("receiver handshake ack message...")
	return nil
}

// HeartbeatHandler 处理心跳
func (m *Manager) HeartbeatHandler(packet *protocol.Packet, c Connection) error {
	logs.Info("receiver heartbeat message:%v", packet.Type)
	var res []byte
	data, _ := json.Marshal(res)
	buf, err := protocol.Encode(packet.Type, data)
	if err != nil {
		logs.Error("encode packet err:%v", err)
		return err
	}
	return c.SendMessage(buf)
}

// MessageHandler 本地消息处理
func (m *Manager) MessageHandler(packet *protocol.Packet, c Connection) error {
	message := packet.MessageBody() //提取消息体
	logs.Info("receiver message body,type=%v,data:%v",
		message.Type, message.Route, string(message.Data))
	routeStr := message.Route
	routers := strings.Split(routeStr, ".") //按.分割消息路由
	if len(routers) != 3 {
		return errors.New("router unsupported")
	}
	serverType := routers[0]
	handlerMethod := fmt.Sprintf("%s.%s", routers[1], routers[2])
	connectorConfig := game.Conf.GetConnectorByServerType(serverType) //获取当前serverType的connector配置
	if connectorConfig != nil {
		//本地connector服务处理器
		handler, ok := m.ConnectorHandlers[handlerMethod]
		if ok { //本地可以处理的请求
			data, err := handler(c.GetSession(), message.Data)
			if err != nil {
				return err
			}
			marshal, _ := json.Marshal(data) //将data序列化为JSON字节数组
			message.Type = protocol.Response //设置消息类型为响应
			message.Data = marshal
			encode, err := protocol.MessageEncode(message) //将消息编码为字节数组
			if err != nil {
				return err
			}
			res, err := protocol.Encode(packet.Type, encode)
			if err != nil {
				return err
			}
			return c.SendMessage(res) //把结果返回给客户端
		}
	} else {
		//nats 远端调用
		dst, err := m.selectDst(serverType) //选择目标服务器
		if err != nil {
			logs.Error("remote send msg selectDst err:%v", err)
			return err
		}
		msg := &remote.Msg{
			Cid:         c.GetSession().Cid, //客户端连接
			Uid:         c.GetSession().Uid,
			Src:         m.ServerId, //当前服务器ID（发送方）
			Dst:         dst,        //目标服务器ID（接收方）
			Router:      handlerMethod,
			Body:        message,
			SessionData: c.GetSession().data,
		}
		data, _ := json.Marshal(msg)
		err = m.RemoteCli.SendMsg(dst, data) //通过nats发送消息到目标服务器
		if err != nil {
			logs.Error("remote send msg err:%v", err)
			return err
		}
	}
	return nil
}

// KickHandler 处理踢出通知
func (m *Manager) KickHandler(packet *protocol.Packet, c Connection) error {
	logs.Info("receiver kick message...")
	return nil
}

func (m *Manager) remoteReadChanHandler() {
	for {
		select {
		case msg := <-m.RemoteReadChan: //读取nats发来的消息
			logs.Info("sub nats msg:%v", string(msg))
		}
	}
}

// selectDst 从服务器列表随机选择一个目标服务器
func (m *Manager) selectDst(serverType string) (string, error) {
	serversConfig, ok := game.Conf.ServersConf.TypeServer[serverType] //查找该类型服务器列表
	if !ok {
		return "", errors.New("no server found")
	}
	rand.New(rand.NewSource(time.Now().UnixNano()))
	index := rand.Intn(len(serversConfig)) //在服务器列表长度范围内生成一个随机下标
	return serversConfig[index].ID, nil
}

func NewManager() *Manager {
	return &Manager{
		ClientReadChan: make(chan *MsgPack, 1024),
		clients:        make(map[string]Connection),
		handlers:       make(map[protocol.PackageType]EventHandler),
		RemoteReadChan: make(chan []byte, 1024),
	}
}
