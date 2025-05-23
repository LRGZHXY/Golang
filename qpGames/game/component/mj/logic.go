package mj

import (
	"common/utils"
	"game/component/mj/alg"
	"game/component/mj/mp"
	"sync"
)

const (
	Wan1  mp.CardID = 1
	Wan2  mp.CardID = 2
	Wan3  mp.CardID = 3
	Wan4  mp.CardID = 4
	Wan5  mp.CardID = 5
	Wan6  mp.CardID = 6
	Wan7  mp.CardID = 7
	Wan8  mp.CardID = 8
	Wan9  mp.CardID = 9
	Tong1 mp.CardID = 11
	Tong2 mp.CardID = 12
	Tong3 mp.CardID = 13
	Tong4 mp.CardID = 14
	Tong5 mp.CardID = 15
	Tong6 mp.CardID = 16
	Tong7 mp.CardID = 17
	Tong8 mp.CardID = 18
	Tong9 mp.CardID = 19
	Tiao1 mp.CardID = 21
	Tiao2 mp.CardID = 22
	Tiao3 mp.CardID = 23
	Tiao4 mp.CardID = 24
	Tiao5 mp.CardID = 25
	Tiao6 mp.CardID = 26
	Tiao7 mp.CardID = 27
	Tiao8 mp.CardID = 28
	Tiao9 mp.CardID = 29
	Dong  mp.CardID = 31
	Nan   mp.CardID = 32
	Xi    mp.CardID = 33
	Bei   mp.CardID = 34
	Zhong mp.CardID = 35
)

type Logic struct {
	sync.RWMutex
	cards    []mp.CardID
	gameType GameType
	qidui    bool
	huLogic  *alg.HuLogic
}

// washCards 洗牌
func (l *Logic) washCards() {
	l.Lock()
	defer l.Unlock()
	l.cards = []mp.CardID{
		Wan1, Wan2, Wan3, Wan4, Wan5, Wan6, Wan7, Wan8, Wan9,
		Wan1, Wan2, Wan3, Wan4, Wan5, Wan6, Wan7, Wan8, Wan9,
		Wan1, Wan2, Wan3, Wan4, Wan5, Wan6, Wan7, Wan8, Wan9,
		Wan1, Wan2, Wan3, Wan4, Wan5, Wan6, Wan7, Wan8, Wan9,
		Tong1, Tong2, Tong3, Tong4, Tong5, Tong6, Tong7, Tong8, Tong9,
		Tong1, Tong2, Tong3, Tong4, Tong5, Tong6, Tong7, Tong8, Tong9,
		Tong1, Tong2, Tong3, Tong4, Tong5, Tong6, Tong7, Tong8, Tong9,
		Tong1, Tong2, Tong3, Tong4, Tong5, Tong6, Tong7, Tong8, Tong9,
		Tiao1, Tiao2, Tiao3, Tiao4, Tiao5, Tiao6, Tiao7, Tiao8, Tiao9,
		Tiao1, Tiao2, Tiao3, Tiao4, Tiao5, Tiao6, Tiao7, Tiao8, Tiao9,
		Tiao1, Tiao2, Tiao3, Tiao4, Tiao5, Tiao6, Tiao7, Tiao8, Tiao9,
		Tiao1, Tiao2, Tiao3, Tiao4, Tiao5, Tiao6, Tiao7, Tiao8, Tiao9,
		Zhong, Zhong, Zhong, Zhong,
	}
	if l.gameType == HongZhong8 {
		l.cards = append(l.cards, Zhong, Zhong, Zhong, Zhong)
	}
	for i := 0; i < 300; i++ { //多轮乱序算法，通过多次交换index和随机random索引的牌，打乱顺序
		index := i % len(l.cards)
		random := utils.Rand(len(l.cards))
		l.cards[index], l.cards[random] = l.cards[random], l.cards[index]
	}
}

// getCards 获取指定数量的牌
func (l *Logic) getCards(num int) []mp.CardID {
	//发牌之后 牌就没了
	l.Lock()
	defer l.Unlock()
	if len(l.cards) < num {
		return nil
	}
	cards := l.cards[:num]
	l.cards = l.cards[num:] //将l.cards从第num张牌开始的部分重新赋值给自身，意味着前num张牌被移除
	return cards
}

// getRestCardsCount 获取剩余牌数
func (l *Logic) getRestCardsCount() int {
	return len(l.cards)
}

func (l *Logic) canHu(cards []mp.CardID, card mp.CardID) bool {
	//胡牌判断 很复杂的一套逻辑
	return l.huLogic.CheckHu(cards, []mp.CardID{Zhong}, card)
}

func NewLogic(gameType GameType, qidui bool) *Logic {
	return &Logic{
		gameType: gameType,
		qidui:    qidui,
		huLogic:  alg.NewHuLogic(),
	}
}
