package practice

func twoSum(nums []int, target int) []int {
	for index1, _ := range nums {
		for index2 := index1 + 1; index2 < len(nums); index2++ {
			if nums[index1]+nums[index2] == target {
				return []int{index1, index2}
			}
		}
	}
	return []int{}
}
