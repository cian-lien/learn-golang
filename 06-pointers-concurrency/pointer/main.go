// 第 6 部分练习 1：指针传参
// 运行：go run ./06-pointers-concurrency/pointer
// 输出：5 100
package main

import "fmt"

func change(a int, b *int) {
	a = 100  // 改的是副本，外面的 x 不受影响
	*b = 100 // 通过地址改到外面的 y
}

func main() {
	x, y := 1, 1
	change(x, &y)
	p := &x
	*p = 5 // 通过地址把 x 改成 5
	fmt.Println(x, y)
}
