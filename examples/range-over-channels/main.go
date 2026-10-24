// Range over Channels：持续接收，直到 channel 已关闭且数据全部收完。
// 参考：https://gobyexample.com/range-over-channels
//
// 在项目根目录运行：go run ./examples/range-over-channels
// 预期输出：one、two，各占一行；完整学习说明见同目录 README.md。
//
// 与 Closing Channels 的联系：range 自动处理接收和 ok 判断。
// channel 暂时为空时会等待；关闭且排空后才自然结束循环。
package main

import "fmt"

func main() {
	// 先发送两次，再开始接收，因此缓冲区需要容纳这两个值。
	queue := make(chan string, 2)
	queue <- "one"
	queue <- "two"

	// 关闭表示以后不再发送，缓冲区中已有的数据仍然可以接收。
	close(queue)

	count := 0
	for elem := range queue {
		fmt.Println(elem)
		count++
	}

	// 自检：关闭后仍应遍历两个已发送的值。
	if count != 2 {
		panic("range 应在接收完两个缓冲值后结束")
	}
}
