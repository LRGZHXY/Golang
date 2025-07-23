package practice

func reverseWords(s string) string {
	b := []byte(s)
	//移除多余空格
	slow := 0
	for fast := 0; fast < len(b); fast++ {
		if b[fast] != ' ' { //一个单词的开头
			if slow != 0 {
				b[slow] = ' '
				slow++
			}
			for fast < len(b) && b[fast] != ' ' {
				b[slow] = b[fast]
				slow++
				fast++
			}
		}
	}
	b = b[0:slow]
	//翻转整个字符串
	reverse1(b)
	//翻转单个单词
	first := 0
	for i := 0; i <= len(b); i++ {
		if i == len(b) || b[i] == ' ' {
			reverse1(b[first:i])
			first = i + 1
		}
	}
	return string(b)
}

func reverse1(s []byte) {
	left := 0
	right := len(s) - 1
	for left < right {
		s[left], s[right] = s[right], s[left]
		left++
		right--
	}
}
