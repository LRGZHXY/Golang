package net

import (
	"common/logs"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
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
}

func (m *Manager) Run(addr string) {
	go m.clientReadChanHandler()
	http.HandleFunc("/", m.serveWS)
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
	logs.Info("receiver message:%v", string(body.Body))
}

func (m *Manager) Close() {
	for cid, v := range m.clients {
		v.Close()
		delete(m.clients, cid)
	}
}

func NewManager() *Manager {
	return &Manager{
		ClientReadChan: make(chan *MsgPack, 1024),
		clients:        make(map[string]Connection),
	}
}
