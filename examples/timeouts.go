// Timeouts：用 select 和 time.After 给等待设置期限。
// 参考：https://gobyexample.com/timeouts
//
// 在项目根目录运行：go run ./examples/timeouts.go
// 预期输出：先输出 timeout 1，再输出 result 2，总耗时约 3 秒（不含编译）。
//
// 学习要点：
//   - select 没有 default 时，会等待某个通信分支就绪，再执行其中一个分支。
//   - time.After 返回只接收 channel，到期后可以从中收到一个时间值。
//   - 超时只表示调用方结束等待，不会自动终止后台 goroutine。
//   - 本例的后台任务只发送一次，容量为 1 可避免接收方离开后发送被阻塞。
package main

import (
	"fmt"
	"time"
)

func main() {
	// 场景一：任务耗时 2 秒，但调用方只愿意等待 1 秒。
	result1 := make(chan string, 1)
	go func() {
		time.Sleep(2 * time.Second) // 模拟耗时操作。
		result1 <- "result 1"
	}()

	select {
	case value := <-result1:
		fmt.Println(value)
	case <-time.After(1 * time.Second):
		fmt.Println("timeout 1")
	}

	// 场景二：任务仍耗时 2 秒，将等待期限放宽到 3 秒。
	// 两个场景按顺序开始；第一个后台任务此时仍可能在执行。
	result2 := make(chan string, 1)
	go func() {
		time.Sleep(2 * time.Second)
		result2 <- "result 2"
	}()

	select {
	case value := <-result2:
		fmt.Println(value) // 约 2 秒后收到结果，无需等满 3 秒。
	case <-time.After(3 * time.Second):
		fmt.Println("timeout 2")
	}
}

// 自测：
// 1. 将第一个超时时间改为 3 秒，观察第一行输出如何变化。
// 2. 去掉 result1 的缓冲，思考超时后发送方会卡在哪里。
// 3. 给 select 添加 default，观察它是否还会等待结果或超时。
// 4. 若多个 case 同时就绪，select 会伪随机选择，不能依赖书写顺序。
