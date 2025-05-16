package route

import (
	"connector/handler"
	"framework/net"
)

// Register
func Register() net.LogicHandler {
	handlers := make(net.LogicHandler)
	entryHandler := handler.NewEntryHandler()
	handlers["entryHandler.entry"] = entryHandler.Entry //将entryHandler.Entry方法注册到路由entryHandler.entry
	return handlers
}
