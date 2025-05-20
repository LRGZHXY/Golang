package route

import (
	"connector/handler"
	"core/repo"
	"framework/net"
)

// Register
func Register(r *repo.Manager) net.LogicHandler {
	handlers := make(net.LogicHandler)
	entryHandler := handler.NewEntryHandler(r)
	handlers["entryHandler.entry"] = entryHandler.Entry //将entryHandler.Entry方法注册到路由entryHandler.entry

	return handlers
}
