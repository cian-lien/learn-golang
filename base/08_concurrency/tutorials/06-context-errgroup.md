# context、errgroup 与优雅退出

前三篇里的取消信号都是自己造的 `done` channel。真实项目里，这件事由 `context` 承担：
它把「请求的截止时间、取消信号、请求级元数据」打包在一起，沿着调用链一路传下去。

这篇讲四件事：取消如何传播、超时如何判断、errgroup 怎么写并发任务、服务如何优雅退出。

**配套代码**：`examples/06_context/main.go`、`examples/06_context/graceful/main.go`

```bash
cd go
go run ./base/08_concurrency/examples/06_context
go run ./base/08_concurrency/examples/06_context/graceful
```

---

## 一、取消是一棵树的传播

### 代码

```go
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
```

### 解析

`context.WithCancel` 返回 `(ctx, cancel)`。取消父节点，所有子节点都会被取消，
形成一棵传播树。示例里的 `child2` 明明设了 5 秒超时，但父节点 30ms 就被取消了，
所以它的 `Err()` 是 `context.Canceled` 而不是 `context.DeadlineExceeded`。

三个必须记住的规则：

1. **`defer cancel()` 不能省。** 即使超时会自己触发，也要写，
   因为它保证函数提前返回时子节点立刻被释放，而不是拖到超时那一刻。
   `go vet` 的 `lostcancel` 检查就是专门抓这个的。
2. **`ctx.Done()` 是只读 channel，不能往里写。** 它只在取消时被关闭，
   `<-ctx.Done()` 因此是标准的等待取消写法。
3. **`ctx.Err()` 只能在 Done 之后再读**，它返回 `Canceled` 或 `DeadlineExceeded`，
   是判断「为什么结束」的唯一可靠依据。

## 二、超时控制与错误判断

### 代码

```go
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
```

### 解析

60ms 超时、200ms 的慢调用，所以必然是超时返回。这里有两个值得注意的点。

**第一，错误要 `%w` 包装，判断用 `errors.Is`。** 因为 `ctx.Err()` 的语义不同，
上层的处理策略也不同：

```go
switch {
case errors.Is(err, context.DeadlineExceeded):
	// 超时：重试通常有意义，也可以上报为依赖超时
case errors.Is(err, context.Canceled):
	// 调用方主动取消：一般不该重试（例如客户端断开了）
}
```

**第二，`done` channel 必须带缓冲 1。** 这点第 3 篇已经讲过：超时后调用方走了，
如果那个 goroutine 发送到无缓冲 channel，就会永久阻塞成泄漏。

**第三，也是最重要的：示例只是「本地放弃等待」，真正的取消要传给下游。**
生产代码里应该这样写：

```go
req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
...
rows, err := db.QueryContext(ctx, "select ...")
```

否则「取消」只发生在你这层，下游还在继续烧资源。

## 三、context.Value：只放请求级元数据

### 代码

```go
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
```

### 解析

可以放：request id、trace id、认证主体、租户信息——这些是**跨所有层都要用、
又不属于业务参数**的元数据。

不要放：业务参数、可选配置、数据库连接、logger（用显式注入更好）。
判断标准是：**换个调用者时这个值还成立吗？** 不成立就不该放 context。

另外 key 一定要用自定义类型，不要用裸 `string`：

```go
ctx = context.WithValue(ctx, "id", 1)   // 反例：任何包都可能覆盖它
```

`ctx.Value` 返回 `any`，取值时要写类型断言 `id, ok := ctx.Value(k).(string)`，
别写不带 `ok` 的版本（key 不存在时会 panic）。

## 四、errgroup：一组任务，任一失败就取消

### 代码

```go
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
```

```go
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
```

### 解析

输出里三个「任务 X 被取消」很关键：任务 4 在 20ms 时失败，
`errOnce.Do` 触发 `cancel()`，当时还在等待的 1/2/3 号任务立刻从 `ctx.Done()` 退出。
这就是「fail fast」——不要等到所有任务都跑完，因为失败之后的等待往往毫无意义。

三个实现细节值得注意：

1. **`errOnce` 保证只记录第一个错误。** 如果多个任务同时失败，
   后来的错误会被丢弃，`Wait()` 返回的是最先到达的那个。
2. **`Wait()` 里也要 `cancel()`。** 所有任务都成功时不会有别的机会调用 cancel，
   不调用就泄漏了 context 资源（`go vet` 会报 `lostcancel`）。
3. **`g.Go` 里的闭包会捕获循环变量 `i`**——在 Go 1.22 之后这是安全的
   （见第 1 篇），老代码里通常能看到 `i := i` 的写法。

生产代码请直接用 `golang.org/x/sync/errgroup`，它还提供并发上限：

```go
g, ctx := errgroup.WithContext(ctx)
g.SetLimit(8) // 最多 8 个任务并发
for _, url := range urls {
	g.Go(func() error { return fetch(ctx, url) })
}
err := g.Wait()
```

## 五、优雅退出：先停止，再排空

### 代码

```go
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
```

### 解析

消费端的退出分两步：**收到停止信号后不再接新活，但要先把已经在队列里的处理完**。
这就是「优雅」的含义——对调用方来说，正在进行的请求都有明确结果，
而不是被莫名其妙地切断。

注意 `close(queue)` 的作用是给排空循环一个终止条件：
`for v := range queue` 只有在 channel 关闭且读空之后才会结束。

## 六、HTTP 服务的优雅退出（完整示例）

生产环境的对应写法是 `signal.NotifyContext` + `http.Server.Shutdown`。
下面是完整代码，程序启动一个服务、发一个慢请求、然后给自己发 SIGTERM，
观察在途请求是否会处理完。

```go
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(300 * time.Millisecond):
			fmt.Fprintln(w, "处理完成")
		case <-r.Context().Done():
			// Shutdown 不会打断在途请求；如果客户端断开，这里会被触发
			fmt.Fprintln(w, "客户端提前断开")
		}
	})

	// 监听随机端口，避免和本机其它服务冲突
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	srv := &http.Server{Handler: mux}

	go func() {
		// Shutdown 会让 Serve 返回 http.ErrServerClosed
		_ = srv.Serve(ln)
	}()
	fmt.Println("→ 服务已启动:", ln.Addr())

	// NotifyContext：收到 SIGINT/SIGTERM 时自动取消 ctx
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 发起一个慢请求，让它停在处理中
	response := make(chan string, 1)
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String() + "/slow")
		if err != nil {
			response <- "请求出错: " + err.Error()
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		response <- strings.TrimSpace(string(body))
	}()

	time.Sleep(50 * time.Millisecond) // 让请求进入处理逻辑
	fmt.Println("→ 模拟运维发出 SIGTERM（真实场景由部署系统发送）")
	if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
		log.Fatal(err)
	}

	<-ctx.Done()
	fmt.Println("→ 收到停止信号，开始优雅退出")

	start := time.Now()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("优雅退出失败:", err)
	}

	fmt.Printf("→ 服务已退出，等待在途请求耗时 %v\n", time.Since(start).Round(10*time.Millisecond))
	fmt.Println("→ 在途请求的结果:", <-response)
	fmt.Println("→ 对比：如果调用 srv.Close()，在途请求会被直接切断")
}
```

运行输出：

```text
→ 服务已启动: 127.0.0.1:60504
→ 模拟运维发出 SIGTERM（真实场景由部署系统发送）
→ 收到停止信号，开始优雅退出
→ 服务已退出，等待在途请求耗时 270ms
→ 在途请求的结果: 处理完成
→ 对比：如果调用 srv.Close()，在途请求会被直接切断
```

几个要点：

1. **`signal.NotifyContext` 把信号变成了一个 ctx**，不用自己维护 signal channel，
   也不用记 `signal.Stop`。
2. **`srv.Shutdown(ctx)` 会先停止接收新连接，再等待在途请求结束。**
   示例里那个 300ms 的请求被完整执行完了（等待 270ms，因为请求已经跑了 30ms）。
3. **`Shutdown` 自己也要一个超时。** 万一某个请求永远不返回，
   不能陪着它一起挂住，2 秒之后强制退出即可。
4. **对比 `Close()`**：它立刻关闭所有连接，在途请求收不到任何响应。
   对用户来说就是「明明点下去了却被 reset」。

注意示例里的 `syscall.Kill` 是给自己发信号，只在类 Unix 系统上可用；
真实服务里不写这段，由部署系统（K8s、systemd）发信号。

## 运行输出

```text
[1] context 的取消会级联到所有子节点
    取消父节点 →
    child1 收到取消: context canceled
    child2 收到取消: context canceled
    child1.Err() = context canceled
    child2.Err() = context canceled （父节点先取消，超时就没机会触发了）
    结论：defer cancel() 是必须的，它能释放整棵子树的资源

[2] 超时控制
    耗时 60ms，结果 ""，错误 调用被取消: context deadline exceeded
    errors.Is(err, context.DeadlineExceeded) = true
    区分两种原因：DeadlineExceeded（超时）还是 Canceled（调用方主动取消）

[3] context.Value：只放请求级元数据
    在处理请求: req-42
    可以放：request id、trace id、认证信息
    不要放：业务参数、数据库连接、可选配置 —— 这些应该出现在函数签名里

[4] errgroup：一组任务，任一失败就取消其余
    任务4 失败 → 触发取消
    任务1 被取消: context canceled
    任务3 被取消: context canceled
    任务2 被取消: context canceled
    最终返回第一个错误: 数据库连接失败

[5] 优雅退出：先停止接收新任务，再排空在途任务
    处理完一个任务: 1
    收到停止信号，改为排空剩余任务
    处理完一个在途任务: 2
    处理完一个在途任务: 3
    排空完成，退出
    退出时没有丢掉任何任务
    真实服务里的对应做法：signal.NotifyContext + http.Server.Shutdown
```

## 常见错误

| 写法 | 后果 |
|---|---|
| 创建 ctx 不 `defer cancel()` | 资源泄漏，`go vet` 报 `lostcancel` |
| 把 ctx 存进 struct 字段 | 生命周期混乱；ctx 应该作为函数第一个参数 |
| 传 `nil` 给需要 ctx 的函数 | 直接 panic |
| 用 `context.Background()` 替换传入的 ctx | 取消链路断掉，超时失效 |
| `Value` 放业务参数 | 依赖关系隐式化，代码难读难测 |
| 只用超时不用取消 | 上游仍然在烧资源 |
| 用 `Close()` 代替 `Shutdown()` | 在途请求被切断 |

`ctx` 的传递约定已经写进了 Go 的社区规范：

```go
func fetch(ctx context.Context, url string) ([]byte, error) // ctx 永远是第一个参数
```

## 练习

1. 把第 5 篇的 pipeline 全部改造成 `context` 版本，并对比 `done` channel 与
   `ctx.Done()` 的写法差异。
2. 用 `errgroup` 实现「并发请求 10 个 URL，最多 3 个并发，任一失败立即取消，
   返回第一个错误」。
3. 给 `graceful` 示例加上一个「卡住的请求」，把 `Shutdown` 超时改成 1 秒，
   观察它超时后如何强制退出。

## 小结

- 取消沿 context 树向下传播，父节点取消会带走所有子节点。
- 用 `errors.Is(err, context.DeadlineExceeded)` 区分超时与主动取消。
- `defer cancel()` 是硬性要求，不是可选项。
- `context.Value` 只放请求级元数据，key 用自定义类型。
- errgroup 把「任一失败即取消 + 返回第一个错误」标准化了。
- 优雅退出 = 停止接收新请求 + 等待在途请求 + 强制超时兜底。

下一篇我们把视角从「怎么写对」转到「怎么查错」：七个最常见的并发陷阱。
