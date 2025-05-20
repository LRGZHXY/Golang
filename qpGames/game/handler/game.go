package handler

import (
	"common"
	"common/biz"
	"core/repo"
	"core/service"
	"encoding/json"
	"fmt"
	"framework/remote"
	"game/logic"
	"game/models/request"
)

type GameHandler struct {
	um          *logic.UnionManager
	userService *service.UserService
}

// RoomMessageNotify 处理用户发往房间的消息请求
func (h *GameHandler) RoomMessageNotify(session *remote.Session, msg []byte) any {
	if len(session.GetUid()) <= 0 { //是否存在有效的用户ID
		return common.F(biz.InvalidUsers)
	}
	var req request.RoomMessageReq
	if err := json.Unmarshal(msg, &req); err != nil { //解析请求体json
		return common.F(biz.RequestDataError)
	}
	//检查房间id是否存在于session中
	roomId, ok := session.Get("roomId")
	if !ok {
		return common.F(biz.NotInRoom)
	}
	//通过roomId获取房间对象
	rm := h.um.GetRoomById(fmt.Sprintf("%v", roomId))
	if rm == nil {
		return common.F(biz.NotInRoom)
	}
	rm.RoomMessageHandle(session, req)
	return nil
}

// GameMessageNotify 处理客户端发送来的游戏逻辑类消息
func (h *GameHandler) GameMessageNotify(session *remote.Session, msg []byte) any {
	//用户是否已登录
	if len(session.GetUid()) <= 0 {
		return common.F(biz.InvalidUsers)
	}
	//获取房间id
	roomId, ok := session.Get("roomId")
	if !ok {
		return common.F(biz.NotInRoom)
	}
	//获取房间对象
	rm := h.um.GetRoomById(fmt.Sprintf("%v", roomId))
	if rm == nil {
		return common.F(biz.NotInRoom)
	}
	rm.GameMessageHandle(session, msg)
	return nil
}

func NewGameHandler(r *repo.Manager, um *logic.UnionManager) *GameHandler {
	return &GameHandler{
		um:          um,
		userService: service.NewUserService(r),
	}
}
