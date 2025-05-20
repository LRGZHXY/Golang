package request

import "game/component/proto"

type RoomMessageReq struct {
	Type proto.RoomMessageType `json:"type"`
	Data RoomMessageData       `json:"data"`
}

type RoomMessageData struct {
	IsReady bool `json:"isReady"` //玩家准备状态
}
