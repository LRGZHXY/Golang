package repo

import "common/database"

type Manager struct {
	Mongo *database.MongoManager
	Redis *database.RedisManager
}

// New 封装MongoDB和Redis管理器
func New() *Manager {
	return &Manager{
		Mongo: database.NewMongo(),
		Redis: database.NewRedis(),
	}
}
