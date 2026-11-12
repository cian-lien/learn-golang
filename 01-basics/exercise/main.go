// 第 1 部分练习：预测输出
// 运行：go run ./01-basics/exercise
// 输出：0 2 8 3
package main

import "fmt"

const (
	A = iota // 0
	B        // 1
	C        // 2
)

func main() {
	var x int   // 零值 0
	y := "Go语言" // 2 个英文 × 1 字节 + 2 个汉字 × 3 字节 = 8
	z := 7 / 2  // 整数相除，结果是 3
	fmt.Println(x, C, len(y), z)
}
