package main

import "fmt"

func main() {
	//定义数组
	var arr1 [5]int
	arr1[0] = 1

	//数组初始化
	var arr2 = [3]int{1, 4, 7}
	fmt.Println(arr2)

	var arr3 = [...]int{2: 66, 0: 33, 1: 99, 3: 88}
	fmt.Println(arr3)
}
