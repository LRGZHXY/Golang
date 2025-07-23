package practice

func repeatedSubstringPattern(s string) bool {
	next := make([]int, len(s))
	j := 0
	next[0] = 0
	for i := 1; i < len(s); i++ {
		for j > 0 && s[i] != s[j] {
			j = next[j-1]
		}
		if s[i] == s[j] {
			j++
		}
		next[i] = j
	}
	if next[len(s)-1] != 0 && len(s)%(len(s)-next[len(s)-1]) == 0 {
		return true
	}
	return false
}
