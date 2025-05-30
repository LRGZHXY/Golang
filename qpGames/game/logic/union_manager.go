package logic

import (
	"common/biz"
	"core/models/entity"
	"fmt"
	"framework/msError"
	"framework/remote"
	"game/component/room"
	"math/rand"
	"sync"
	"time"
)

type UnionManager struct {
	sync.RWMutex
	unionList map[int64]*Union
}

func NewUnionManager() *UnionManager {
	return &UnionManager{
		unionList: make(map[int64]*Union),
	}
}

// GetUnion 根据传入的unionId获取对应的联盟对象
func (u *UnionManager) GetUnion(unionId int64) *Union {
	u.Lock()
	u.Unlock()
	union, ok := u.unionList[unionId]
	if ok {
		return union
	}
	union = NewUnion(u)
	u.unionList[unionId] = union
	return union
}

// CreateRoomId 生成一个不重复的房间id
func (u *UnionManager) CreateRoomId() string {
	//生成一个随机房间号
	roomId := u.genRoomId()
	for _, v := range u.unionList {
		_, ok := v.RoomList[roomId]
		if ok {
			return u.CreateRoomId()
		}
	}
	return roomId
}

// genRoomId 生成一个随机房间号
func (u *UnionManager) genRoomId() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	//房间号是6位数
	roomIdInt := rand.Int63n(999999)
	if roomIdInt < 100000 {
		roomIdInt += 100000
	}
	return fmt.Sprintf("%d", roomIdInt)
}

// GetRoomById 根据房间id查找对应的房间
func (u *UnionManager) GetRoomById(roomId string) *room.Room {
	for _, v := range u.unionList {
		r, ok := v.RoomList[roomId]
		if ok {
			return r
		}
	}
	return nil
}

// JoinRoom 让用户加入指定的房间
func (u *UnionManager) JoinRoom(session *remote.Session, roomId string, data *entity.User) *msError.Error {
	for _, v := range u.unionList {
		r, ok := v.RoomList[roomId]
		if ok {
			return r.JoinRoom(session, data)
		}
	}
	return biz.RoomNotExist
}
