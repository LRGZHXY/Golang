package sz

import (
	"common/logs"
	"common/utils"
	"encoding/json"
	"framework/remote"
	"game/component/base"
	"game/component/proto"
	"github.com/jinzhu/copier"
	"time"
)

type GameFrame struct {
	r          base.RoomFrame
	gameRule   proto.GameRule
	gameData   *GameData
	logic      *Logic
	gameResult *GameResult
}

func (g *GameFrame) GameMessageHandle(user *proto.RoomUser, session *remote.Session, msg []byte) {
	//1. 解析参数
	var req MessageReq
	json.Unmarshal(msg, &req)
	//2. 根据不同的类型 触发不同的操作
	if req.Type == GameLookNotify { //看牌
		g.onGameLook(user, session, req.Data.Cuopai)
	} else if req.Type == GamePourScoreNotify { //下注
		g.onGamePourScore(user, session, req.Data.Score, req.Data.Type)
	} else if req.Type == GameCompareNotify { //比牌
		g.onGameCompare(user, session, req.Data.ChairID)
	} else if req.Type == GameAbandonNotify { //弃牌
		g.onGameAbandon(user, session)
	}
}

func NewGameFrame(rule proto.GameRule, r base.RoomFrame) *GameFrame {
	gameData := initGameData(rule)
	return &GameFrame{
		r:        r,
		gameRule: rule,
		gameData: gameData,
		logic:    NewLogic(),
	}
}

// initGameData 初始化游戏数据
func initGameData(rule proto.GameRule) *GameData {
	g := &GameData{
		GameType:   GameType(rule.GameFrameType),
		BaseScore:  rule.BaseScore,
		ChairCount: rule.MaxPlayerCount,
	}
	g.PourScores = make([][]int, g.ChairCount)                                                      // 每个玩家的下注记录（多轮下注）
	g.HandCards = make([][]int, g.ChairCount)                                                       // 每个玩家的手牌（牌型示例）
	g.LookCards = make([]int, g.ChairCount)                                                         // 玩家是否看牌的标记（0/1）
	g.CurScores = make([]int, g.ChairCount)                                                         // 当前得分
	g.UserStatusArray = make([]UserStatus, g.ChairCount)                                            // 玩家状态，比如准备、游戏中、弃牌等
	g.UserTrustArray = []bool{false, false, false, false, false, false, false, false, false, false} // 托管状态（是否托管）
	g.Loser = make([]int, 0)                                                                        // 输家玩家座位ID列表
	g.Winner = make([]int, 0)                                                                       // 胜利玩家座位ID列表
	return g
}

func (g *GameFrame) GetGameData(session *remote.Session) any {
	user := g.r.GetUsers()[session.GetUid()]
	//判断当前用户 是否是已经看牌 如果已经看牌 返回牌 但是对其他用户仍旧是隐藏状态
	//深拷贝游戏数据，避免外部直接修改内部状态
	var gameData GameData
	copier.CopyWithOption(&gameData, g.gameData, copier.Option{DeepCopy: true})
	// 默认隐藏所有玩家的手牌（用空切片代替）
	for i := 0; i < g.gameData.ChairCount; i++ {
		if g.gameData.HandCards[i] != nil {
			gameData.HandCards[i] = make([]int, 3)
		} else {
			gameData.HandCards[i] = nil
		}
	}
	if g.gameData.LookCards[user.ChairID] == 1 { // 如果当前用户已经看牌，则展示该用户的真实手牌
		gameData.HandCards[user.ChairID] = g.gameData.HandCards[user.ChairID]
	}
	return gameData
}

func (g *GameFrame) ServerMessagePush(users []string, data any, session *remote.Session) {
	session.Push(users, data, "ServerMessagePush")
}

// StartGame 游戏开始的核心流程
func (g *GameFrame) StartGame(session *remote.Session, user *proto.RoomUser) {
	//1.用户信息变更推送（金币变化） {"gold": 9958, "pushRouter": 'UpdateUserInfoPush'}
	users := g.getAllUsers()
	g.ServerMessagePush(users, UpdateUserInfoPushGold(user.UserInfo.Gold), session)
	//2.庄家推送 {"type":414,"data":{"bankerChairID":0},"pushRouter":"GameMessagePush"}
	if g.gameData.CurBureau == 0 {
		//庄家是每次开始游戏 首次进行操作的座次
		g.gameData.BankerChairID = utils.Rand(len(users))
	}
	g.gameData.CurChairID = g.gameData.BankerChairID
	g.ServerMessagePush(users, GameBankerPushData(g.gameData.BankerChairID), session)
	//3.局数推送{"type":411,"data":{"curBureau":6},"pushRouter":"GameMessagePush"}
	g.gameData.CurBureau++
	g.ServerMessagePush(users, GameBureauPushData(g.gameData.CurBureau), session)
	//4.游戏状态推送 分两步推送 第一步 推送 发牌 牌发完之后 第二步 推送下分 需要用户操作了 推送操作
	//{"type":401,"data":{"gameStatus":1,"tick":0},"pushRouter":"GameMessagePush"}
	g.gameData.GameStatus = SendCards
	g.ServerMessagePush(users, GameStatusPushData(g.gameData.GameStatus, 0), session)
	//5.发牌推送
	g.sendCards(session)
	//6.下分推送
	//先推送下分状态
	g.gameData.GameStatus = PourScore
	g.ServerMessagePush(users, GameStatusPushData(g.gameData.GameStatus, 30), session)
	g.gameData.CurScore = g.gameRule.AddScores[0] * g.gameRule.BaseScore
	for _, v := range g.r.GetUsers() {
		g.ServerMessagePush([]string{v.UserInfo.Uid}, GamePourScorePushData(v.ChairID, g.gameData.CurScore, g.gameData.CurScore, 1, 0), session)
	}
	//7. 轮数推送
	g.gameData.Round = 1
	g.ServerMessagePush(users, GameRoundPushData(g.gameData.Round), session)
	//8. 操作推送
	for _, v := range g.r.GetUsers() {
		//GameTurnPushData ChairID是做操作的座次号（是哪个用户在做操作）
		g.ServerMessagePush([]string{v.UserInfo.Uid}, GameTurnPushData(g.gameData.CurChairID, g.gameData.CurScore), session)
	}
}

// getAllUsers 获取所有玩家的uid
func (g *GameFrame) getAllUsers() []string {
	users := make([]string, 0)
	for _, v := range g.r.GetUsers() {
		users = append(users, v.UserInfo.Uid)
	}
	return users
}

// sendCards 发牌
func (g *GameFrame) sendCards(session *remote.Session) {
	//1.洗牌 然后发牌
	g.logic.washCards()
	for i := 0; i < g.gameData.ChairCount; i++ {
		if g.IsPlayingChairID(i) {
			g.gameData.HandCards[i] = g.logic.getCards()
		}
	}
	//发牌后 推送的时候 如果没有看牌的话 暗牌
	hands := make([][]int, g.gameData.ChairCount)
	for i, v := range g.gameData.HandCards {
		if v != nil {
			hands[i] = []int{0, 0, 0}
		}
	}
	g.ServerMessagePush(g.getAllUsers(), GameSendCardsPushData(hands), session)
}

// IsPlayingChairID 判断玩家是否在游戏中
func (g *GameFrame) IsPlayingChairID(chairID int) bool {
	for _, v := range g.r.GetUsers() {
		if v.ChairID == chairID && v.UserStatus == proto.Playing {
			return true
		}
	}
	return false
}

// onGameLook 看牌
func (g *GameFrame) onGameLook(user *proto.RoomUser, session *remote.Session, cuopai bool) {
	if g.gameData.GameStatus != PourScore || g.gameData.CurChairID != user.ChairID {
		logs.Warn("ID:%s room, sanzhang game look err:gameStatus=%d,curChairID=%d,chairID=%d",
			g.r.GetId(), g.gameData.GameStatus, g.gameData.CurChairID, user.ChairID)
		return
	}
	if !g.IsPlayingChairID(user.ChairID) { //只有处于Playing状态的玩家可以看牌
		logs.Warn("ID:%s room, sanzhang game look err: not playing",
			g.r.GetId())
		return
	}
	//记录当前玩家已看牌
	g.gameData.UserStatusArray[user.ChairID] = Look
	g.gameData.LookCards[user.ChairID] = 1
	for _, v := range g.r.GetUsers() {
		if g.gameData.CurChairID == v.ChairID {
			// 向自己推送真实牌数据
			g.ServerMessagePush([]string{v.UserInfo.Uid}, GameLookPushData(g.gameData.CurChairID, g.gameData.HandCards[v.ChairID], cuopai), session)
		} else { // 向其他玩家推送：某人看牌了，但不给牌数据
			g.ServerMessagePush([]string{v.UserInfo.Uid}, GameLookPushData(g.gameData.CurChairID, nil, cuopai), session)

		}
	}
}

// onGamePourScore 下分
func (g *GameFrame) onGamePourScore(user *proto.RoomUser, session *remote.Session, score int, t int) {
	if g.gameData.GameStatus != PourScore || g.gameData.CurChairID != user.ChairID {
		logs.Warn("ID:%s room, sanzhang onGamePourScore err:gameStatus=%d,curChairID=%d,chairID=%d",
			g.r.GetId(), g.gameData.GameStatus, g.gameData.CurChairID, user.ChairID)
		return
	}
	if !g.IsPlayingChairID(user.ChairID) { //避免非参与玩家进行操作
		logs.Warn("ID:%s room, sanzhang onGamePourScore err: not playing",
			g.r.GetId())
		return
	}
	if score < 0 { //防止负数下注
		logs.Warn("ID:%s room, sanzhang onGamePourScore err: score lt zero",
			g.r.GetId())
		return
	}
	if g.gameData.PourScores[user.ChairID] == nil {
		g.gameData.PourScores[user.ChairID] = make([]int, 0)
	}
	g.gameData.PourScores[user.ChairID] = append(g.gameData.PourScores[user.ChairID], score)
	//所有人的分数
	scores := 0
	for i := 0; i < g.gameData.ChairCount; i++ {
		if g.gameData.PourScores[i] != nil {
			for _, sc := range g.gameData.PourScores[i] {
				scores += sc
			}
		}
	}
	//当前座次的总分
	chairCount := 0
	for _, sc := range g.gameData.PourScores[user.ChairID] {
		chairCount += sc
	}
	g.ServerMessagePush(g.getAllUsers(), GamePourScorePushData(user.ChairID, score, chairCount, scores, t), session)
	//2. 结束下分 座次移动到下一位 推送轮次 推送游戏状态 推送操作的座次
	g.endPourScore(session)
}

func (g *GameFrame) endPourScore(session *remote.Session) {
	//1. 推送轮次 TODO 轮数大于规则的限制 结束游戏 进行结算
	round := g.getCurRound()
	g.ServerMessagePush(g.getAllUsers(), GameRoundPushData(round), session)
	//判断当前的玩家 排除掉已经失败的玩家,如果只剩一个有效玩家,则直接进入结算阶段
	gamerCount := 0
	for i := 0; i < g.gameData.ChairCount; i++ {
		if g.IsPlayingChairID(i) && !utils.Contains(g.gameData.Loser, i) {
			gamerCount++
		}
	}
	if gamerCount == 1 {
		g.startResult(session)
	} else {
		//寻找下一个未失败的玩家
		for i := 0; i < g.gameData.ChairCount; i++ {
			g.gameData.CurChairID++
			g.gameData.CurChairID = g.gameData.CurChairID % g.gameData.ChairCount
			if g.IsPlayingChairID(g.gameData.CurChairID) {
				break
			}
		}
		//推送游戏状态
		g.gameData.GameStatus = PourScore
		g.ServerMessagePush(g.getAllUsers(), GameStatusPushData(g.gameData.GameStatus, 30), session)
		//推送操作玩家
		g.ServerMessagePush(g.getAllUsers(), GameTurnPushData(g.gameData.CurChairID, g.gameData.CurScore), session)

	}
}

// getCurRound 找出从当前操作者往后，第一个处于在玩状态的玩家，并返回该玩家的下注次数
func (g *GameFrame) getCurRound() int {
	cur := g.gameData.CurChairID
	for i := 0; i < g.gameData.ChairCount; i++ {
		cur++
		cur = cur % g.gameData.ChairCount
		if g.IsPlayingChairID(cur) {
			return len(g.gameData.PourScores[cur])
		}
	}
	return 1
}

// onGameCompare 比牌
func (g *GameFrame) onGameCompare(user *proto.RoomUser, session *remote.Session, chairID int) {
	//1. TODO 先下分 跟注结束之后 进行比牌
	//2. 比牌
	fromChairID := user.ChairID
	toChairID := chairID
	result := g.logic.CompareCards(g.gameData.HandCards[fromChairID], g.gameData.HandCards[toChairID])
	//3. 处理比牌结果 推送轮数 状态 显示结果等信息
	if result == 0 { //如果两人牌一样大，主动发起比牌的人判负
		result = -1
	}
	winChairID := -1
	loseChairID := -1
	if result > 0 {
		g.ServerMessagePush(g.getAllUsers(), GameComparePushData(fromChairID, toChairID, fromChairID, toChairID), session)
		winChairID = fromChairID
		loseChairID = toChairID
	} else if result < 0 {
		g.ServerMessagePush(g.getAllUsers(), GameComparePushData(fromChairID, toChairID, toChairID, fromChairID), session)
		winChairID = toChairID
		loseChairID = fromChairID
	}
	if winChairID != -1 && loseChairID != -1 {
		g.gameData.UserStatusArray[winChairID] = Win
		g.gameData.UserStatusArray[loseChairID] = Lose
		g.gameData.Loser = append(g.gameData.Loser, loseChairID)
		g.gameData.Winner = append(g.gameData.Winner, winChairID)
	}
	if winChairID == fromChairID {
		//TODO 赢了之后 继续和其他人进行比牌
	}
	g.endPourScore(session)
}

// startResult 处理结算并推送结果给所有玩家
func (g *GameFrame) startResult(session *remote.Session) {
	g.gameData.GameStatus = Result //设置当前游戏状态为结果阶段
	g.ServerMessagePush(g.getAllUsers(), GameStatusPushData(g.gameData.GameStatus, 0), session)
	if g.gameResult == nil {
		g.gameResult = new(GameResult)
	}
	g.gameResult.Winners = g.gameData.Winner
	g.gameResult.HandCards = g.gameData.HandCards
	g.gameResult.CurScores = g.gameData.CurScores
	g.gameResult.Losers = g.gameData.Loser
	winScores := make([]int, g.gameData.ChairCount) //记录每个玩家的最终输赢分数
	for i := range winScores {
		if g.gameData.PourScores[i] != nil {
			scores := 0
			for _, v := range g.gameData.PourScores[i] {
				scores += v
			}
			winScores[i] = -scores
			for win := range g.gameData.Winner { //将输的分数平分给赢家
				winScores[win] += scores / len(g.gameData.Winner)
			}
		}
	}
	g.gameResult.WinScores = winScores
	g.ServerMessagePush(g.getAllUsers(), GameResultPushData(g.gameResult), session)
	//结算完成 重置游戏 开始下一把
	g.resetGame(session)
	g.gameEnd(session)
}

// resetGame 重置游戏数据状态，为下一局游戏做准备
func (gf *GameFrame) resetGame(session *remote.Session) {
	g := &GameData{
		GameType:   GameType(gf.gameRule.GameFrameType),
		BaseScore:  gf.gameRule.BaseScore,
		ChairCount: gf.gameRule.MaxPlayerCount,
	}
	g.PourScores = make([][]int, g.ChairCount)
	g.HandCards = make([][]int, g.ChairCount)
	g.LookCards = make([]int, g.ChairCount)
	g.CurScores = make([]int, g.ChairCount)
	g.UserStatusArray = make([]UserStatus, g.ChairCount)
	g.UserTrustArray = []bool{false, false, false, false, false, false, false, false, false, false}
	g.Loser = make([]int, 0)
	g.Winner = make([]int, 0)
	g.GameStatus = GameStatus(None)
	gf.gameData = g
	gf.SendGameStatus(g.GameStatus, 0, session)
	gf.r.EndGame(session)
}

// SendGameStatus 推送当前游戏状态
func (g *GameFrame) SendGameStatus(status GameStatus, tick int, session *remote.Session) {
	g.ServerMessagePush(g.getAllUsers(), GameStatusPushData(status, tick), session)
}

func (g *GameFrame) gameEnd(session *remote.Session) {
	//赢家当庄家
	for i := 0; i < g.gameData.ChairCount; i++ {
		if g.gameResult.WinScores[i] > 0 {
			g.gameData.BankerChairID = i
			g.gameData.CurChairID = g.gameData.BankerChairID
		}
	}
	//延迟自动准备下一局
	time.AfterFunc(5*time.Second, func() {
		for _, v := range g.r.GetUsers() {
			g.r.UserReady(v.UserInfo.Uid, session)
		}
	})
}

// onGameAbandon 弃牌
func (g *GameFrame) onGameAbandon(user *proto.RoomUser, session *remote.Session) {
	if !g.IsPlayingChairID(user.ChairID) {
		return
	}
	if utils.Contains(g.gameData.Loser, user.ChairID) {
		return
	}
	g.gameData.Loser = append(g.gameData.Loser, user.ChairID) //将该玩家加入失败者列表
	for i := 0; i < g.gameData.ChairCount; i++ {
		if g.IsPlayingChairID(i) && i != user.ChairID {
			g.gameData.Winner = append(g.gameData.Winner, i)
		}
	}
	g.gameData.UserStatusArray[user.ChairID] = Abandon                                           //设置玩家状态为弃牌
	g.send(GameAbandonPushData(user.ChairID, g.gameData.UserStatusArray[user.ChairID]), session) //推送弃牌消息给所有玩家
	time.AfterFunc(time.Second, func() { //延迟1秒执行结算逻辑
		g.endPourScore(session)
	})
}

// send 把data发送给所有在线玩家
func (g *GameFrame) send(data any, session *remote.Session) {
	g.ServerMessagePush(g.getAllUsers(), data, session)
}
