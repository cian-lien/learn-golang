// 第 5 部分练习 1：值接收者与指针接收者
// 运行：go run ./05-structs-interfaces/counter
// 输出：2 1
package main

import "fmt"

type Counter struct{ n int }

func (c Counter) AddV()  { c.n++ } // 值接收者：改的是副本
func (c *Counter) AddP() { c.n++ } // 指针接收者：改的是原来的

func main() {
	a := Counter{}
	b := a   // 结构体赋值会复制，a 和 b 互不影响
	a.AddV() // 无效，a.n 还是 0
	a.AddP() // a.n = 1
	a.AddP() // a.n = 2
	b.AddP() // b.n = 1
	fmt.Println(a.n, b.n)
}
