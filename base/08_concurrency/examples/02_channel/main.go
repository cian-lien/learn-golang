// 第 2 章：channel 的语义、关闭与所有权
//
// 运行（在 go/ 目录下执行）：
//
//	go run ./base/08_concurrency/examples/02_channel
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	demoUnbuffered()
	demoBuffered()
	demoClose()
	demoClosePanic()
	demoNilChannel()
	demoDirectionAndOwnership()
	demoSemaphore()
}

// demoUnbuffered 演示无缓冲 channel 的「同步握手」语义。
func demoUnbuffered() {
	fmt.Println("[1] 无缓冲 channel：发送和接收必须碰面")
	ch := make(chan string)

	go func() {
		fmt.Println("    sender : 准备发送（此刻会阻塞，直到有人接收）")
		ch <- "ping"
		fmt.Println("    sender : 发送完成（说明接收方已经就位）")
	}()

	time.Sleep(50 * time.Millisecond) // 让 sender 先跑起来，方便观察顺序
	fmt.Println("    main   : 现在开始接收")
	fmt.Println("    main   : 收到", <-ch)
	time.Sleep(20 * time.Millisecond) // 等 sender 打印完最后一行
}

// demoBuffered 演示缓冲区容量对阻塞点的影响。
func demoBuffered() {
	fmt.Println("\n[2] 有缓冲 channel：容量内发送不阻塞")
	ch := make(chan int, 2)
	ch <- 1
	ch <- 2
	fmt.Printf("    len=%d cap=%d：两个元素已经入队，发送方没有被阻塞\n", len(ch), cap(ch))

	// 用 select + default 探测「这一次发送会不会阻塞」，避免程序真的卡住
	select {
	case ch <- 3:
		fmt.Println("    第 3 次发送成功")
	default:
		fmt.Println("    缓冲已满：第 3 次发送会阻塞，被 default 挡下")
	}

	fmt.Println("    收走一个元素:", <-ch)
	select {
	case ch <- 3:
		fmt.Println("    腾出空间后，第 3 次发送成功")
	default:
		fmt.Println("    缓冲仍然是满的")
	}
	fmt.Println("    缓冲区剩余内容:", drain(ch))
}

// drain 读出 channel 中当前所有数据（不等待新的数据）。
func drain(ch <-chan int) []int {
	out := []int{}
	for {
		select {
		case v := <-ch:
			out = append(out, v)
		default:
			return out
		}
	}
}

// demoClose 演示 close 的三种效果：range 结束、零值 + ok、广播唤醒。
func demoClose() {
	fmt.Println("\n[3] close：关闭是广播，不是清空")

	ch := make(chan int, 3)
	for i := 1; i <= 3; i++ {
		ch <- i
	}
	close(ch)

	fmt.Print("    关闭后 range 仍能读出剩余数据:")
	for v := range ch { // 缓冲耗尽且已关闭时，range 自动结束
		fmt.Print(" ", v)
	}
	fmt.Println("  → range 正常结束")

	v, ok := <-ch
	fmt.Printf("    再接收一次: v=%d ok=%v（ok=false 表示 channel 已关闭且没有数据）\n", v, ok)

	// close 是一次广播：所有等待者同时被唤醒
	done := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-done // 三个 goroutine 一起等
		}()
	}
	time.Sleep(20 * time.Millisecond)
	close(done)
	wg.Wait()
	fmt.Println("    close(done) 一次性唤醒 3 个等待者（这就是取消信号的经典用法）")
}

// demoClosePanic 演示两种会 panic 的关闭/发送动作。
func demoClosePanic() {
	fmt.Println("\n[4] 关闭相关的两个 panic")

	ch := make(chan int, 1)
	ch <- 1
	close(ch)

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("    向已关闭的 channel 发送 →", r)
			}
		}()
		ch <- 2 // panic: send on closed channel
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("    重复 close →", r)
			}
		}()
		close(ch) // panic: close of closed channel
	}()

	fmt.Println("    结论：只有发送方能关闭自己拥有的 channel，且必须确保只关一次")
}

// demoNilChannel 演示 nil channel 永远阻塞的特性。
func demoNilChannel() {
	fmt.Println("\n[5] nil channel：永远阻塞")

	var never chan int // 零值是 nil

	select {
	case v := <-never:
		fmt.Println("    永远不会执行", v)
	case <-time.After(30 * time.Millisecond):
		fmt.Println("    nil channel 的分支永不就绪，只能走超时分支")
	}

	fmt.Println("    实战技巧：在 select 里把暂时不关心的 channel 置为 nil，即可关闭该分支")
}

// produce 返回只读 channel。
//
// 约定：channel 的所有者是发送方，也由发送方负责 close。
// 返回 <-chan int 让调用方在编译期就无法往里面写。
func produce(nums []int) <-chan int {
	out := make(chan int, len(nums))
	go func() {
		defer close(out) // 只有发送方知道「没有更多数据了」
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

func demoDirectionAndOwnership() {
	fmt.Println("\n[6] 单向 channel 与所有权约定")

	ch := produce([]int{1, 2, 3, 4})
	sum := 0
	for v := range ch { // 调用方只管读，读到关闭自然结束
		sum += v
	}
	fmt.Println("    生产者 close、消费者 range:", sum)
}

// demoSemaphore 用带缓冲 channel 当信号量，限制同时运行的任务数。
func demoSemaphore() {
	fmt.Println("\n[7] 用带缓冲 channel 做信号量：限制并发数")

	sem := make(chan struct{}, 2) // 最多 2 个任务同时在跑
	var wg sync.WaitGroup
	var mu sync.Mutex
	finished := []int{}

	start := time.Now()
	for i := 1; i <= 4; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}        // 取令牌：满了就在这里等
			defer func() { <-sem }() // 释放令牌
			time.Sleep(50 * time.Millisecond)
			mu.Lock()
			finished = append(finished, i)
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	fmt.Printf("    完成顺序: %v，总耗时 %v（4 个任务 × 50ms，限流 2 后约 100ms）\n",
		finished, time.Since(start).Round(10*time.Millisecond))
}
