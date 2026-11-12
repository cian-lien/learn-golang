// 第 6 部分练习 2：并发计算 1 到 5 的平方和
// 运行：go run ./06-pointers-concurrency/squares
// 输出：55
package main

import "fmt"

func main() {
	ch := make(chan int)

	// 注意边界：从 1 开始，并且用 <= 5
	for i := 1; i <= 5; i++ {
		go func(n int) {
			ch <- n * n
		}(i)
	}

	sum := 0
	for i := 0; i < 5; i++ {
		sum += <-ch // 接收 5 次，拿不到会等待
	}
	fmt.Println(sum)
}
