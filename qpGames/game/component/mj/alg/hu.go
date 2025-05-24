package alg

import (
	"game/component/mj/mp"
)

var table = NewTable()

type HuLogic struct {
}

func NewHuLogic() *HuLogic {
	return &HuLogic{}
}

// A 万（1-9 *4） B 筒 （1-9 *4） C 条（1-9 *4） D 风 （东南西北 白板 发财等等 1-9 * 4）
// A B C D（红中） 36+36+36+红中数量
// 胡牌的逻辑=N*连子+M*刻子+1*将  连子 = 1,2,3 刻子=2,2,2 将=3,3（对子）
// 13 + 1= 14 在这个牌的基础上去判断  1A 2A 3A 4A 4A 4A 6A 6A 6A 2B 3B 4B 5C 5C
// 算法：1. 编码的操作 1-9A  000000000 每一个位置代表牌有几个 1A 2A 3A 4A 4A 4A 6A 6A 6A （111303000）
// 2. 生成胡牌的信息：111303000编码 代表此牌胡了
// 这样类似的胡的编码非常多，我们把这种很多种可能 叫做穷举法，需要将所有的可能的胡牌的排列 计算出来，转换成编码
// 1A2A5A5A 110020000 如果要胡 0鬼 3A 鬼1 无将 胡3A5A都行 有将 直接胡 我们需要计算这种能胡的所有可能性
// 无鬼 n可能 鬼1 n种可能 鬼2 n种可能 .... 8个 如果有8个鬼 直接胡的 0-7
// 3. 已经把胡牌所有的可能计算出来了，然后将其加载进内存，空间换时间，进行胡牌检测的时候，直接进行匹配即可，查表法
// 1A 2A 3A 4A 4A 4A 6A 6A 6A 2B 3B 4B 5C 5C  111303000  011100000 000020000  = hu
// 先去生成表（所有胡牌的可能） 8张表   feng 8张

// CheckHu cards当前需要判断的玩家的牌 guiList 鬼牌 代替任何一种牌（红中） card 摸到的牌/吃牌
func (h *HuLogic) CheckHu(cards []mp.CardID, guiList []mp.CardID, card mp.CardID) bool {
	if card > 0 && card < 36 && len(cards) < 14 {
		cards = append(cards, card)
	}
	//guiList []{Zhong}
	return h.isHu(cards, guiList)
}

// isHu 判断是否可以胡牌
func (h *HuLogic) isHu(cardList []mp.CardID, guiList []mp.CardID) bool {
	cards := [][]int{
		{0, 0, 0, 0, 0, 0, 0, 0, 0}, // 万
		{0, 0, 0, 0, 0, 0, 0, 0, 0}, // 条
		{0, 0, 0, 0, 0, 0, 0, 0, 0}, // 筒
		{0, 0, 0, 0, 0, 0, 0, 0, 0}, // 风牌
	}
	guiCount := 0
	for _, card := range cardList {
		if IndexOf(guiList, card) != -1 { //是鬼牌
			guiCount++
		} else {
			i := int(card) / 10   // 花色（0万，1条，2筒，3风）
			j := int(card)%10 - 1 // 数值（1-9 转为 0-8）
			cards[i][j]++
		}
	}
	cardData := &CardData{
		guiCount: guiCount,
		jiang:    false,
	}
	for i := 0; i < 4; i++ {
		feng := i == 3
		cardData.cards = cards[i]
		if !h.checkCards(cardData, 0, feng) {
			return false
		}
	}
	if !cardData.jiang && cardData.guiCount%3 == 2 { //用2张鬼牌当将
		return true
	}
	if cardData.jiang && cardData.guiCount%3 == 0 {
		return true
	}
	return false
}

// checkCards 判断是否可以胡牌
func (h *HuLogic) checkCards(data *CardData, guiCount int, feng bool) bool {
	totalCardCount := table.calTotalCardCount(data.cards)
	if totalCardCount == 0 {
		return true
	}
	// 查表 如果表没有 那么就加一个鬼
	if !table.findCards(data.cards, guiCount, feng) {
		if guiCount < data.guiCount {
			//递归 每次鬼+1
			return h.checkCards(data, guiCount+1, feng)
		} else { // 鬼牌用尽，无法胡
			return false
		}
	} else { //找到了组合
		//检查将是否满足条件
		if (totalCardCount+guiCount)%3 == 2 { //胡牌:(牌数 + 鬼牌数) % 3 == 2  <带一个对子>
			if !data.jiang {
				data.jiang = true
			} else if guiCount < data.guiCount {
				//需要再次尝试+1鬼 看是否胡 只能有一个将
				return h.checkCards(data, guiCount+1, feng)
			} else {
				return false
			}
		}
		data.guiCount = data.guiCount - guiCount //减去本次尝试使用的鬼牌数，表示这些鬼牌已经用掉了
	}
	return true
}

type CardData struct {
	cards    []int
	guiCount int
	jiang    bool
}

// IndexOf 查找元素的索引
func IndexOf[T mp.CardID](list []T, v T) int {
	for index, value := range list {
		if value == v {
			return index
		}
	}
	return -1
}
