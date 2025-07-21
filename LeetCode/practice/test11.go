package practice

func canConstruct(ransomNote string, magazine string) bool {
	recode := make([]int, 26)
	for _, v := range magazine {
		recode[v-'a']++
	}
	for _, v := range ransomNote {
		recode[v-'a']--
		if recode[v-'a'] < 0 {
			return false
		}
	}
	return true
}
