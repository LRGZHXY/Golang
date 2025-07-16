package practice

func generateMatrix(n int) [][]int {
	startX, startY := 0, 0
	offset := 1
	count := 1
	loop := n / 2
	mid := n / 2
	result := make([][]int, n)
	for i := 0; i < n; i++ {
		result[i] = make([]int, n)
	}
	for loop > 0 {
		i, j := startX, startY
		for j = startY; j < n-offset; j++ {
			result[startX][j] = count
			count++
		}
		for i = startX; i < n-offset; i++ {
			result[i][j] = count
			count++
		}
		for ; j > startY; j-- {
			result[i][j] = count
			count++
		}
		for ; i > startX; i-- {
			result[i][j] = count
			count++
		}
		loop--
		startX++
		startY++
		offset++
	}
	if n%2 == 1 {
		result[mid][mid] = count
	}
	return result
}
