// 第 5 章：并发模式（worker pool / pipeline / fan-out / 限流 / singleflight）
//
// 运行（在 go/ 目录下执行）：
//
//	go run ./base/08_concurrency/examples/05_patterns
package main

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	demoWorkerPool()
	demoPipeline()
	demoPipelineCancel()
	demoFanOutFanIn()
	demoRateLimit()
	demoSingleFlight()
}

// workerPool 固定 3 个 worker 处理一批任务。
//
// 关闭顺序是这段代码的关键：
// 生产者 close(jobs) → worker 的 range 结束 → wg.Wait() → close(results)。
// 少了最后一步的等待，就会出现「往已关闭的 channel 发送」的 panic。
func workerPool(tasks []int, size int) []int {
	jobs := make(chan int, len(tasks))
	results := make(chan int, len(tasks))

	var wg sync.WaitGroup
	for w := 1; w <= size; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs { // jobs 关闭后，range 自然结束
				time.Sleep(20 * time.Millisecond) // 模拟耗时处理
				results <- j * j
			}
		}()
	}

	for _, t := range tasks {
		jobs <- t
	}
	close(jobs) // 发送方关闭：worker 读到关闭后退出

	go func() {
		wg.Wait()      // 等所有 worker 结束
		close(results) // 再关结果
	}()

	out := make([]int, 0, len(tasks))
	for r := range results {
		out = append(out, r)
	}
	sort.Ints(out)
	return out
}

func demoWorkerPool() {
	fmt.Println("[1] worker pool：固定并发度，复用 goroutine")

	tasks := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	start := time.Now()
	out := workerPool(tasks, 3)

	fmt.Printf("    结果: %v\n", out)
	fmt.Printf("    10 个任务 × 20ms，3 个 worker，耗时 %v（约 4 批）\n",
		time.Since(start).Round(10*time.Millisecond))
	fmt.Println("    要点：并发上限由 worker 数量决定，任务再多也不会创建更多 goroutine")
}

// gen 是 pipeline 的第一个阶段：把切片变成 channel。
func gen(ctx context.Context, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case out <- n:
			case <-ctx.Done():
				fmt.Println("    gen 收到取消，提前退出")
				return
			}
		}
	}()
	return out
}

// square 是第二阶段：把输入平方后送给下一阶段。
func square(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			select {
			case out <- v * v:
			case <-ctx.Done():
				fmt.Println("    square 收到取消，提前退出")
				return
			}
		}
	}()
	return out
}

func demoPipeline() {
	fmt.Println("\n[2] pipeline：每个阶段一个 goroutine，用 channel 串联")

	ctx := context.Background()
	sum := 0
	for v := range square(ctx, gen(ctx, 1, 2, 3, 4, 5)) {
		sum += v
	}

	fmt.Println("    1²+2²+3²+4²+5² =", sum)
	fmt.Println("    每个阶段都遵守同一条约定：defer close(out)，上游关闭则自己关闭")
}

func demoPipelineCancel() {
	fmt.Println("\n[3] pipeline + 取消：只需要前 3 个结果")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nums := make([]int, 0, 100)
	for i := 1; i <= 100; i++ {
		nums = append(nums, i)
	}

	got := 0
	for v := range square(ctx, gen(ctx, nums...)) {
		fmt.Println("    收到", v)
		got++
		if got == 3 {
			fmt.Println("    下游已经满足 → cancel() 通知所有上游停下来")
			cancel()
			break
		}
	}

	time.Sleep(30 * time.Millisecond) // 等上游打印退出日志
	fmt.Println("    100 个数字没跑完就收工了，当前 goroutine 数:", runtime.NumGoroutine())
}

// merge 把多条支路合并成一条（fan-in）。
func merge(chans ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup

	for _, c := range chans {
		wg.Add(1)
		go func(c <-chan int) {
			defer wg.Done()
			for v := range c {
				out <- v
			}
		}(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func demoFanOutFanIn() {
	fmt.Println("\n[4] fan-out / fan-in")

	ctx := context.Background()

	// fan-out：同一条数据流拆成 3 条并行处理链
	c1 := square(ctx, gen(ctx, 1, 2, 3))
	c2 := square(ctx, gen(ctx, 4, 5, 6))
	c3 := square(ctx, gen(ctx, 7, 8, 9))

	// fan-in：把 3 条支路的结果合并回一条流
	all := []int{}
	for v := range merge(c1, c2, c3) {
		all = append(all, v)
	}
	sort.Ints(all)

	fmt.Println("    合并后的结果:", all)
	fmt.Println("    无论支路完成顺序如何，merge 都能把结果收齐（所以这里可以排序后输出）")
}

func demoRateLimit() {
	fmt.Println("\n[5] 限流：令牌桶控制速率，信号量控制在途数量")

	limiter := time.NewTicker(30 * time.Millisecond) // 每 30ms 放行一个请求
	defer limiter.Stop()
	sem := make(chan struct{}, 2) // 最多 2 个请求同时在途

	var wg sync.WaitGroup
	start := time.Now()

	for i := 1; i <= 5; i++ {
		<-limiter.C // 拿令牌：控制发起速率
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}        // 控制并发：满了就在这排队
			defer func() { <-sem }() // 释放

			time.Sleep(50 * time.Millisecond) // 模拟下游处理
			fmt.Printf("    请求%d 完成（%v）\n", i, time.Since(start).Round(10*time.Millisecond))
		}(i)
	}
	wg.Wait()

	fmt.Println("    两者解决不同问题：令牌桶限「每秒多少个」，信号量限「同时在途多少个」")
}

// call 表示一次正在进行的请求。
type call struct {
	wg  sync.WaitGroup
	val string
	err error
}

// flightGroup 是 singleflight 的最小实现：
// 同一个 key 的并发请求只会真正执行一次，其余请求搭车等结果。
// 生产代码直接用 golang.org/x/sync/singleflight 即可。
type flightGroup struct {
	mu sync.Mutex
	m  map[string]*call
}

func (g *flightGroup) Do(key string, fn func() (string, error)) (string, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait() // 搭车：等第一个请求的结果
		return c.val, c.err
	}

	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn() // 只有第一个请求真正打到下游
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key) // 完成后清除，后续新请求重新发起
	g.mu.Unlock()

	return c.val, c.err
}

func demoSingleFlight() {
	fmt.Println("\n[6] singleflight：合并重复请求")

	var (
		g     flightGroup
		calls atomic.Int64
		wg    sync.WaitGroup
	)

	start := make(chan struct{})
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start // 让 10 个请求同时发出去
			_, _ = g.Do("user:1", func() (string, error) {
				calls.Add(1)
				time.Sleep(50 * time.Millisecond) // 模拟慢查询 / 慢接口
				return "user:1 的数据", nil
			})
		}()
	}
	close(start)
	wg.Wait()

	fmt.Printf("    10 个并发请求，下游只被调用了 %d 次\n", calls.Load())
	fmt.Println("    常用于：缓存击穿保护、同一资源的重复查询、元数据刷新")
}
