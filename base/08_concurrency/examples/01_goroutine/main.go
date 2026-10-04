// 第 1 章：goroutine 的启动、等待与生命周期
//
// 运行（在 go/ 目录下执行）：
//
//	go run ./base/08_concurrency/examples/01_goroutine
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

const jobDelay = 100 * time.Millisecond

// work 模拟一次耗时的外部调用：网络、磁盘、RPC 都可以抽象成这样。
func work(d time.Duration) { time.Sleep(d) }

// serial 串行执行 n 个任务。
func serial(n int) time.Duration {
	start := time.Now()
	for i := 1; i <= n; i++ {
		work(jobDelay)
	}
	return time.Since(start)
}

// sumOfSquares 作为并发结果的校验基准。
func sumOfSquares(n int) int {
	total := 0
	for i := 1; i <= n; i++ {
		total += i * i
	}
	return total
}

// parallel 并发执行 n 个任务，用 WaitGroup 等它们全部结束。
func parallel(n int) time.Duration {
	start := time.Now()

	var wg sync.WaitGroup
	squares := make([]int, n) // 预先分配，每个 goroutine 只写自己那一格

	for i := 1; i <= n; i++ {
		wg.Add(1) // 必须在 go 之前 +1，否则 Wait 可能提前返回
		go func(i int) {
			defer wg.Done() // 任何返回路径都会正确计数
			work(jobDelay)
			squares[i-1] = i * i // 写不同下标，不构成数据竞争
		}(i)
	}

	wg.Wait() // 阻塞到计数器归零

	total := 0
	for _, v := range squares {
		total += v
	}
	fmt.Printf("    并发求和 = %d，校验通过: %v\n", total, total == sumOfSquares(n))
	return time.Since(start)
}

// demoGoIsNotBlocking 说明 go 语句本身几乎立刻返回，主 goroutine 不会被拖住。
func demoGoIsNotBlocking() {
	start := time.Now()
	result := make(chan int, 1) // 有缓冲：发送方不必等接收方

	go func() {
		work(120 * time.Millisecond)
		result <- 42
	}()

	fmt.Printf("    go 语句之后主 goroutine 继续往下走，只花了 %v\n",
		time.Since(start).Round(time.Millisecond))
	v := <-result
	fmt.Printf("    %v 之后才从 channel 里拿到结果: %d\n",
		time.Since(start).Round(time.Millisecond), v)
}

// demoPerIterationVar 展示 Go 1.22 起 for 循环变量每轮独立。
func demoPerIterationVar() {
	values := make([]int, 4)

	var wg sync.WaitGroup
	for i := range 4 { // Go 1.22+ 支持 range over int
		wg.Add(1)
		go func() {
			defer wg.Done()
			values[i] = i // 捕获的是本轮独有的 i
		}()
	}
	wg.Wait()
	fmt.Println("    闭包直接捕获循环变量:", values)
	fmt.Println("    Go 1.22 之前这里往往是 [4 4 4 4]，当时必须写 i := i 或作为参数传入")
}

// demoRuntimeInfo 打印与调度相关的运行期信息。
func demoRuntimeInfo() {
	fmt.Printf("    逻辑 CPU = %d，GOMAXPROCS = %d，当前 goroutine 数 = %d\n",
		runtime.NumCPU(), runtime.GOMAXPROCS(0), runtime.NumGoroutine())
	fmt.Println("    GOMAXPROCS 限制的是同时执行 Go 代码的线程数（并行度）")
	fmt.Println("    它不限制 goroutine 的数量：几万个 goroutine 可以只用 8 个线程跑（并发度）")
}

func main() {
	fmt.Println("[1] 串行 vs 并发")
	const n = 5
	fmt.Printf("    串行执行 %d 个 %v 的任务: %v\n", n, jobDelay, serial(n).Round(time.Millisecond))
	fmt.Printf("    并发执行同样的任务: %v\n", parallel(n).Round(time.Millisecond))

	fmt.Println("\n[2] go 语句立刻返回")
	demoGoIsNotBlocking()

	fmt.Println("\n[3] 循环变量捕获")
	demoPerIterationVar()

	fmt.Println("\n[4] 运行期信息")
	demoRuntimeInfo()

	fmt.Println("\n[5] 主 goroutine 退出 = 进程退出")
	fmt.Println("    下面这个 goroutine 要 200ms 后才打印，但 main 马上就要返回了：")
	go func() {
		time.Sleep(200 * time.Millisecond)
		fmt.Println("    这一行永远看不到：进程已退出，goroutine 被静默终止")
	}()
	// main 返回 → 进程退出 → 其他 goroutine 直接消失，不会有任何清理或提示。
}
