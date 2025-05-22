package sz

import (
	"common/utils"
	"sort"
	"sync"
)

type Logic struct {
	sync.RWMutex
	cards []int //52张牌
}

func NewLogic() *Logic {
	return &Logic{
		cards: make([]int, 0),
	}
}

// washCards 洗牌
func (l *Logic) washCards() {
	l.cards = []int{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, //方块 0000  1-13
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, //梅花 0001  17-29
		0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, //红桃 0010  33-45
		0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, //黑桃 0011  49-61
	}
	for i, v := range l.cards { //打乱顺序
		random := utils.Rand(len(l.cards))
		l.cards[i] = l.cards[random]
		l.cards[random] = v
	}
}

// getCards 获取三张手牌
func (l *Logic) getCards() []int {
	cards := make([]int, 3)
	l.RLock()
	defer l.RUnlock()
	for i := 0; i < 3; i++ { //抽3张牌
		if len(cards) == 0 {
			break
		}
		card := l.cards[len(l.cards)-1]
		l.cards = l.cards[:len(l.cards)-1]
		cards[i] = card
	}
	return cards
}

// CompareCards 比牌 result 0 he 大于0 win 小于0 lose
func (l *Logic) CompareCards(from []int, to []int) int {
	//获取牌类型 散牌、对子、顺子等
	fromType := l.getCardsType(from)
	toType := l.getCardsType(to)
	if fromType != toType { // 如果两手牌牌型不同，直接返回它们牌型的差值（用牌型大小来判定输赢）
		return int(fromType - toType)
	}
	//牌型相同且是对子
	if fromType == DuiZi {
		duiFrom, danFrom := l.getDuiZi(from)
		duiTo, danTo := l.getDuiZi(to)
		if duiFrom != duiTo { // 先比较对子大小
			return duiFrom - duiTo
		}
		return danFrom - danTo // 如果对子相等，比较单牌大小
	}
	// 其他牌型，获取每手牌的点数排序数组（一般是从小到大）
	valuesFrom := l.getCardValues(from)
	valuesTo := l.getCardValues(to)
	// 依次比较三张牌的点数，先比较最大的那张牌（索引2）
	if valuesFrom[2] != valuesTo[2] {
		return valuesFrom[2] - valuesTo[2]
	}
	// 如果最大牌相同，比较第二大的牌（索引1）
	if valuesFrom[1] != valuesTo[1] {
		return valuesFrom[1] - valuesTo[1]
	}
	// 如果前两张牌都相同，比较第三张牌（索引0）
	if valuesFrom[0] != valuesTo[0] {
		return valuesFrom[0] - valuesTo[0]
	}
	return 0 // 三张牌点数都相同，判定为平局
}

// getCardsType 获取牌型
func (l *Logic) getCardsType(cards []int) CardsType {
	//先分别获取三张牌的点数（数字部分）
	one := l.getCardsNumber(cards[0])
	two := l.getCardsNumber(cards[1])
	three := l.getCardsNumber(cards[2])
	// 判断是否为豹子（三张点数相等）
	if one == two && two == three {
		return BaoZi
	}
	// 判断是否为金花（同花色）
	jinhua := false
	oneColor := l.getCardsColor(cards[0])
	twoColor := l.getCardsColor(cards[1])
	threeColor := l.getCardsColor(cards[2])
	if oneColor == twoColor && twoColor == threeColor {
		jinhua = true
	}
	// 判断是否为顺子（连续的牌点）
	shunzi := false
	values := l.getCardValues(cards)
	oneV := values[0]
	twoV := values[1]
	threeV := values[2]
	// 判断是否连续，比如 3,4,5 或者 A,2,3 这种特殊顺子
	if oneV+1 == twoV && twoV+1 == threeV {
		shunzi = true
	}
	if oneV == 2 && twoV == 3 && threeV == 14 {
		shunzi = true
	}
	if jinhua && shunzi { // 顺金：既是顺子又是金花
		return ShunJin
	}
	if jinhua { // 金花：同花但不是顺子
		return JinHua
	}
	if shunzi { // 顺子：不是同花的连续牌
		return ShunZi
	}
	if oneV == twoV || twoV == threeV { // 对子：有两张牌点数相同
		return DuiZi
	}
	return DanZhang

}

// getCardValues 获取牌的点数数组,并从小到大排序
func (l *Logic) getCardValues(cards []int) []int {
	v := make([]int, len(cards))
	for i, card := range cards {
		v[i] = l.getCardsValue(card)
	}
	sort.Ints(v)
	return v
}

// getCardsValue 提取牌的点数，并特殊处理A牌，将点数1变为14（方便顺子判断中将A作为最大牌）
func (l *Logic) getCardsValue(card int) int {
	value := card & 0x0f //1-13 2-14
	if value == 1 {
		value += 13
	}
	return value
}

// getCardsNumber 获取牌的点数
func (l *Logic) getCardsNumber(card int) int {
	return card & 0x0f // 牌值&0x0F:保留低四位(点数部分)
}

// getCardsColor 获取牌的花色
func (l *Logic) getCardsColor(card int) string {
	colors := []string{"方块", "梅花", "红桃", "黑桃"} //0→方块，1→梅花，2→红桃，3→黑桃
	//取模  1-13 /16 0 17-29/16 = 1
	if card >= 0x01 && card <= 0x3d {
		return colors[card/0x10]
	}
	return ""
}

// getDuiZi 针对三张牌中有一对的情况，返回对子点数和单牌点数
func (l *Logic) getDuiZi(cards []int) (int, int) {
	// AAB BAA
	values := l.getCardValues(cards)
	if values[0] == values[1] {
		//AAB
		return values[0], values[2]
	}
	return values[1], values[0] //BAA
}
