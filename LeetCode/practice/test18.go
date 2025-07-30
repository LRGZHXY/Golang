package practice

func isAnagram(s string, t string) bool {
	recode := [26]int{}
	for _, v := range s {
		recode[v-rune('a')]++
	}
	for _, v := range t {
		recode[v-rune('a')]--
	}
	return recode == [26]int{}
}
