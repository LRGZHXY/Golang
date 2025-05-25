package room

import (
	"common/logs"
	"core/models/entity"
	"framework/msError"
	"framework/remote"
	"game/component/base"
	"game/component/mj"
	"game/component/proto"
	"game/component/sz"
	"game/models/request"
	"sync"
	"time"
)

type Room struct {
	sync.RWMutex
	Id            string
	unionID       int64
	gameRule      proto.GameRule
	users         map[string]*proto.RoomUser
	RoomCreator   *proto.RoomCreator
	GameFrame     GameFrame
	kickSchedules map[string]*time.Timer //踢出玩家定时器
	union         base.UnionBase
	roomDismissed bool //房间是否已解散
	gameStarted   bool //游戏是否已开始
	askDismiss    map[int]struct{}
}

// UserReady 用户准备
func (r *Room) UserReady(uid string, session *remote.Session) {
	r.userReady(uid, session)
}

// EndGame 游戏结束
func (r *Room) EndGame(session *remote.Session) {
	r.gameStarted = false
	for k := range r.users {
		r.users[k].UserStatus = proto.None
	}
}

// UserEntryRoom 用户进入房间
func (r *Room) UserEntryRoom(session *remote.Session, data *entity.User) *msError.Error {
	r.RoomCreator = &proto.RoomCreator{
		Uid: data.Uid,
	}
	if r.unionID == 1 {
		r.RoomCreator.CreatorType = proto.UserCreatorType
	} else {
		r.RoomCreator.CreatorType = proto.UnionCreatorType
	}
	//最多6人参加 0-5有6个号
	chairID := r.getEmptyChairID()
	_, ok := r.users[data.Uid]
	if !ok {
		r.users[data.Uid] = proto.ToRoomUser(data, chairID)
	}
	//2. 将房间号 推送给客户端 更新数据库 当前房间号存储起来
	r.UpdateUserInfoRoomPush(session, data.Uid)
	session.Put("roomId", r.Id)
	//3. 将游戏类型 推送给客户端 （用户进入游戏的推送）
	r.SelfEntryRoomPush(session, data.Uid)
	//4.告诉其他人 此用户进入房间了
	r.OtherUserEntryRoomPush(session, data.Uid)
	go r.addKickScheduleEvent(session, data.Uid)
	return nil
}

// UpdateUserInfoRoomPush 更新用户房间信息
func (r *Room) UpdateUserInfoRoomPush(session *remote.Session, uid string) {
	//{roomID: '336842', pushRouter: 'UpdateUserInfoPush'}
	pushMsg := map[string]any{
		"roomID":     r.Id,
		"pushRouter": "UpdateUserInfoPush",
	}
	session.Push([]string{uid}, pushMsg, "ServerMessagePush") //推送消息
}

// SelfEntryRoomPush 告诉用户自己成功进入房间
func (r *Room) SelfEntryRoomPush(session *remote.Session, uid string) {
	//{gameType: 1, pushRouter: 'SelfEntryRoomPush'}
	pushMsg := map[string]any{
		"gameType":   r.gameRule.GameType,
		"pushRouter": "SelfEntryRoomPush",
	}
	session.Push([]string{uid}, pushMsg, "ServerMessagePush")
}

func (r *Room) RoomMessageHandle(session *remote.Session, req request.RoomMessageReq) {
	if req.Type == proto.UserReadyNotify {
		r.userReady(session.GetUid(), session) //准备
	}
	if req.Type == proto.GetRoomSceneInfoNotify {
		r.getRoomSceneInfoPush(session) //推送房间场景信息
	}
	if req.Type == proto.AskForDismissNotify {
		r.askForDismiss(session, req.Data.IsExit) //解散房间
	}
}

// getRoomSceneInfoPush 获取房间场景信息
func (r *Room) getRoomSceneInfoPush(session *remote.Session) {
	//
	userInfoArr := make([]*proto.RoomUser, 0)
	for _, v := range r.users {
		userInfoArr = append(userInfoArr, v)
	}
	data := map[string]any{
		"type":       proto.GetRoomSceneInfoPush,
		"pushRouter": "RoomMessagePush",
		"data": map[string]any{
			"roomID":          r.Id,
			"roomCreatorInfo": r.RoomCreator,
			"gameRule":        r.gameRule,
			"roomUserInfoArr": userInfoArr,
			"gameData":        r.GameFrame.GetGameData(session),
		},
	}
	session.Push([]string{session.GetUid()}, data, "ServerMessagePush")
}

// addKickScheduleEvent 如果用户在30秒内没有准备，则自动将该用户踢出房间
func (r *Room) addKickScheduleEvent(session *remote.Session, uid string) {
	r.Lock()
	defer r.Unlock()
	t, ok := r.kickSchedules[uid]
	if ok { //如果已经存在踢出定时器，则先取消之前的定时任务
		t.Stop()
		delete(r.kickSchedules, uid)
	}
	r.kickSchedules[uid] = time.AfterFunc(30*time.Second, func() {
		logs.Info("kick 定时执行，代表 用户长时间未准备,uid=%v", uid)
		//删除定时器
		timer, ok := r.kickSchedules[uid]
		if ok {
			timer.Stop()
		}
		delete(r.kickSchedules, uid)
		//需要判断用户是否该踢出
		user, ok := r.users[uid]
		if ok {
			if user.UserStatus < proto.Ready {
				r.kickUser(user, session)
				//踢出房间之后，如果房间内没有用户了，则解散房间
				if len(r.users) == 0 {
					r.dismissRoom()
				}
			}
		}
	})
}

func (r *Room) ServerMessagePush(users []string, data any, session *remote.Session) {
	session.Push(users, data, "ServerMessagePush")
}

func (r *Room) kickUser(user *proto.RoomUser, session *remote.Session) {
	//将roomId设为空
	r.ServerMessagePush([]string{user.UserInfo.Uid}, proto.UpdateUserInfoPush(""), session)
	//通知其他人此用户离开房间
	users := make([]string, 0)
	for _, v := range r.users {
		users = append(users, v.UserInfo.Uid)
	}
	r.ServerMessagePush(users, proto.UserLeaveRoomPushData(user), session)
	delete(r.users, user.UserInfo.Uid) //删除该用户的房间信息
}

// dismissRoom 解散房间
func (r *Room) dismissRoom() {
	if r.TryLock() { //如果重复枷锁会导致阻塞，解散后不能重新创建房间
		r.Lock()
		defer r.Unlock()
	}
	if r.roomDismissed {
		return
	}
	r.roomDismissed = true
	r.cancelAllScheduler() //取消房间内所有的定时任务
	r.union.DismissRoom(r.Id)
}

// cancelAllScheduler 将房间所有的任务都取消掉
func (r *Room) cancelAllScheduler() {
	for uid, v := range r.kickSchedules {
		v.Stop()
		delete(r.kickSchedules, uid)
	}
}

// userReady 用户准备
func (r *Room) userReady(uid string, session *remote.Session) {
	user, ok := r.users[uid]
	if !ok {
		return
	}
	user.UserStatus = proto.Ready
	timer, ok := r.kickSchedules[uid]
	if ok { //用户一旦准备，就不应该再被自动踢出，所以停止并删除踢人定时器
		timer.Stop()
		delete(r.kickSchedules, uid)
	}
	allUsers := r.AllUsers()
	r.ServerMessagePush(allUsers, proto.UserReadyPushData(user.ChairID), session) //给所有房间内用户推送某用户已准备状态
	if r.IsStartGame() {                                                          //判断是否开局
		r.startGame(session, user)
	}
}

func (r *Room) JoinRoom(session *remote.Session, data *entity.User) *msError.Error {

	return r.UserEntryRoom(session, data)
}

// OtherUserEntryRoomPush 当某个用户进入房间时，给房间其他的用户推送有新用户加入房间的消息
func (r *Room) OtherUserEntryRoomPush(session *remote.Session, uid string) {
	others := make([]string, 0)
	for _, v := range r.users {
		if v.UserInfo.Uid != uid {
			others = append(others, v.UserInfo.Uid)
		}
	}
	user, ok := r.users[uid]
	if ok {
		r.ServerMessagePush(others, proto.OtherUserEntryRoomPushData(user), session)
	}
}

// AllUsers 获取房间内所有用户uid
func (r *Room) AllUsers() []string {
	users := make([]string, 0)
	for _, v := range r.users {
		users = append(users, v.UserInfo.Uid)
	}
	return users
}

// getEmptyChairID 分配一个房间中还未被占用的座位号
func (r *Room) getEmptyChairID() int {
	if len(r.users) == 0 {
		return 0
	}
	r.Lock()
	defer r.Unlock()
	chairID := 0
	for _, v := range r.users {
		if v.ChairID == chairID {
			//座位号被占用了
			chairID++
		}
	}
	return chairID
}

// IsStartGame 判断房间是否满足开始游戏的条件
func (r *Room) IsStartGame() bool {
	//房间内准备的人数 已经大于等于 最小开始游戏人数
	userReadyCount := 0
	for _, v := range r.users {
		if v.UserStatus == proto.Ready {
			userReadyCount++
		}
	}
	if r.gameRule.GameType == int(proto.HongZhong) {
		if len(r.users) == userReadyCount && userReadyCount >= r.gameRule.MaxPlayerCount {
			return true
		}
	}
	if len(r.users) == userReadyCount && userReadyCount >= r.gameRule.MinPlayerCount {
		return true
	}
	return false
}

// startGame 开始游戏
func (r *Room) startGame(session *remote.Session, user *proto.RoomUser) {
	if r.gameStarted {
		return
	}
	r.gameStarted = true
	for _, v := range r.users {
		v.UserStatus = proto.Playing
	}
	r.GameFrame.StartGame(session, user)
}

func NewRoom(id string, unionID int64, rule proto.GameRule, u base.UnionBase) *Room {
	r := &Room{
		Id:            id,
		unionID:       unionID,
		gameRule:      rule,
		users:         make(map[string]*proto.RoomUser),
		kickSchedules: make(map[string]*time.Timer),
		union:         u,
	}
	if rule.GameType == int(proto.PinSanZhang) {
		r.GameFrame = sz.NewGameFrame(rule, r)
	}
	if rule.GameType == int(proto.HongZhong) {
		r.GameFrame = mj.NewGameFrame(rule, r) // !!!!!
	}
	return r
}

func (r *Room) GetUsers() map[string]*proto.RoomUser {
	return r.users
}
func (r *Room) GetId() string {
	return r.Id
}
func (r *Room) GameMessageHandle(session *remote.Session, msg []byte) {
	//需要游戏去处理具体的消息
	user, ok := r.users[session.GetUid()]
	if !ok {
		return
	}
	r.GameFrame.GameMessageHandle(user, session, msg)
}

// askForDismiss 解散房间
func (r *Room) askForDismiss(session *remote.Session, exist bool) {
	r.Lock()
	defer r.Unlock()
	if exist { //同意解散
		if r.askDismiss == nil {
			r.askDismiss = make(map[int]struct{})
		}
		user := r.users[session.GetUid()]
		r.askDismiss[user.ChairID] = struct{}{} //当前玩家投票支持解散

		nameArr := make([]string, len(r.users))
		chairIDArr := make([]any, len(r.users))
		avatarArr := make([]string, len(r.users))
		onlinrArr := make([]bool, len(r.users))
		for _, v := range r.users {
			nameArr[v.ChairID] = v.UserInfo.Nickname
			avatarArr[v.ChairID] = v.UserInfo.Avatar
			_, ok := r.askDismiss[v.ChairID]
			if ok {
				chairIDArr[v.ChairID] = true
			}
			onlinrArr[v.ChairID] = true
		}
		data := proto.DismissPushData{
			NameArr:    nameArr,
			ChairIDArr: chairIDArr,
			AskChairId: user.ChairID,
			OnlineArr:  onlinrArr,
			AvatarArr:  avatarArr,
			Tm:         30, //倒计时
		}
		r.sendData(proto.AskForDismissPushData(&data), session)
		if len(r.askDismiss) == len(r.users) { //所有人都同意解散
			for _, v := range r.users {
				r.kickUser(v, session)
			}
			if len(r.users) == 0 {
				r.dismissRoom() //解散房间
			}
		}

	} else { //不同意解散
		user := r.users[session.GetUid()]
		nameArr := make([]string, len(r.users))
		chairIDArr := make([]any, len(r.users))
		avatarArr := make([]string, len(r.users))
		onlinrArr := make([]bool, len(r.users))
		for _, v := range r.users {
			nameArr[v.ChairID] = v.UserInfo.Nickname
			avatarArr[v.ChairID] = v.UserInfo.Avatar
			_, ok := r.askDismiss[v.ChairID]
			if ok {
				chairIDArr[v.ChairID] = true
			}
			onlinrArr[v.ChairID] = true
		}
		data := proto.DismissPushData{
			NameArr:    nameArr,
			ChairIDArr: chairIDArr,
			AskChairId: user.ChairID,
			OnlineArr:  onlinrArr,
			AvatarArr:  avatarArr,
			Tm:         30, //倒计时
		}
		r.sendData(proto.AskForDismissPushData(&data), session)
	}
}

// sendData 向房间内所有用户推送消息
func (r *Room) sendData(data any, session *remote.Session) {
	r.ServerMessagePush(r.AllUsers(), data, session)
}
