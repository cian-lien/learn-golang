// 第 0 篇：零基础的第一个并发程序
//
// 运行（在 go/ 目录下执行）：
//
//	go run ./base/08_concurrency/examples/00_hello_concurrency
package main

import (
	"fmt"
	"sync"
	"time"
)

// taskDelay 表示每件事要花多久，100ms 是"能明显感觉到"的等待时间。
const taskDelay = 100 * time.Millisecond

// doWork 模拟一件需要等待的事：发一次网络请求、读一个文件、查一次数据库，
// 本质都是"发出请求，然后等结果"，代码里用 time.Sleep 来代表这段等待。
func doWork(name string) {
	time.Sleep(taskDelay)
	fmt.Println("    完成:", name)
}

func main() {
	// ── 1. 先看不并发的写法：三件事一件接一件做 ──────────────────────
	fmt.Println("[1] 串行：三件事排队执行")
	start := time.Now()
	doWork("A")
	doWork("B")
	doWork("C")
	fmt.Printf("    总耗时 %v（每件 100ms，三件排队就是 300ms）\n\n",
		time.Since(start).Round(10*time.Millisecond))

	// ── 2. 加一个 go 关键字：三件事同时开始 ────────────────────────
	// go doWork("A") 的意思是：把这次调用交给一个新的 goroutine 去执行，
	// 当前 goroutine 不等它，立刻继续往下走。
	fmt.Println("[2] 并发：用 go 让三件事同时进行")
	start = time.Now()
	go doWork("A")
	go doWork("B")
	go doWork("C")

	// 注意这里用 Sleep 来"等"它们做完，是猜的：加一个任务就要改一次时间。
	// 正确做法见下面第 3 步。
	time.Sleep(150 * time.Millisecond)
	fmt.Printf("    总耗时 %v（三件事同时等待，约等于一件的时间）\n\n",
		time.Since(start).Round(10*time.Millisecond))

	// ── 3. 用 WaitGroup 代替猜时间 ──────────────────────────────────
	// Add(1)：还有一件事没做完。Done()：做完了一件。Wait()：等到一件都不剩。
	fmt.Println("[3] 用 WaitGroup 正确地等待")
	start = time.Now()

	var wg sync.WaitGroup
	for _, name := range []string{"A", "B", "C"} {
		wg.Add(1)
		go func(n string) { // 把 name 作为参数传进去，每个 goroutine 拿到自己的那份
			defer wg.Done()
			doWork(n)
		}(name)
	}
	wg.Wait() // 阻塞在这里，直到计数器回到 0

	fmt.Printf("    总耗时 %v，而且不靠猜：有几个任务就等几个\n\n",
		time.Since(start).Round(10*time.Millisecond))

	// ── 4. 没人等的 goroutine 会随进程一起消失 ──────────────────────
	fmt.Println("[4] 主 goroutine 一退出，整个进程就结束了")
	go doWork("没人等的任务") // 这个任务要 100ms 才完成
	fmt.Println("    main 执行到这里就返回了，上面那个 goroutine 等不到执行完")
}
