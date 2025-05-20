package net

import "sync"

// Session 存储会话信息
type Session struct {
	sync.RWMutex
	Cid  string // Connection ID，连接标识
	Uid  string // User ID，用户唯一标识
	data map[string]any
}

// NewSession 创建一个新的会话
func NewSession(cid string) *Session {
	return &Session{
		Cid:  cid,
		data: make(map[string]any),
	}
}

// Put 向data中插入一个键值对
func (s *Session) Put(key string, v any) {
	s.Lock()
	defer s.Unlock()
	s.data[key] = v
}

// Get 从data中根据key获取对应的值
func (s *Session) Get(key string) (any, bool) {
	s.RLock()
	defer s.RUnlock()
	v, ok := s.data[key]
	return v, ok
}

// SetData 把新的数据批量添加到data中
func (s *Session) SetData(uid string, data map[string]any) {
	s.Lock()
	defer s.Unlock()
	if s.Uid == uid {
		for k, v := range data {
			s.data[k] = v
		}
	}
}
