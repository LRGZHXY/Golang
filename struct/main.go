package main

import "fmt"

type Teacher struct {
	//变量名字大写外界可以访问这个属性
	Name   string
	Age    int
	School string
}

func main() {
	//创建结构体实例的几种方式
	var t1 Teacher

	var t2 Teacher = Teacher{"小明", 19, "西安邮电大学"}

	var t3 *Teacher = new(Teacher)

	var t4 *Teacher = &Teacher{"小青", 13, "西安财经大学"}

	fmt.Println(t1)
	fmt.Println(t2)
	fmt.Println(t3)
	fmt.Println(t4)

	//test
}
