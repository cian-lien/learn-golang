// 第 6 章：context 与 errgroup
//
// 运行（在 go/ 目录下执行）：
//
//	go run ./base/08_concurrency/examples/06_context
//	go run ./base/08_concurrency/examples/06_context/graceful   # HTTP 优雅退出示例
package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

func main() {
	demoCancelTree()
	demoTimeout()
	demoValue()
	demoErrGroup()
	demoGraceful()
}

// watch 一直等到 ctx 被取消。
func watch(name string, ctx context.Context) {
	<-ctx.Done()
	fmt.Printf("    %s 收到取消: %v\n", name, ctx.Err())
}

func demoCancelTree() {
	fmt.Println("[1] context 的取消会级联到所有子节点")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	child1, cancel1 := context.WithCancel(ctx)
	defer cancel1()

	child2, cancel2 := context.WithTimeout(ctx, 5*time.Second)
	defer cancel2()

	go watch("child1", child1)
	go watch("child2", child2)

	time.Sleep(30 * time.Millisecond)
	fmt.Println("    取消父节点 →")
	cancel()
	time.Sleep(30 * time.Millisecond)

	fmt.Println("    child1.Err() =", child1.Err())
	fmt.Println("    child2.Err() =", child2.Err(), "（父节点先取消，超时就没机会触发了）")
	fmt.Println("    结论：defer cancel() 是必须的，它能释放整棵子树的资源")
}

// slowCall 模拟一次支持取消的下游调用。
func slowCall(ctx context.Context) (string, error) {
	done := make(chan string, 1) // 缓冲 1：即使超时返回，发送方也不会永久阻塞
	go func() {
		time.Sleep(200 * time.Millisecond)
		done <- "下游返回的数据"
	}()

	select {
	case v := <-done:
		return v, nil
	case <-ctx.Done():
		// 真实代码里应该把 ctx 传给 http.NewRequestWithContext / db.QueryContext，
		// 让下游自己感知取消，而不是像这里一样只在本地放弃。
		return "", fmt.Errorf("调用被取消: %w", ctx.Err())
	}
}

func demoTimeout() {
	fmt.Println("\n[2] 超时控制")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()

	start := time.Now()
	result, err := slowCall(ctx)

	fmt.Printf("    耗时 %v，结果 %q，错误 %v\n",
		time.Since(start).Round(10*time.Millisecond), result, err)
	fmt.Println("    errors.Is(err, context.DeadlineExceeded) =", errors.Is(err, context.DeadlineExceeded))
	fmt.Println("    区分两种原因：DeadlineExceeded（超时）还是 Canceled（调用方主动取消）")
}

// ctxKey 是自定义的 key 类型，避免不同包用裸 string 互相覆盖。
type ctxKey string

const requestIDKey ctxKey = "request-id"

func handleRequest(ctx context.Context) {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		fmt.Println("    在处理请求:", id)
	}
}

func demoValue() {
	fmt.Println("\n[3] context.Value：只放请求级元数据")

	ctx := context.WithValue(context.Background(), requestIDKey, "req-42")
	handleRequest(ctx)

	fmt.Println("    可以放：request id、trace id、认证信息")
	fmt.Println("    不要放：业务参数、数据库连接、可选配置 —— 这些应该出现在函数签名里")
}

// Group 是 golang.org/x/sync/errgroup 的最小实现，用于教学。
type Group struct {
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	errOnce sync.Once
	err     error
}

// withContext 返回绑定 ctx 的 Group：任一任务出错，ctx 会被取消。
func withContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{cancel: cancel}, ctx
}

func (g *Group) Go(f func() error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if err := f(); err != nil {
			g.errOnce.Do(func() { // 只保留第一个错误
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
	}()
}

func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel() // 释放 context 资源
	}
	return g.err
}

func demoErrGroup() {
	fmt.Println("\n[4] errgroup：一组任务，任一失败就取消其余")

	g, ctx := withContext(context.Background())

	for i := 1; i <= 3; i++ {
		g.Go(func() error {
			select {
			case <-time.After(time.Duration(i*30) * time.Millisecond):
				fmt.Printf("    任务%d 完成\n", i)
				return nil
			case <-ctx.Done():
				fmt.Printf("    任务%d 被取消: %v\n", i, ctx.Err())
				return ctx.Err()
			}
		})
	}

	g.Go(func() error {
		time.Sleep(20 * time.Millisecond)
		fmt.Println("    任务4 失败 → 触发取消")
		return errors.New("数据库连接失败")
	})

	if err := g.Wait(); err != nil {
		fmt.Println("    最终返回第一个错误:", err)
	}
}

func demoGraceful() {
	fmt.Println("\n[5] 优雅退出：先停止接收新任务，再排空在途任务")

	ctx, stop := context.WithCancel(context.Background())
	queue := make(chan int, 8)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case v := <-queue:
				time.Sleep(30 * time.Millisecond)
				fmt.Println("    处理完一个任务:", v)
			case <-ctx.Done():
				fmt.Println("    收到停止信号，改为排空剩余任务")
				for v := range queue { // 队列关闭后自然结束
					time.Sleep(30 * time.Millisecond)
					fmt.Println("    处理完一个在途任务:", v)
				}
				fmt.Println("    排空完成，退出")
				return
			}
		}
	}()

	for i := 1; i <= 3; i++ {
		queue <- i
	}
	time.Sleep(10 * time.Millisecond)

	stop() // 等价于收到 SIGTERM：不再接新活
	time.Sleep(5 * time.Millisecond)
	close(queue) // 上游也停止生产

	wg.Wait()
	fmt.Println("    退出时没有丢掉任何任务")
	fmt.Println("    真实服务里的对应做法：signal.NotifyContext + http.Server.Shutdown")
}
