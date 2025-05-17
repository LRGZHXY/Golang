package net

import "sync"

// Session 存储会话信息
type Session struct {
	sync.RWMutex
	Cid  string
	Uid  string
	data map[string]any
}

func NewSession(cid string) *Session {
	return &Session{
		Cid: cid,
	}
}
