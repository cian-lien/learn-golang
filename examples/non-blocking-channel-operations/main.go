// Non-Blocking Channel Operations：用 select + default 实现非阻塞通信。
// 参考：https://gobyexample.com/non-blocking-channel-operations
//
// 在项目根目录运行：go run ./examples/non-blocking-channel-operations
// 预期输出：
// no message received
// no message sent
// no activity
//
// 与上一篇 examples/timeouts.go 的联系：
//   - select + time.After：等待结果，但给等待设置期限。
//   - select + default：当前无法通信就执行备用分支，不等待将来的结果。
//
// 学习要点：
//   - 只有所有通信分支当前都无法执行时，才会选择 default。
//   - 多个通信分支同时就绪时，会伪随机选择一个；default 不参与竞争。
//   - 非阻塞只表示不等待通信就绪，分支内的代码仍可能阻塞。
package main

import "fmt"

func main() {
	// 不传容量参数，创建无缓冲 channel；本例也没有启动其他 goroutine。
	messages := make(chan string)
	signals := make(chan bool)

	// 场景一：非阻塞接收。
	// 没有发送者，当前无法接收，因此立即执行 default。
	select {
	case msg := <-messages:
		fmt.Println("received message", msg)
	default:
		fmt.Println("no message received")
	}

	// 场景二：非阻塞发送。
	// messages 没有缓冲，也没有接收者，因此当前无法发送。
	// 上一个 select 已经结束，不会继续等待这次发送。
	// 走 default 表示消息没有发出去，不会暂存，也不会自动重试。
	msg := "hi"
	select {
	case messages <- msg:
		fmt.Println("sent message", msg)
	default:
		fmt.Println("no message sent")
	}

	// 场景三：同时尝试从多个 channel 接收。
	// 两个 channel 都没有发送者，所以仍然立即执行 default。
	// 如果其中一个接收操作已经就绪，就执行该分支，不会选择 default。
	select {
	case msg := <-messages:
		fmt.Println("received message", msg)
	case sig := <-signals:
		fmt.Println("received signal", sig)
	default:
		fmt.Println("no activity")
	}
}

// 自测：
// 1. 将 messages 改为 make(chan string, 1)，预测三次 select 的输出。
// 2. 若缓冲区已满，再次执行非阻塞发送，会进入哪个分支？
// 3. 在 timeouts.go 的 select 中添加 default，还会等待到超时吗？
// 4. 将 select + 空 default 放进没有其他工作的无限循环，会有什么问题？
