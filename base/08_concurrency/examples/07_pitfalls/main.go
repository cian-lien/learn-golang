// 第 7 章：常见陷阱与排查思路
//
// 本章代码本身是「干净」的：go run -race 不应该报任何问题。
// 故意写错的版本放在 broken/ 目录里，用 //go:build ignore 标记，
// 需要单独指定文件运行：
//
//	go run -race ./base/08_concurrency/examples/07_pitfalls/broken/race_counter.go
//	go run ./base/08_concurrency/examples/07_pitfalls/broken/deadlock.go
//	go run ./base/08_concurrency/examples/07_pitfalls/broken/map_write.go
package main

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

func main() {
	demoGoroutineLeak()
	demoLeakFixed()
	demoStackDump()
	demoDeadlockSafe()
	demoWaitGroupTiming()
	demoChannelCloseOrder()
	demoTimerLeak()
	fmt.Println("\n把 -race 加进日常命令：go test -race ./...，出问题前先让它报出来")
}

// leak 启动的 goroutine 没有任何退出路径。
func leak(n int) {
	ch := make(chan int) // 无缓冲，而且没有接收者
	for i := 0; i < n; i++ {
		go func() { ch <- 1 }() // 永久阻塞在这一行，直到进程结束
	}
}

func demoGoroutineLeak() {
	fmt.Println("[1] goroutine 泄漏：没有退出路径的 goroutine")

	base := runtime.NumGoroutine()
	leak(50)
	time.Sleep(50 * time.Millisecond)

	fmt.Printf("    基线 %d → 泄漏后 %d（这 50 个 goroutine 会一直活着）\n", base, runtime.NumGoroutine())
	fmt.Println("    它们各自带着栈内存，这就是「内存缓慢增长」的常见来源")
}

func demoLeakFixed() {
	fmt.Println("\n[2] 修复：给每个 goroutine 一条退路")

	before := runtime.NumGoroutine() // 这里已经包含了第 1 步泄漏的那 50 个
	done := make(chan struct{})
	ch := make(chan int)

	for i := 0; i < 50; i++ {
		go func() {
			select {
			case ch <- 1:
			case <-done: // 没人接收时也能退出
			}
		}()
	}
	close(done)
	time.Sleep(50 * time.Millisecond)

	fmt.Printf("    同样启动 50 个带退路的 goroutine：%d → %d（全部正常退出，没有增加）\n",
		before, runtime.NumGoroutine())
	fmt.Println("    规范：写下 go 之前，先回答「这个 goroutine 什么时候、由谁负责结束」")
}

func demoStackDump() {
	fmt.Println("\n[3] 排查手段：打印所有 goroutine 的栈")

	buf := make([]byte, 1<<20)
	n := runtime.Stack(buf, true) // true = 抓取所有 goroutine
	stacks := string(buf[:n])

	blocked := strings.Count(stacks, "chan send")
	fmt.Printf("    快照 %d 字节，其中 %d 处卡在「chan send」（就是第 1 步留下的那批）\n", n, blocked)
	fmt.Println("    线上更省事的做法：curl 'http://localhost:6060/debug/pprof/goroutine?debug=2'")
}

func demoDeadlockSafe() {
	fmt.Println("\n[4] 死锁：运行时只在「所有 goroutine 都睡死」时报错")
	// 反例（会直接崩溃，且无法 recover）：
	//   ch := make(chan int)
	//   ch <- 1          // 同一个 goroutine 里先发后收，永远等不到接收者
	//   fmt.Println(<-ch)
	// 报错：fatal error: all goroutines are asleep - deadlock!
	// 完整可运行的反例见 broken/deadlock.go（两个 goroutine 交叉加锁）

	ch := make(chan int)
	go func() { ch <- 1 }() // 换一个 goroutine 去发送

	select {
	case v := <-ch:
		fmt.Println("    收到:", v)
	case <-time.After(time.Second):
		fmt.Println("    超时兜底：至少不会永久卡死")
	}
}

func demoWaitGroupTiming() {
	fmt.Println("\n[5] WaitGroup 的 Add 必须在 Wait 之前完成")
	// 反例：
	//   for i := 0; i < 3; i++ {
	//       go func() { wg.Add(1); defer wg.Done(); ... }()  // Add 可能发生在 Wait 之后
	//   }
	//   wg.Wait()   // 可能立刻返回，任务还在后台跑

	var wg sync.WaitGroup
	results := make([]int, 3)

	for i := range 3 {
		wg.Add(1) // 先 Add，再启动
		go func() {
			defer wg.Done()
			time.Sleep(20 * time.Millisecond)
			results[i] = i * i
		}()
	}
	wg.Wait()

	fmt.Println("    正确写法的结果:", results, "—— main 返回前所有任务都已结束")
}

func demoChannelCloseOrder() {
	fmt.Println("\n[6] 关闭顺序：等发送者都退出后再关 channel")

	results := make(chan int, 10)
	var wg sync.WaitGroup

	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results <- i * 10
		}(i)
	}

	// 反例：在这里直接 close(results)，可能 panic（往已关闭的 channel 发送）
	go func() {
		wg.Wait()      // 先等所有发送者结束
		close(results) // 再关闭
	}()

	sum := 0
	for v := range results {
		sum += v
	}
	fmt.Println("    结果总和:", sum)
}

func demoTimerLeak() {
	fmt.Println("\n[7] 定时器：不用了要 Stop")
	// 反例：
	//   for {
	//       select {
	//       case <-ch:
	//       case <-time.After(5 * time.Second):    // 每轮都新建一个定时器
	//       }
	//   }
	//   ticker := time.Tick(time.Second)            // Go 1.23 前不会被回收

	ctx, cancel := context.WithTimeout(context.Background(), 130*time.Millisecond)
	defer cancel()

	ticker := time.NewTicker(40 * time.Millisecond)
	defer ticker.Stop() // 显式释放

	for i := 1; ; i++ {
		select {
		case <-ticker.C:
			fmt.Println("    心跳", i)
		case <-ctx.Done():
			fmt.Println("    退出，定时器已 Stop")
			return
		}
	}
}
