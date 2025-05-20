package logic

import (
	"core/models/entity"
	"core/service"
	"framework/msError"
	"framework/remote"
	"game/component/room"
	"game/models/request"
	"sync"
)

type Union struct {
	sync.RWMutex
	Id       int64
	m        *UnionManager
	RoomList map[string]*room.Room
}

func (u *Union) CreateRoom(service *service.UserService, session *remote.Session, req request.CreateRoomReq, userData *entity.User) *msError.Error {
	//1. 需要创建一个房间 生成一个房间号
	roomId := u.m.CreateRoomId()
	newRoom := room.NewRoom(roomId, req.UnionID, req.GameRule, u)
	u.RoomList[roomId] = newRoom                    //将新房间加入联盟的房间列表
	return newRoom.UserEntryRoom(session, userData) //让用户加入新建的房间
}

func (u *Union) DismissRoom(roomId string) {
	u.Lock()
	defer u.Unlock()
	delete(u.RoomList, roomId) //删除指定房间ID对应的房间，表示解散该房间
}

func NewUnion(m *UnionManager) *Union {
	return &Union{
		RoomList: make(map[string]*room.Room),
		m:        m,
	}
}
