package mj

import (
	"common/logs"
	"common/utils"
	"framework/remote"
	"game/component/base"
	"game/component/mj/mp"
	"game/component/proto"
	"github.com/jinzhu/copier"
	"time"
)

type GameFrame struct {
	r        base.RoomFrame
	gameRule proto.GameRule
	gameData *GameData
	logic    *Logic
}

func (g *GameFrame) GetGameData(session *remote.Session) any {
	//获取场景 获取游戏的数据
	//mj打牌的时候 不能看到别人的牌
	chairID := g.r.GetUsers()[session.GetUid()].ChairID
	var gameData GameData
	copier.CopyWithOption(&gameData, g.gameData, copier.Option{IgnoreEmpty: true, DeepCopy: true}) //拷贝原始游戏数据
	handCards := make([][]mp.CardID, g.gameData.ChairCount)
	for i := range gameData.HandCards {
		if i == chairID { //自己的牌
			handCards[i] = gameData.HandCards[i]
		} else {
			//每张牌置为36（表示看不见）
			handCards[i] = make([]mp.CardID, len(gameData.HandCards[i]))
			for j := range g.gameData.HandCards[i] {
				handCards[i][j] = 36
			}
		}
	}
	gameData.HandCards = handCards
	if g.gameData.GameStatus == GameStatusNone {
		gameData.RestCardsCount = 9*3*4 + 4
		if g.gameRule.GameFrameType == HongZhong8 {
			gameData.RestCardsCount = 9*3*4 + 8
		}
	}
	return gameData
}

// StartGame 开始游戏
func (g GameFrame) StartGame(session *remote.Session, user *proto.RoomUser) {
	//1.游戏状态 初始状态 推送
	g.gameData.GameStarted = true
	g.gameData.GameStatus = Dices //掷骰子
	g.sendData(GameStatusPushData(g.gameData.GameStatus, GameStatusTmDices), session)
	//2.庄家推送
	if g.gameData.CurBureau == 0 { //第一局默认0号玩家为庄家
		g.gameData.BankerChairID = 0
	} else {
		//TODO win是庄家
	}
	g.sendData(GameBankerPushData(g.gameData.BankerChairID), session)
	//3.摇骰子推送
	dice1 := utils.Rand(6) + 1
	dice2 := utils.Rand(6) + 1
	g.sendData(GameDicesPushData(dice1, dice2), session)
	//4.发牌推送
	g.sendHandCards(session)

	//10.当前局数推送
	g.gameData.CurBureau++ //局数+1
	g.sendData(GameBureauPushData(g.gameData.CurBureau), session)
}

func (g GameFrame) GameMessageHandle(user *proto.RoomUser, session *remote.Session, msg []byte) {
	//TODO implement me
	panic("implement me")
}

// sendDataUsers 向指定的users推送消息
func (g *GameFrame) sendDataUsers(users []string, data any, session *remote.Session) {
	g.ServerMessagePush(users, data, session)
}

// sendData 把data发送给所有在线玩家
func (g *GameFrame) sendData(data any, session *remote.Session) {
	g.ServerMessagePush(g.getAllUsers(), data, session)
}
func (g *GameFrame) ServerMessagePush(users []string, data any, session *remote.Session) {
	session.Push(users, data, "ServerMessagePush")
}

// getAllUsers 获取所有在线用户的uid
func (g *GameFrame) getAllUsers() []string {
	users := make([]string, 0)
	for _, v := range g.r.GetUsers() {
		users = append(users, v.UserInfo.Uid)
	}
	return users
}

// sendHandCards 发牌
func (g *GameFrame) sendHandCards(session *remote.Session) {
	g.logic.washCards() //洗牌
	for i := 0; i < g.gameData.ChairCount; i++ {
		g.gameData.HandCards[i] = g.logic.getCards(13) //给每位玩家发13张牌
	}
	for i := 0; i < g.gameData.ChairCount; i++ {
		handCards := make([][]mp.CardID, g.gameData.ChairCount)
		for j := 0; j < g.gameData.ChairCount; j++ {
			if i == j { //为每个玩家构造一个只包含自己手牌，其他人手牌全部隐藏
				handCards[i] = g.gameData.HandCards[i]
			} else {
				handCards[j] = make([]mp.CardID, len(g.gameData.HandCards[j]))
				for k := range g.gameData.HandCards[j] {
					handCards[j][k] = 36
				}
			}
		}
		uid := g.getUserByChairID(i).UserInfo.Uid
		g.sendDataUsers([]string{uid}, GameSendCardsPushData(handCards, i), session) //推送发牌数据
	}
	restCardsCount := g.logic.getRestCardsCount() //获取剩余牌堆的牌数并推送给所有玩家
	g.sendData(GameRestCardsCountPushData(restCardsCount), session)
	time.AfterFunc(time.Second, func() {
		g.gameData.GameStatus = Playing //游戏状态
		g.sendData(GameStatusPushData(g.gameData.GameStatus, GameStatusTmPlay), session)
		//玩家的操作
		g.setTurn(g.gameData.BankerChairID, session)
	})

}

// getUserByChairID 根据座位号获取对应玩家信息
func (g *GameFrame) getUserByChairID(chairID int) *proto.RoomUser {
	for _, v := range g.r.GetUsers() {
		if v.ChairID == chairID {
			return v
		}
	}
	return nil
}

// setTurn 抽牌
func (g *GameFrame) setTurn(chairID int, session *remote.Session) {
	g.gameData.CurChairID = chairID
	if len(g.gameData.HandCards[chairID]) >= 14 { //牌不能大于14
		logs.Warn("已经拿过牌了")
		return
	}
	card := g.logic.getCards(1)[0] //抽牌
	g.gameData.HandCards[chairID] = append(g.gameData.HandCards[chairID], card)
	operateArray := g.getMyOperateArray(session, chairID, card)
	for i := 0; i < g.gameData.ChairCount; i++ {
		uid := g.getUserByChairID(i).UserInfo.Uid
		if i == chairID { // 给当前玩家发明牌
			g.sendDataUsers([]string{uid}, GameTurnPushData(i, card, OperateTime, operateArray), session)
			g.gameData.OperateArrays[i] = operateArray
			g.gameData.OperateRecord = append(g.gameData.OperateRecord, OperateRecord{
				ChairID: i,
				Card:    card,
				Operate: Get,
			})
		} else {
			//给其他玩家发暗牌
			g.sendDataUsers([]string{uid}, GameTurnPushData(i, 36, OperateTime, operateArray), session)
		}
	}
	restCardsCount := g.logic.getRestCardsCount() //剩余牌数推送
	g.sendData(GameRestCardsCountPushData(restCardsCount), session)
}

func (g *GameFrame) getMyOperateArray(session *remote.Session, chairID int, card mp.CardID) []OperateType {
	//需要获取用户可操作的行为，比如 弃牌 碰牌 杠牌 胡牌等
	//TODO
	var operateArray = []OperateType{Qi}
	if g.logic.canHu(g.gameData.HandCards[chairID], -1) {
		operateArray = append(operateArray, HuZhi) //自己拿牌
	}

	return operateArray
}

func NewGameFrame(rule proto.GameRule, r base.RoomFrame) *GameFrame {
	gameData := initGameData(rule)
	return &GameFrame{
		r:        r,
		gameRule: rule,
		gameData: gameData,
		logic:    NewLogic(GameType(rule.GameFrameType), rule.Qidui),
	}
}

func initGameData(rule proto.GameRule) *GameData {
	g := new(GameData)
	g.ChairCount = rule.MaxPlayerCount //座位数
	g.HandCards = make([][]mp.CardID, g.ChairCount)
	g.GameStatus = GameStatusNone
	g.OperateRecord = make([]OperateRecord, 0)
	g.OperateArrays = make([][]OperateType, g.ChairCount)
	g.CurChairID = -1
	g.RestCardsCount = 9*3*4 + 4
	if rule.GameFrameType == HongZhong8 {
		g.RestCardsCount = 9*3*4 + 8
	}
	return g
}
