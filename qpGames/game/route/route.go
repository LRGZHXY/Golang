package route

import (
	"core/repo"
	"framework/node"
	"game/handler"
	"game/logic"
)

func Register(r *repo.Manager) node.LogicHandler {
	handlers := make(node.LogicHandler)
	um := logic.NewUnionManager()
	unionHandler := handler.NewUnionHandler(r, um)
	handlers["unionHandler.createRoom"] = unionHandler.CreateRoom //创建房间
	handlers["unionHandler.joinRoom"] = unionHandler.JoinRoom     //加入房间
	gameHandler := handler.NewGameHandler(r, um)
	handlers["gameHandler.roomMessageNotify"] = gameHandler.RoomMessageNotify //房间消息
	handlers["gameHandler.gameMessageNotify"] = gameHandler.GameMessageNotify //游戏消息
	return handlers
}
