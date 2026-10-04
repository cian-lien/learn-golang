package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func worker(jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs { // channel 关闭后 range 会自然结束
		results <- j * j
	}
}

func main() {
	// 1) goroutine + WaitGroup + Mutex
	var wg sync.WaitGroup
	var mu sync.Mutex
	total := 0
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			total += i
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Println("WaitGroup + Mutex 求和:", total)

	// 2) 无缓冲 channel：收发必须配对，会阻塞
	ch := make(chan string)
	go func() { ch <- "来自 goroutine" }()
	fmt.Println("无缓冲 channel:", <-ch)

	// 3) 有缓冲 channel 与 close
	buf := make(chan int, 3)
	for i := 1; i <= 3; i++ {
		buf <- i
	}
	close(buf)
	fmt.Print("关闭后 range 读到: ")
	for v := range buf {
		fmt.Print(v, " ")
	}
	fmt.Println()
	v, ok := <-buf
	fmt.Println("关闭后继续接收 → 零值", v, "ok =", ok)

	// 4) 生产者 / 消费者（worker pool）
	jobs := make(chan int, 5)
	out := make(chan int, 5)
	var wg2 sync.WaitGroup
	for w := 0; w < 2; w++ { // 2 个 worker
		wg2.Add(1)
		go worker(jobs, out, &wg2)
	}
	for i := 1; i <= 5; i++ {
		jobs <- i
	}
	close(jobs)
	wg2.Wait()
	close(out)
	sum := 0
	for r := range out {
		sum += r
	}
	fmt.Println("worker pool 平方和:", sum)

	// 5) select：多路等待 + 超时
	never := make(chan string)
	select {
	case msg := <-never:
		fmt.Println("不会到这里:", msg)
	case <-time.After(20 * time.Millisecond):
		fmt.Println("select 超时分支")
	}

	done := make(chan struct{})
	close(done)
	select {
	case <-done:
		fmt.Println("select 选中已关闭的 channel")
	default:
		fmt.Println("没有任何 case 就绪时走 default")
	}

	// 6) context 取消
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	select {
	case <-ctx.Done():
		fmt.Println("context 结束原因:", ctx.Err())
	case <-time.After(time.Second):
		fmt.Println("不该走到这里")
	}

	// 7) sync.Once
	var once sync.Once
	for i := 0; i < 3; i++ {
		once.Do(func() { fmt.Println("sync.Once 只打印一次") })
	}
}
