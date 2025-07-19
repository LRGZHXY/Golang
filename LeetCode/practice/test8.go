package practice

func twoSum2(nums []int, target int) []int {
	m := make(map[int]int)
	for i, v := range nums {
		v2 := target - v
		if i2, exist := m[v2]; exist {
			return []int{i, i2}
		}
		m[v] = i
	}
	return []int{}
}
