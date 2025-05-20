package net

import (
	"common/logs"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"sync/atomic"
	"time"
)

var cidBase uint64 = 10000

var (
	pongWait             = 10 * time.Second    // 等待客户端Pong响应时间
	writeWait            = 10 * time.Second    // 写入超时时间
	pingInterval         = (pongWait * 9) / 10 // 每隔多久发送一次Ping，确保连接活跃
	maxMessageSize int64 = 1024                // 单条消息最大大小限制
)

type WsConnection struct {
	Cid        string
	Conn       *websocket.Conn
	manager    *Manager
	ReadChan   chan *MsgPack // 用于接收解析后的消息包的通道
	WriteChan  chan []byte   // 写协程从该通道中读取消息并发送给客户端
	Session    *Session
	pingTicker *time.Ticker // 用于定时发送Ping消息
}

// GetSession 返回当前连接绑定的用户会话信息
func (c *WsConnection) GetSession() *Session {
	return c.Session
}

func (c *WsConnection) SendMessage(buf []byte) error {
	c.WriteChan <- buf
	return nil
}

func (c *WsConnection) Close() {
	if c.Conn != nil {
		c.Conn.Close()
	}
	if c.pingTicker != nil {
		c.pingTicker.Stop()
	}
}

func (c *WsConnection) Run() {
	go c.readMessage() // 启动读取消息的协程
	go c.writeMessage()
	//做一些心跳检测 websocket中 ping pong机制
	c.Conn.SetPongHandler(c.PongHandler)
}

// writeMessage 服务端给客户端写消息
func (c *WsConnection) writeMessage() {

	c.pingTicker = time.NewTicker(pingInterval)
	for {
		select {
		case message, ok := <-c.WriteChan:
			if !ok { //通道关闭
				if err := c.Conn.WriteMessage(websocket.CloseMessage, nil); err != nil { //写入消息
					logs.Error("connection closed, %v", err)
				}
				return
			}
			if err := c.Conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
				logs.Error("client[%s] write message err :%v", c.Cid, err)
			}
		case <-c.pingTicker.C:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil { //设置写入超时时间
				logs.Error("client[%s] ping SetWriteDeadline err :%v", c.Cid, err)
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil { //发送ping消息，检查连接是否正常
				logs.Error("client[%s] ping  err :%v", c.Cid, err)
				c.Close()
			}
		}
	}
}

// readMessage 读取客户端发来的消息
func (c *WsConnection) readMessage() {
	defer func() {
		c.manager.removeClient(c)
	}()
	c.Conn.SetReadLimit(maxMessageSize)                                      //设置读取数据最大长度
	if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil { //设置读取超时时间
		logs.Error("SetReadDeadline err:%v", err)
		return
	}
	for {
		messageType, message, err := c.Conn.ReadMessage() //读取消息
		if err != nil {
			break
		}
		//客户端发来的消息是二进制消息
		if messageType == websocket.BinaryMessage {
			if c.ReadChan != nil {
				c.ReadChan <- &MsgPack{ //将消息封装成MsgPack对象，发送到读取通道
					Cid:  c.Cid,
					Body: message,
				}
			}
		} else {
			logs.Error("unsupported message type : %d", messageType)
		}
	}
}

// PongHandler 处理Pong消息
func (c *WsConnection) PongHandler(data string) error {
	if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil { //延长读取超时时间，延长连接活跃期
		return err
	}
	return nil
}

// NewWsConnection 创建一个新的WebSocket连接
func NewWsConnection(conn *websocket.Conn, manager *Manager) *WsConnection {
	cid := fmt.Sprintf("%s-%s-%d", uuid.New().String(), manager.ServerId, atomic.AddUint64(&cidBase, 1))
	return &WsConnection{
		Conn:      conn,
		manager:   manager,
		Cid:       cid, //唯一连接id
		WriteChan: make(chan []byte, 1024),
		ReadChan:  manager.ClientReadChan,
		Session:   NewSession(cid),
	}
}
