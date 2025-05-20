package net

type Connection interface {
	Close()
	SendMessage(buf []byte) error
	GetSession() *Session
}

// 消息包
type MsgPack struct {
	Cid  string
	Body []byte
}
