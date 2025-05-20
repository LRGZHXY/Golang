package proto

import "core/models/entity"

// RoomCreator 房间创建者信息
type RoomCreator struct {
	Uid         string      `json:"uid"`
	CreatorType CreatorType `json:"creatorType"`
}

type CreatorType int

const (
	UserCreatorType  CreatorType = 1 // 用户创建
	UnionCreatorType             = 2 // 联盟创建
)

// RoomUser 房间中的用户信息
type RoomUser struct {
	UserInfo   UserInfo   `json:"userInfo"`
	ChairID    int        `json:"chairID"`    // 座位ID
	UserStatus UserStatus `json:"userStatus"` // 用户状态
}

// UserInfo 用户详细信息
type UserInfo struct {
	Uid          string `json:"uid"`
	Nickname     string `json:"nickname"`
	Avatar       string `json:"avatar"`
	Gold         int64  `json:"gold"`
	FrontendId   string `json:"frontendId"`
	Address      string `json:"address"`
	Location     string `json:"location"`
	LastLoginIP  string `json:"lastLoginIP"`
	Sex          int    `json:"sex"`
	Score        int    `json:"score"`
	SpreaderID   string `json:"spreaderID"` //推广ID
	ProhibitGame bool   `json:"prohibitGame"`
	RoomID       string `json:"roomID"`
}

type UserStatus int

const (
	None    UserStatus = 0 // 初始状态
	Ready              = 1 // 准备状态
	Playing            = 2 // 游戏中
	Offline            = 4 // 离线
	Dismiss            = 8 // 解散
)

// ToRoomUser 将实体用户转换为房间用户
func ToRoomUser(data *entity.User, chairID int) *RoomUser {
	userInfo := UserInfo{
		Uid:      data.Uid,
		Nickname: data.Nickname,
		Avatar:   data.Avatar,
		Gold:     data.Gold,
		Sex:      data.Sex,
		Address:  data.Address,
	}
	return &RoomUser{
		UserInfo:   userInfo,
		ChairID:    chairID,
		UserStatus: None,
	}
}
