package mj

import (
	"common/logs"
	"common/utils"
	"encoding/json"
	"framework/remote"
	"game/component/base"
	"game/component/mj/mp"
	"game/component/proto"
	"github.com/jinzhu/copier"
	"sync"
	"time"
)

type GameFrame struct {
	sync.RWMutex
	r             base.RoomFrame
	gameRule      proto.GameRule
	gameData      *GameData
	logic         *Logic
	testCardArray []mp.CardID
	turnSchedule  []*time.Timer
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

// GameMessageHandle 处理客户端发送来的游戏消息
func (g GameFrame) GameMessageHandle(user *proto.RoomUser, session *remote.Session, msg []byte) {
	var req MessageReq
	json.Unmarshal(msg, &req)
	if req.Type == GameChatNotify { //聊天
		g.onGameChat(user, session, req.Data)
	} else if req.Type == GameTurnOperateNotify { //玩家操作
		g.onGameTurnOperate(user, session, req.Data)
	} else if req.Type == GameGetCardNotify { //拿测试的牌
		g.onGetCard(user, session, req.Data)
	}
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
		if i == 1 {
			g.gameData.HandCards[i] = []mp.CardID{ //给第一个玩家发的牌
				mp.Wan1, mp.Wan1, mp.Wan2, mp.Wan2, mp.Wan3, mp.Wan5, mp.Wan5, mp.Wan5,
				mp.Tong1, mp.Tong1, mp.Tong1, mp.Zhong, mp.Tong4,
			}
		}
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
		logs.Warn("已经拿过牌了,chairID:%d", chairID)
		return
	}
	card := g.testCardArray[chairID] //测试牌
	if card > 0 && card < 36 {
		//从牌堆中拿指定的牌
		card = g.logic.getCard(card)
		g.testCardArray[chairID] = 0 //将测试牌数组该位置清零，避免下次再继续使用
	}
	if card <= 0 || card >= 36 {
		cards := g.logic.getCards(1) //抽牌
		if cards == nil || len(cards) == 0 {
			return
		}
		card = cards[0]
	}

	g.gameData.HandCards[chairID] = append(g.gameData.HandCards[chairID], card)
	operateArray := g.getMyOperateArray(session, chairID, card)
	for i := 0; i < g.gameData.ChairCount; i++ {
		uid := g.getUserByChairID(i).UserInfo.Uid
		if i == chairID { // 给当前玩家发明牌
			g.gameTurn([]string{uid}, chairID, card, operateArray, session)
			//g.sendDataUsers([]string{uid}, GameTurnPushData(chairID, card, OperateTime, operateArray), session)
			g.gameData.OperateArrays[i] = operateArray
			g.gameData.OperateRecord = append(g.gameData.OperateRecord, OperateRecord{
				ChairID: i,
				Card:    card,
				Operate: Get,
			})
			g.turnScheduleExecute(chairID, card, operateArray, session)
		} else {
			g.gameTurn([]string{uid}, i, 36, operateArray, session) //其他玩家看到暗牌
			//给其他玩家发暗牌
			//g.sendDataUsers([]string{uid}, GameTurnPushData(chairID, 36, OperateTime, operateArray), session)
		}
	}
	restCardsCount := g.logic.getRestCardsCount() //剩余牌数推送
	g.sendData(GameRestCardsCountPushData(restCardsCount), session)
}

// gameTurn 向指定用户或所有用户发送轮到谁出牌的消息
func (g *GameFrame) gameTurn(uids []string, chairID int, card mp.CardID, operateArray []OperateType, session *remote.Session) {
	g.gameData.Tick = OperateTime //操作倒计时
	if uids == nil {              //发送给所有人
		g.sendData(GameTurnPushData(chairID, card, g.gameData.Tick, operateArray), session)
	} else { //只发送给指定用户
		g.sendDataUsers(uids, GameTurnPushData(chairID, card, g.gameData.Tick, operateArray), session)
	}
}

// turnScheduleExecute 设置定时器，到时间后自动执行操作
func (g *GameFrame) turnScheduleExecute(chairID int, card mp.CardID, operateArray []OperateType, session *remote.Session) {
	if g.turnSchedule[chairID] != nil {
		g.turnSchedule[chairID].Stop() //如果已有一个定时器任务在运行,先停止（防止重复设置）
	}
	g.turnSchedule[chairID] = time.AfterFunc(time.Second, func() {
		if g.gameData.Tick <= 0 {
			if g.turnSchedule[chairID] != nil {
				g.turnSchedule[chairID].Stop() //取消定时
			}
			g.userAutoOperate(chairID, card, operateArray, session) //自动操作
		} else {
			g.gameData.Tick--
			g.turnSchedule[chairID].Reset(time.Second) //重设定时器下一秒再次执行
		}
	})
}

func (g *GameFrame) getMyOperateArray(session *remote.Session, chairID int, card mp.CardID) []OperateType {
	//需要获取用户可操作的行为，比如 弃牌 碰牌 杠牌 胡牌等
	//TODO
	var operateArray = []OperateType{Qi}
	if g.logic.canHu(g.gameData.HandCards[chairID], -1) { //自摸胡
		operateArray = append(operateArray, HuZi)
	}
	cardCount := 0
	for _, v := range g.gameData.HandCards[chairID] {
		if v == card {
			cardCount++ //新摸的牌在手牌中出现的次数
		}
	}
	if cardCount == 4 { //自摸杠
		operateArray = append(operateArray, GangZi)
	}
	//补杠：已经碰了，再摸到一张一样的牌，可以和碰的牌组成杠
	//已经拿牌之后再进行操作
	for _, v := range g.gameData.OperateRecord {
		if v.ChairID == chairID && v.Operate == Peng && v.Card == card {
			operateArray = append(operateArray, GangBu)
		}
	}
	return operateArray
}

// oneGameChat 聊天消息转发
func (g *GameFrame) onGameChat(user *proto.RoomUser, session *remote.Session, data MessageData) {
	g.sendData(GameChatPushData(user.ChairID, data.Type, data.Msg, data.RecipientID), session)
}

func (g *GameFrame) onGameTurnOperate(user *proto.RoomUser, session *remote.Session, data MessageData) {
	if g.turnSchedule[user.ChairID] != nil {
		g.turnSchedule[user.ChairID].Stop() //取消定时
	}

	if data.Operate == Qi { //弃牌
		//向所有人通告 当前用户做了什么操作
		g.sendData(GameTurnOperatePushData(user.ChairID, data.Card, data.Operate, true), session)
		//删除弃掉的牌
		g.gameData.HandCards[user.ChairID] = g.delCards(g.gameData.HandCards[user.ChairID], data.Card, 1)
		//记录本次操作
		g.gameData.OperateRecord = append(g.gameData.OperateRecord, OperateRecord{user.ChairID, data.Card, data.Operate})
		g.gameData.OperateArrays[user.ChairID] = nil //清空该玩家的操作选项
		g.nextTurn(data.Card, session)               //轮到下一个玩家
	} else if data.Operate == Guo {
		g.sendData(GameTurnOperatePushData(user.ChairID, data.Card, data.Operate, true), session)
		g.gameData.OperateRecord = append(g.gameData.OperateRecord, OperateRecord{user.ChairID, data.Card, data.Operate})
		//TODO 如果牌是14 先去弃牌 然后才能做其他的
		//继续操作
		g.setTurn(user.ChairID, session)
	} else if data.Operate == Peng { //碰
		if data.Card == 0 { //客户端未提供要碰的牌
			length := len(g.gameData.OperateRecord)
			if length == 0 {
				//上一个玩家打出的牌操作记录为空
				logs.Error("用户碰操作，但是没有上一个操作记录")
			} else {
				data.Card = g.gameData.OperateRecord[length-1].Card //找到上一个玩家打出的那张牌
			}
		}
		g.sendData(GameTurnOperatePushData(user.ChairID, data.Card, data.Operate, true), session)
		//g.gameData.HandCards[user.ChairID] = append(g.gameData.HandCards[user.ChairID], data.Card) //给当前玩家的牌加上要碰的牌
		//碰相当于损失了2张牌 当用户重新进入房间时 加载gameData handcards中碰的牌放在左下角
		g.gameData.HandCards[user.ChairID] = g.delCards(g.gameData.HandCards[user.ChairID], data.Card, 2)
		//记录本次操作
		g.gameData.OperateRecord = append(g.gameData.OperateRecord, OperateRecord{user.ChairID, data.Card, data.Operate})
		g.gameData.OperateArrays[user.ChairID] = []OperateType{Qi} //玩家碰牌后必须出一张牌
		g.sendData(GameTurnPushData(user.ChairID, 0, OperateTime, g.gameData.OperateArrays[user.ChairID]), session)
		g.gameData.CurChairID = user.ChairID //碰后需要自己出牌
	} else if data.Operate == GangChi { //杠
		if data.Card == 0 { //客户端未提供要杠的牌
			length := len(g.gameData.OperateRecord)
			if length == 0 {
				//上一个玩家打出的牌操作记录为空
				logs.Error("用户吃杠操作，但是没有上一个操作记录")
			} else {
				data.Card = g.gameData.OperateRecord[length-1].Card //找到上一个玩家打出的那张牌
			}
		}
		g.sendData(GameTurnOperatePushData(user.ChairID, data.Card, data.Operate, true), session)
		//g.gameData.HandCards[user.ChairID] = append(g.gameData.HandCards[user.ChairID], data.Card) //给当前玩家的牌加上要杠的牌
		//杠相当于损失了3张牌 当用户重新进入房间时 加载gameData handcards中杠的牌放在左下角
		g.gameData.HandCards[user.ChairID] = g.delCards(g.gameData.HandCards[user.ChairID], data.Card, 3)
		//记录本次操作
		g.gameData.OperateRecord = append(g.gameData.OperateRecord, OperateRecord{user.ChairID, data.Card, data.Operate})
		g.gameData.OperateArrays[user.ChairID] = []OperateType{Qi} //玩家杠牌后必须出一张牌
		g.sendData(GameTurnPushData(user.ChairID, 0, OperateTime, g.gameData.OperateArrays[user.ChairID]), session)
		g.gameData.CurChairID = user.ChairID //杠后需要自己出牌
	} else if data.Operate == HuChi { //胡别人打出的牌（吃胡）
		if data.Card == 0 { //客户端未提供相应的牌
			length := len(g.gameData.OperateRecord)
			if length == 0 {
				//上一个玩家打出的牌操作记录为空
				logs.Error("用户吃胡操作，但是没有上一个操作记录")
			} else {
				data.Card = g.gameData.OperateRecord[length-1].Card //找到上一个玩家打出的那张牌
			}
		}
		g.sendData(GameTurnOperatePushData(user.ChairID, data.Card, data.Operate, true), session)
		g.gameData.HandCards[user.ChairID] = append(g.gameData.HandCards[user.ChairID], data.Card) //把胡的牌加入手牌
		g.gameData.OperateRecord = append(g.gameData.OperateRecord, OperateRecord{user.ChairID, data.Card, data.Operate})
		g.gameData.OperateArrays[user.ChairID] = nil
		g.gameData.CurChairID = user.ChairID //出牌
		g.gameEnd(data.Operate, session)
	} else if data.Operate == HuZi { //自己摸到牌后胡牌
		g.sendData(GameTurnOperatePushData(user.ChairID, data.Card, data.Operate, true), session)
		//g.gameData.HandCards[user.ChairID] = append(g.gameData.HandCards[user.ChairID], data.Card)
		g.gameData.OperateRecord = append(g.gameData.OperateRecord, OperateRecord{user.ChairID, data.Card, data.Operate})
		g.gameData.OperateArrays[user.ChairID] = nil
		g.gameData.CurChairID = user.ChairID //出牌
		g.gameEnd(data.Operate, session)
	} else if data.Operate == GangZi { //自摸杠
		card := g.gameData.HandCards[user.ChairID][len(g.gameData.HandCards[user.ChairID])-1] //取刚摸的牌
		//自摸杠是暗杠 其他玩家看不到杠的牌
		for i := 0; i < g.gameData.ChairCount; i++ {
			if i == user.ChairID { //当前执行杠操作的玩家
				g.sendDataUsers([]string{g.getUserByChairID(i).UserInfo.Uid}, GameTurnOperatePushData(user.ChairID, card, data.Operate, true), session)
			} else {
				g.sendDataUsers([]string{g.getUserByChairID(i).UserInfo.Uid}, GameTurnOperatePushData(user.ChairID, data.Card, data.Operate, true), session)
			}
		}
		g.gameData.HandCards[user.ChairID] = g.delCards(g.gameData.HandCards[user.ChairID], card, 4)
		g.gameData.OperateRecord = append(g.gameData.OperateRecord, OperateRecord{user.ChairID, card, data.Operate})
		//继续操作
		g.setTurn(user.ChairID, session)
	} else if data.Operate == GangBu {
		//1.自摸杠补
		if g.gameData.CurChairID == user.ChairID {
			card := g.gameData.HandCards[user.ChairID][len(g.gameData.HandCards[user.ChairID])-1]
			for i := 0; i < g.gameData.ChairCount; i++ {
				if i == user.ChairID { //当前执行操作的玩家
					g.sendDataUsers([]string{g.getUserByChairID(i).UserInfo.Uid}, GameTurnOperatePushData(user.ChairID, card, data.Operate, true), session)
				} else {
					g.sendDataUsers([]string{g.getUserByChairID(i).UserInfo.Uid}, GameTurnOperatePushData(user.ChairID, data.Card, data.Operate, true), session)
				}
			}
			g.gameData.HandCards[user.ChairID] = g.delCards(g.gameData.HandCards[user.ChairID], card, 1)
			g.gameData.OperateRecord = append(g.gameData.OperateRecord, OperateRecord{user.ChairID, card, data.Operate})
			//继续操作
			g.setTurn(user.ChairID, session)
		} else {
			//2.吃牌杠补 （这是特殊情况，有些麻将的实现中不允许这个操作）
			/*if data.Card == 0 { //客户端未提供要杠的牌
				length := len(g.gameData.OperateRecord)
				if length == 0 {
					//上一个玩家打出的牌操作记录为空
					logs.Error("用户吃杠操作，但是没有上一个操作记录")
				} else {
					data.Card = g.gameData.OperateRecord[length-1].Card //找到上一个玩家打出的那张牌
				}
			}
			g.sendData(GameTurnOperatePushData(user.ChairID, data.Card, data.Operate, true), session)
			//g.gameData.HandCards[user.ChairID] = append(g.gameData.HandCards[user.ChairID], data.Card) //给当前玩家的牌加上要杠的牌
			//杠相当于损失了3张牌 当用户重新进入房间时 加载gameData handcards中杠的牌放在左下角
			//g.gameData.HandCards[user.ChairID] = g.delCards(g.gameData.HandCards[user.ChairID], data.Card, 3)
			//记录本次操作
			g.gameData.OperateRecord = append(g.gameData.OperateRecord, OperateRecord{user.ChairID, data.Card, data.Operate})
			g.setTurn(user.ChairID, session)
			g.gameData.OperateArrays[user.ChairID] = []OperateType{Qi} //玩家杠牌后必须出一张牌
			g.sendData(GameTurnPushData(user.ChairID, 0, OperateTime, g.gameData.OperateArrays[user.ChairID]), session)
			g.gameData.CurChairID = user.ChairID //杠后需要自己出牌*/
		}

	}

}

// delCards 删除牌
func (g *GameFrame) delCards(cards []mp.CardID, card mp.CardID, times int) []mp.CardID { //times删除几张牌
	g.Lock()
	defer g.Unlock()
	newCards := make([]mp.CardID, 0)
	//循环删除多个元素 有越界风险
	count := 0
	for _, v := range cards {
		if v != card {
			newCards = append(newCards, v) //当前牌不是目标牌，直接加入新数组
		} else { //是目标牌
			if count == times {
				newCards = append(newCards, v) //已经删够times次，保留
			} else { //删除
				count++
				continue
			}
		}
	}
	return newCards
}

// nextTurn 轮到下一个玩家
func (g *GameFrame) nextTurn(lastCard mp.CardID, session *remote.Session) {
	//在下一个用户摸牌之前，需要判断其他玩家是否有人可以进行碰 杠 胡 等操作
	hasOtherOperate := false
	if lastCard > 0 && lastCard < 36 {
		for i := 0; i < g.gameData.ChairCount; i++ {
			if i == g.gameData.CurChairID {
				continue //跳过当前出牌的玩家（不能对自己刚打出的牌进行操作）
			}
			operateArray := g.logic.getOperateArray(g.gameData.HandCards[i], lastCard)
			/*	for _, v := range g.gameData.OperateRecord {
				if v.ChairID == i && v.Operate == Peng && v.Card == lastCard { //可以补杠
					operateArray = append(operateArray, GangBu)
				}
			}*/
			if len(operateArray) > 0 { //该玩家可以进行某些操作
				hasOtherOperate = true
				g.gameData.Tick = OperateTime
				g.sendData(GameTurnPushData(i, lastCard, OperateTime, operateArray), session)
				g.gameData.OperateArrays[i] = operateArray
				g.turnScheduleExecute(i, 0, operateArray, session)
			}
		}
	}
	if !hasOtherOperate {
		//直接让下一个用户进行摸排牌
		nextTurnID := (g.gameData.CurChairID + 1) % g.gameData.ChairCount // (当前玩家ID + 1) % 总玩家数
		g.setTurn(nextTurnID, session)
	}
}

// gameEnd 游戏结束
func (g *GameFrame) gameEnd(operate OperateType, session *remote.Session) {
	g.gameData.GameStatus = Result //结算
	g.sendData(GameStatusPushData(g.gameData.GameStatus, 0), session)
	scores := make([]int, g.gameData.ChairCount)
	//结算推送
	/*for i := 0; i < g.gameData.ChairCount; i++ {

	}*/
	l := len(g.gameData.OperateRecord)
	if l <= 0 {
		logs.Error("没有操作记录，不可能游戏结束，请检查")
		return
	}
	lastOperateRecord := g.gameData.OperateRecord[l-1]
	if lastOperateRecord.Operate != HuChi && lastOperateRecord.Operate != HuZi {
		logs.Error("最后一次操作，不是胡牌，不可能游戏结束，请检查")
		return
	}
	result := GameResult{
		Scores:          scores,
		HandCards:       g.gameData.HandCards,
		RestCards:       g.logic.getRestCards(),
		WinChairIDArray: []int{lastOperateRecord.ChairID},
		HuType:          lastOperateRecord.Operate,
		MyMaCards:       []MyMaCard{},
		FangGangArray:   []int{},
	}
	g.gameData.Result = &result
	g.sendData(GameResultPushData(result), session) //推送结算结果

	time.AfterFunc(3*time.Second, func() { //延迟3秒重置游戏
		g.r.EndGame(session)
		g.resetGame(session)
	})
	//倒计时30秒 如果用户未操作 自动准备或者踢出房间
}

// resetGame 重置游戏数据
func (g *GameFrame) resetGame(session *remote.Session) {
	g.gameData.GameStarted = false
	g.gameData.GameStatus = GameStatusNone
	g.sendData(GameStatusPushData(g.gameData.GameStatus, 0), session)
	g.sendData(GameRestCardsCountPushData(g.logic.getRestCardsCount()), session)
	for i := 0; i < g.gameData.ChairCount; i++ { //清除每个玩家的数据
		g.gameData.HandCards[i] = nil
		g.gameData.OperateArrays[i] = nil
	}
	g.gameData.OperateRecord = make([]OperateRecord, 0)
	g.gameData.CurChairID = -1
	g.gameData.Result = nil
}

// onGetCard 记录玩家手动指定的测试牌
func (g *GameFrame) onGetCard(user *proto.RoomUser, session *remote.Session, data MessageData) {
	g.testCardArray[user.ChairID] = data.Card //将客户端传来的牌data.Card存储到g.testCardArray中
}

// userAutoOperate 自动执行操作
func (g *GameFrame) userAutoOperate(chairID int, card mp.CardID, operateArray []OperateType, session *remote.Session) {
	indexOf := IndexOf(operateArray, Qi) //判断操作列表中是否包含弃牌
	user := g.getUserByChairID(chairID)
	if indexOf != -1 {
		/*//操作有弃牌
		//向所有人通告 当前用户做了什么操作
		g.sendData(GameTurnOperatePushData(chairID, card, Qi, true), session)
		//删除弃掉的牌
		g.gameData.HandCards[chairID] = g.delCards(g.gameData.HandCards[chairID], card, 1)
		//记录本次操作
		g.gameData.OperateRecord = append(g.gameData.OperateRecord, OperateRecord{chairID, card, Qi})
		g.gameData.OperateArrays[chairID] = nil //清空该玩家的操作选项
		g.nextTurn(card, session)               //轮到下一个玩家*/
		g.onGameTurnOperate(user, session, MessageData{Operate: Qi, Card: card})
	} else if IndexOf(operateArray, Guo) != -1 {
		g.onGameTurnOperate(user, session, MessageData{Operate: Guo, Card: 0})
		//操作过
		/*g.sendData(GameTurnOperatePushData(chairID, card, Guo, true), session)
		g.gameData.OperateRecord = append(g.gameData.OperateRecord, OperateRecord{chairID, card, Guo})*/
	}
}

func NewGameFrame(rule proto.GameRule, r base.RoomFrame) *GameFrame {
	gameData := initGameData(rule)
	return &GameFrame{
		r:             r,
		gameRule:      rule,
		gameData:      gameData,
		logic:         NewLogic(GameType(rule.GameFrameType), rule.Qidui),
		testCardArray: make([]mp.CardID, gameData.ChairCount),
		turnSchedule:  make([]*time.Timer, gameData.ChairCount),
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
