package practice

func isValid(s string) bool {
	stack := make([]rune, 0)
	m := map[rune]rune{
		'(': ')',
		'{': '}',
		'[': ']',
	}
	for _, c := range s {
		if c == '(' || c == '{' || c == '[' {
			stack = append(stack, m[c])
		} else {
			if len(stack) == 0 {
				return false
			}
			first := stack[len(stack)-1]
			if first != c {
				return false
			}
			stack = stack[:len(stack)-1]
		}
	}
	return len(stack) == 0
}
