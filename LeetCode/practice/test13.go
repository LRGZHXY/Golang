package practice

func reverseStr(s string, k int) string {
	str := []byte(s)
	for i := 0; i < len(s); i += 2 * k {
		if i+k >= len(s) {
			reverse(str[i:len(s)])
		} else {
			reverse(str[i : i+k])
		}
	}
	return string(str)
}

func reverse(s []byte) {
	left := 0
	right := len(s) - 1
	for left < right {
		temp := s[left]
		s[left] = s[right]
		s[right] = temp
		left++
		right--
	}
}
