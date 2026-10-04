# 五种并发模式

前四篇讲的是零件：goroutine、channel、select、锁与原子操作。
这篇把它们组装成五个可以直接复用的结构，覆盖了绝大多数工程场景。

**配套代码**：`examples/05_patterns/main.go`

```bash
cd go
go run ./base/08_concurrency/examples/05_patterns
```

---

## 模式一：worker pool（固定并发度）

### 代码

```go
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
```

### 解析

10 个任务、3 个 worker、每个 20ms，实测 80ms，正好是 4 批
（10 个任务分给 3 个 worker：4 + 3 + 3，最慢的那个跑了 4 批）。

这段代码的灵魂是**关闭顺序**：

1. 生产者写完就 `close(jobs)`——发送方负责关闭。
2. worker 用 `for j := range jobs` 消费，读到关闭自然退出。
3. `wg.Wait()` 等所有 worker 都退出。
4. 最后一个动作由独立的 goroutine 完成：`close(results)`。

少了第 3 步的等待，就会出现「某个 worker 还在发送，结果 channel 已经被关闭」，
直接 panic。这个模式值得背下来：

```go
go func() {
	wg.Wait()      // 等所有发送者结束
	close(results) // 再关
}()
```

它解决的根本问题是**并发上限**：worker 数量就是并发上限，
也就是给下游服务装上的背压阀门。任务再多也不会创建更多 goroutine。

## 模式二：pipeline（流水线）

### 代码

```go
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
```

串联起来就是一条流水线，每个阶段各有一个 goroutine：

```go
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
```

### 解析

每个阶段都遵守同一条约定：**上游关闭，我就关闭我的输出**。
写成代码就是 `defer close(out)` 加上 `for v := range in`。
只要每个阶段都守约，整条链就能自然、正确地结束，不需要任何额外的协调。

取消同样沿着这条链传播。下面的例子里消费者只要前 3 个结果，直接 `cancel()`：

```go
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
```

运行时会看到两行「收到取消，提前退出」，最后 `goroutine 数: 1`——
100 个数字没跑完，整条流水线已经退干净了。

注意每个阶段的发送都包在 `select` 里：

```go
select {
case out <- v:
case <-ctx.Done():
	return
}
```

如果写成裸的 `out <- v`，下游一走，这个阶段就会永久阻塞在发送上——
上游的取消信号也就传不过来了。

## 模式三：fan-out / fan-in

### 代码

```go
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
```

### 解析

fan-out 是把一条流拆成多条并行处理，fan-in 是把多条流合并回一条。
`merge` 用「每个输入一个搬运 goroutine + WaitGroup + 统一 close」实现，
和第 2 篇的关闭约定完全一致。

注意输出是**排序后**的：合并结果的自然顺序取决于哪条支路先完成，
每次运行都可能不同。需要稳定顺序时：在末端排序，或者给每个结果带上序号再重排。

两种 fan-out 也很容易混淆，这里区分一下：

| 形式 | 含义 | 例子 |
|---|---|---|
| 同一个输入被多个消费者竞争 | 分摊工作量 | worker pool：N 个 worker 读同一个 jobs channel |
| 同一条流被复制成多条处理链 | 并行处理不同分片 | 本文的例子：3 条 `square(gen(...))` |

## 模式四：限流

### 代码

```go
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
```

### 解析

这两种限流经常被混为一谈，其实解决的是不同问题：

| 机制 | 控制什么 | 典型场景 |
|---|---|---|
| 令牌桶（`<-limiter.C`） | 发起速率：每秒多少个 | 第三方 API 的调用配额 |
| 信号量（`sem <- struct{}{}`） | 在途数量：同时多少个 | 保护数据库连接池 |

从输出可以看到请求 1~5 每隔约 30ms 完成一个：速率被 ticker 卡住了。
如果去掉 ticker 只留信号量，5 个请求会几乎同时打出去（只是并发不超过 2）。

生产代码里更常用 `golang.org/x/time/rate` 做速率限制：
它支持突发流量（burst）和更平滑的算法。这里的 ticker 只是让你看清原理。

## 模式五：singleflight（合并重复请求）

### 代码

```go
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
```

```go
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
```

### 解析

输出是「下游只被调用了 1 次」：10 个 goroutine 同时请求 `user:1`，
只有第一个真正执行了 `fn`，其余 9 个在 `c.wg.Wait()` 上等结果。

实现只有几十行，但几个细节很关键：

1. **持锁时间要短。** `mu` 只用来保护 map，`fn()` 是在**释放锁之后**执行的。
   如果锁住整段逻辑，singleflight 就退化成了串行执行。
2. **完成后必须 `delete`。** 否则这个 key 会永远留在 map 里，
   后续请求永远拿到同一个旧值——这既是内存泄漏，也是数据过期问题。
3. **`wg.Add(1)` 在第一次调用时执行，`wg.Done()` 在 `fn` 返回后执行**，
   搭车者只要 `Wait()` 就能拿到 `val` 和 `err`。

典型场景是**缓存击穿保护**：热点 key 过期瞬间，避免成千上万个请求同时打到数据库。

## 把所有演示串起来

```go
func main() {
	demoWorkerPool()
	demoPipeline()
	demoPipelineCancel()
	demoFanOutFanIn()
	demoRateLimit()
	demoSingleFlight()
}
```

## 运行输出

```text
[1] worker pool：固定并发度，复用 goroutine
    结果: [1 4 9 16 25 36 49 64 81 100]
    10 个任务 × 20ms，3 个 worker，耗时 80ms（约 4 批）
    要点：并发上限由 worker 数量决定，任务再多也不会创建更多 goroutine

[2] pipeline：每个阶段一个 goroutine，用 channel 串联
    1²+2²+3²+4²+5² = 55
    每个阶段都遵守同一条约定：defer close(out)，上游关闭则自己关闭

[3] pipeline + 取消：只需要前 3 个结果
    收到 1
    收到 4
    收到 9
    下游已经满足 → cancel() 通知所有上游停下来
    gen 收到取消，提前退出
    square 收到取消，提前退出
    100 个数字没跑完就收工了，当前 goroutine 数: 1

[4] fan-out / fan-in
    合并后的结果: [1 4 9 16 25 36 49 64 81]
    无论支路完成顺序如何，merge 都能把结果收齐（所以这里可以排序后输出）

[5] 限流：令牌桶控制速率，信号量控制在途数量
    请求1 完成（80ms）
    请求2 完成（110ms）
    请求3 完成（140ms）
    请求4 完成（170ms）
    请求5 完成（200ms）
    两者解决不同问题：令牌桶限「每秒多少个」，信号量限「同时在途多少个」

[6] singleflight：合并重复请求
    10 个并发请求，下游只被调用了 1 次
    常用于：缓存击穿保护、同一资源的重复查询、元数据刷新
```

## 选择哪个模式

| 你的问题 | 用哪个 |
|---|---|
| 有一批任务，要控制并发上限 | worker pool |
| 数据要经过多个处理阶段 | pipeline |
| 同一批数据要并行处理多个分片，再汇总 | fan-out / fan-in |
| 要限制发起速率或在途数量 | 令牌桶 + 信号量 |
| 同一资源被并发重复查询 | singleflight |

## 常见错误

| 写法 | 问题 |
|---|---|
| 生产者和 worker 都去 `close(jobs)` | 重复 close，panic |
| 在 `wg.Wait()` 之前 `close(results)` | 往已关闭的 channel 发送，panic |
| 流水线的发送没包在 `select` 里 | 下游退出后上游卡死 |
| 用 channel 广播而不是 `close` | 只有一个接收者收到信号 |
| singleflight 里持锁执行 `fn` | 退化为串行，还可能有死锁 |
| singleflight 忘记 `delete` | 返回值永久陈旧，map 无限增长 |

## 练习

1. 给 `workerPool` 加上 `context` 取消和错误返回：任一任务失败就停止发放新任务，
   并让排队中的任务快速退出。
2. 写一个「并发处理文件、结果按输入顺序输出」的 pipeline（提示：给结果带上序号）。
3. 用 singleflight + 缓存实现 `GetUser(id)`：缓存未命中时合并并发回源，
   并写测试证明 100 个并发请求只回源一次。

## 小结

- worker pool 解决并发上限，关闭顺序是 `close(jobs) → wg.Wait() → close(results)`。
- pipeline 的阶段约定是「上游关闭，我关闭我的输出」，每个发送都要能响应取消。
- fan-in 就是「多个搬运 goroutine + WaitGroup + 统一 close」。
- 令牌桶管速率，信号量管在途数量。
- singleflight 用 map + WaitGroup 合并重复请求，完成后必须删除 key。

下一篇处理取消与生命周期的工程化：`context`、`errgroup` 与优雅退出。
