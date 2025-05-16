package net

import "sync"

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
