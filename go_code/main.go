package main

import (
	"fmt"
	"unsafe"
)

func main() {
	fmt.Println("hello world")

	//定义变量的三种方式
	var i int
	var str string
	i = 10
	fmt.Println("i=", i)
	str = fmt.Sprintf("%d", i)
	fmt.Printf("str type %T str=%q\n", str, str)

	var num = 10.4
	fmt.Println("num=", num)
	//fmt.Printf()可以用于做格式化输出
	fmt.Printf("num的类型 %T\n", num)
	fmt.Printf("num占用的字节数是 %d\n", unsafe.Sizeof(num))

	//:=左侧的变量不应该是已经声明过的
	name := "tom"
	fmt.Println("name=", name)

	//for循环
	for i := 1; i < 10; i++ {
		fmt.Println("hello", i)
	}

	j := 2
	for j < 9 {
		fmt.Println("hi~", j)
		j++
	}

	var str2 string = "nihao,厦门！"
	for index, val := range str2 {
		fmt.Printf("index=%d,val=%c \n", index, val)
	}

	var id string
	fmt.Scanln(&id)
	fmt.Println("id=", id)

	var number int
	fmt.Scanf("%d", &number)
	fmt.Println("number=", number)
}
