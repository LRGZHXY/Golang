package route

import (
	"core/repo"
	"framework/node"
	"hall/handler"
)

// Register
func Register(r *repo.Manager) node.LogicHandler {
	handlers := make(node.LogicHandler)
	userHandler := handler.NewUserHandler(r)
	//将userHandler.UpdateUserAddress方法注册到路由userHandler.updateUserAddress
	handlers["userHandler.updateUserAddress"] = userHandler.UpdateUserAddress

	return handlers
}
