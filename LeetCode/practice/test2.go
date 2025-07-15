package practice

func sortedSquares(nums []int) []int {
	n := len(nums)
	i, j, k := 0, n-1, n-1
	result := make([]int, n)
	for i <= j {
		left := nums[i] * nums[i]
		right := nums[j] * nums[j]
		if left < right {
			result[k] = right
			k--
			j--
		}
		if left >= right {
			result[k] = left
			k--
			i++
		}
	}
	return result
}
