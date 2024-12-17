package main

import (
	"fmt"
	"strconv"
	"strings"
)

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

var (
	path []string
	res  []string
)

func restoreIpAddresses(s string) []string {
	path, res = make([]string, 0, len(s)), make([]string, 0)
	dfs(s, 0)
	return res
}

func dfs(s string, start int) {
	// 检查是否已经形成了 4 个段（有效的 IP 地址有 4 个段）
	if len(path) == 4 {
		// 如果已经使用了整个字符串，则找到了一个有效的 IP 地址
		if start == len(s) {
			// 将段用点连接起来并添加到结果中
			str := strings.Join(path, ".")
			res = append(res, str)
		}
		return
	}

	// 循环遍历字符串以创建段
	for i := start; i < len(s); i++ {
		// 检查前导零：如果当前段以 '0' 开头且不是唯一数字，则终止
		if i != start && s[start] == '0' {
			break // 因为前导零无效
		}
		// 提取当前段
		str := s[start : i+1]
		// 将段转换为整数
		num, _ := strconv.Atoi(str)
		// 检查该段是否为有效数字（0-255）
		if num >= 0 && num <= 255 {
			path = append(path, str)
			dfs(s, i+1)
			path = path[:len(path)-1] // 回溯 移除最后添加的段
		} else {
			// 如果数字超过 255，则不需要检查后续段
			break
		}
	}
}
