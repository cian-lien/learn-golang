// 第 3 章：select、超时与取消
//
// 运行（在 go/ 目录下执行）：
//
//	go run ./base/08_concurrency/examples/03_select_timeout
package main

import (
	"fmt"
	"time"
)

func main() {
	demoSelectBasics()
	demoSelectRandom()
	demoTimeout()
	demoReusableTimer()
	demoNonBlocking()
	demoOrDone()
	fmt.Println("\n[7] select{} 不带任何 case 会永久阻塞（适合让 main 不退出，不要在需要退出的地方写）")
}

// demoSelectBasics 演示 select 最基本的行为：谁先就绪就执行谁。
func demoSelectBasics() {
	fmt.Println("[1] select：同时等待多个 channel")
	fast := make(chan string)
	slow := make(chan string)

	go func() { time.Sleep(20 * time.Millisecond); fast <- "fast 的结果" }()
	go func() { time.Sleep(150 * time.Millisecond); slow <- "slow 的结果" }()

	for i := 0; i < 2; i++ {
		select {
		case v := <-fast:
			fmt.Println("    收到:", v)
		case v := <-slow:
			fmt.Println("    收到:", v)
		}
	}
	fmt.Println("    第二次 select 会一直阻塞，直到 slow 也把结果送过来")
}

// demoSelectRandom 演示多个 case 同时就绪时 select 的随机性。
func demoSelectRandom() {
	fmt.Println("\n[2] 多个 case 同时就绪：随机选一个")

	a := make(chan int, 1)
	b := make(chan int, 1)
	counts := map[string]int{}

	for i := 0; i < 200; i++ {
		a <- 1
		b <- 1
		select {
		case <-a:
			counts["a"]++
		case <-b:
			counts["b"]++
		}
		// 清空残留，保证下一轮两个 case 都是就绪状态
		for len(a) > 0 {
			<-a
		}
		for len(b) > 0 {
			<-b
		}
	}

	fmt.Printf("    200 次里 a=%d、b=%d\n", counts["a"], counts["b"])
	fmt.Println("    随机是为了避免某个分支长期被饿死，所以不要依赖 select 的顺序")
}

// demoTimeout 演示超时退出，以及超时之后上游 goroutine 的处境。
func demoTimeout() {
	fmt.Println("\n[3] 超时：给「等不到结果」一个出口")

	// 关键：缓冲为 1。即使调用方超时走了，发送方也不会永久阻塞。
	slow := make(chan string, 1)
	go func() {
		time.Sleep(200 * time.Millisecond)
		slow <- "迟到的结果"
	}()

	select {
	case v := <-slow:
		fmt.Println("    正常收到:", v)
	case <-time.After(50 * time.Millisecond):
		fmt.Println("    50ms 内没有结果 → 超时返回")
	}

	time.Sleep(200 * time.Millisecond)
	fmt.Println("    超时只是「调用方不等了」，goroutine 还在跑；")
	fmt.Println("    如果它往无缓冲 channel 发送，那次发送会永久阻塞 —— 这就是典型的 goroutine 泄漏")
}

// waitIdle 反复读取事件，超过 idle 没有新事件就退出。
// 全程只用一个定时器：循环里反复 time.After 会不断产生新的定时器。
func waitIdle(ch <-chan int, idle time.Duration) {
	timer := time.NewTimer(idle)
	defer timer.Stop() // 一定要 Stop，避免未触发的定时器一直占着资源

	for {
		select {
		case v, ok := <-ch:
			if !ok {
				fmt.Println("    上游关闭，退出")
				return
			}
			fmt.Println("    收到事件:", v)
			if !timer.Stop() { // 重置前先停掉并排空，避免读到过期事件
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(idle)
		case <-timer.C:
			fmt.Printf("    %v 没有新事件，空闲退出\n", idle)
			return
		}
	}
}

func demoReusableTimer() {
	fmt.Println("\n[4] 定时器复用")

	events := make(chan int, 8)
	go func() {
		for i := 1; i <= 3; i++ {
			time.Sleep(40 * time.Millisecond)
			events <- i
		}
		// 故意不 close：让 waitIdle 走「空闲超时」分支
	}()

	waitIdle(events, 100*time.Millisecond)

	fmt.Println("    反例：for { select { case <-ch: case <-time.After(5 * time.Second): } }")
	fmt.Println("    每轮都新建定时器，5 秒内循环 1000 次就会留下最多 1000 个尚未触发的定时器")
}

// demoNonBlocking 演示 default 分支带来的非阻塞语义。
func demoNonBlocking() {
	fmt.Println("\n[5] default：非阻塞操作")

	ch := make(chan int, 1)

	select {
	case v := <-ch:
		fmt.Println("    读到", v)
	default:
		fmt.Println("    通道里没有数据 → 不等待，立刻执行 default")
	}

	ch <- 7
	select {
	case v := <-ch:
		fmt.Println("    读到", v)
	default:
		fmt.Println("    通道里没有数据")
	}

	fmt.Println("    警告：在 for 循环里用 default 会退化成 100% CPU 的忙轮询")
}

// orDone 把「可取消」包装进一个只读 channel。
// 无论上游被取消，还是消费者提前退出，它都不会留下阻塞的 goroutine。
func orDone(done <-chan struct{}, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- v:
				case <-done: // 下游已经走了，不要在这里死等
					return
				}
			}
		}
	}()
	return out
}

// countUp 每 interval 产生一个递增数字，直到 done 被关闭。
// 注意两次 select：等待节拍时会响应取消，发送时也会响应取消。
func countUp(done <-chan struct{}, interval time.Duration) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		i := 0
		for {
			select {
			case <-done:
				fmt.Println("    上游 goroutine 收到取消，已退出")
				return
			case <-ticker.C:
				i++
				select {
				case out <- i:
				case <-done:
					fmt.Println("    上游 goroutine 在发送时收到取消，已退出")
					return
				}
			}
		}
	}()
	return out
}

// demoOrDone 演示 done channel 的取消传播：消费者收工，上游立刻全部退出。
func demoOrDone() {
	fmt.Println("\n[6] 取消模式：done channel + orDone")

	done := make(chan struct{})
	numbers := orDone(done, countUp(done, 30*time.Millisecond))

	count := 0
	for v := range numbers {
		fmt.Println("    消费:", v)
		count++
		if count == 3 {
			fmt.Println("    消费者决定收工 → close(done)")
			close(done)
		}
	}

	time.Sleep(20 * time.Millisecond) // 等上游打印退出日志，仅为了让输出整齐
	fmt.Println("    range 正常结束：done 关闭后，上游所有 goroutine 都已退出")
}
