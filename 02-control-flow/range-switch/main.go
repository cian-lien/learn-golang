// 第 2 部分练习 1：预测输出
// 运行：go run ./02-control-flow/range-switch
// 输出：
// 0 2 5
// A
package main

import "fmt"

func main() {
	// range 按字符遍历，但索引是字节位置
	for i, ch := range "Hi你好" {
		if ch == 'i' {
			continue
		}
		fmt.Print(i, " ")
	}
	fmt.Println()

	// switch 只执行第一个为 true 的 case，不会穿透
	x := 15
	switch {
	case x > 10:
		fmt.Println("A")
	case x > 5:
		fmt.Println("B")
	}
}
