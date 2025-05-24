package alg

import (
	"fmt"
	"game/component/mj/mp"
	"testing"
)

func TestGen(t *testing.T) {
	table := NewTable()
	table.gen()
}

func TestCheckHu(t *testing.T) {
	h := NewHuLogic()
	cards := []mp.CardID{
		mp.Wan1, mp.Wan1, mp.Wan1, mp.Wan2, mp.Wan3, mp.Wan5, mp.Wan5, mp.Wan5,
		mp.Tong1, mp.Tong1, mp.Tong1, mp.Zhong, mp.Tong4,
	}
	checkHu := h.CheckHu(cards, []mp.CardID{mp.Zhong}, mp.Tong2)
	fmt.Println(checkHu)
}
