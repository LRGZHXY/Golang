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
	pongWait             = 10 * time.Second
	writeWait            = 10 * time.Second
	pingInterval         = (pongWait * 9) / 10
	maxMessageSize int64 = 1024
)

type WsConnection struct {
	Cid       string
	Conn      *websocket.Conn
	manager   *Manager
	ReadChan  chan *MsgPack
	WriteChan chan []byte
}

func (c *WsConnection) SendMessage(buf []byte) error {
	c.WriteChan <- buf
	return nil
}

func (c WsConnection) Close() {
	if c.Conn != nil {
		c.Conn.Close()
	}
}

func (c WsConnection) Run() {
	go c.readMessage()
	go c.writeMessage()
	//做一些心跳检测 websocket中 ping pong机制
	c.Conn.SetPongHandler(c.PongHandler)
}

// writeMessage 服务端给客户端写消息
func (c *WsConnection) writeMessage() {
	ticker := time.NewTicker(pingInterval)
	for {
		select {
		case message, ok := <-c.WriteChan:
			if !ok { //通道关闭
				if err := c.Conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
					logs.Error("connection closed,%v", err)
				}
				return
			}
			if err := c.Conn.WriteMessage(websocket.BinaryMessage, message); err != nil { //写入消息
				logs.Error("client[%s] write message failed,err:%v", c.Cid, err)
			}
		case <-ticker.C:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil { //设置写入超时时间
				logs.Error("client[%s] ping SetWriteDeadline err:%v", c.Cid, err)
			}
			//logs.Info("ping...")
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil { //发送ping消息，检查连接是否正常
				logs.Error("client[%s] ping err:%v", c.Cid, err)
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
			logs.Error("unsupported message type:%d", messageType)
		}
	}
}

func (c *WsConnection) PongHandler(data string) error {
	//logs.Info("pong...")
	if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil { //延长读取超时时间，延长连接活跃期
		return err
	}
	return nil
}

func NewWsConnection(conn *websocket.Conn, manager *Manager) *WsConnection {
	cid := fmt.Sprintf("%s-%s-%d", uuid.New().String(), manager.ServerId, atomic.AddUint64(&cidBase, 1))
	return &WsConnection{
		Conn:      conn,
		manager:   manager,
		Cid:       cid, //唯一连接id
		WriteChan: make(chan []byte, 1024),
		ReadChan:  manager.ClientReadChan,
	}
}
