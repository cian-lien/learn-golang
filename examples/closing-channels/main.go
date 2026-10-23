// Closing Channels：用 close 通知接收方，后面不会再发送数据。
// 参考：https://gobyexample.com/closing-channels
//
// 在项目根目录运行：go run ./examples/closing-channels
// 发送和接收日志可能交错；最后一行是 received more jobs: false。
//
// 学习要点：
//   - channel 关闭后，缓冲区中的数据仍可接收，more 仍为 true。
//   - 关闭且排空后，接收立即返回零值和 false，需要检查第二个返回值。
//   - close(jobs) 表示任务发完；<-done 等待任务处理完毕的通知。
//   - 与上一篇的联系：已关闭 channel 的接收始终就绪，不会走 default。
//   - 通常由发送方关闭；重复关闭或向已关闭 channel 发送都会 panic。
package main

import "fmt"

func main() {
	jobs := make(chan int, 5)
	done := make(chan bool)

	go func() {
		for {
			job, more := <-jobs
			if !more {
				// 必须在关闭且排空后退出，避免循环不断收到零值。
				fmt.Println("received all jobs")
				done <- true
				return
			}
			fmt.Println("received job", job)
		}
	}()

	for job := 1; job <= 3; job++ {
		jobs <- job
		fmt.Println("sent job", job)
	}

	// main 是唯一发送方，能确定此后不会再有发送操作。
	close(jobs)
	fmt.Println("sent all jobs")

	// 关闭 jobs 不会等待 worker；通过 done 等待处理完毕的通知。
	<-done

	// 自检：关闭且排空后再次接收，应立即得到 int 的零值和 false。
	value, ok := <-jobs
	if value != 0 || ok {
		panic("关闭且排空的 channel 应返回 0 和 false")
	}
	fmt.Println("received more jobs:", ok)
}

// 自测：
// 1. 删除 close(jobs)，worker 和 main 分别会在哪里等待？
// 2. 将发送的某个任务值改为 0，该次接收的 more 会是 false 吗？
// 3. 将最后一次接收放进带 default 的 select，会执行哪个分支？
