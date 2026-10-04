# 七个最常见的并发陷阱

前面六篇讲「怎么写对」，这篇讲「怎么查错」。
并发 bug 的特点是：**不崩溃、不报错、只在生产环境偶发**。
所以这一篇的重点不是背结论，而是掌握把问题变可见的手段。

**配套代码**：`examples/07_pitfalls/main.go`（干净版本）、
`examples/07_pitfalls/broken/`（故意写错的版本）、`examples/07_pitfalls/leak_test.go`

```bash
cd go
go run ./base/08_concurrency/examples/07_pitfalls

# 故意写错的反例，单独运行（文件带 //go:build ignore，不参与 go build ./...）
go run -race ./base/08_concurrency/examples/07_pitfalls/broken/race_counter.go
go run ./base/08_concurrency/examples/07_pitfalls/broken/deadlock.go
go run ./base/08_concurrency/examples/07_pitfalls/broken/map_write.go
```

---

## 陷阱一：goroutine 泄漏

### 代码

```go
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
```

### 运行证据

```text
基线 1 → 泄漏后 51
```

**这 50 个 goroutine 永远不会消失**，它们各自保留着自己的栈（初始 2KB，可以更大）。
在真实服务里，这个数字会随着每个请求、每个 tick 慢慢爬升，最后表现为内存只涨不降。

### 修复：给每个 goroutine 一条退路

```go
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
```

### 判定标准

写下一个 `go` 之前，先回答：**这个 goroutine 什么时候、由谁负责结束？**

对阻塞在 send/receive 上的 goroutine，退出路径只有三种：

1. 有人接收/发送（正常完成）
2. channel 被关闭
3. `select` 里有一个 `<-done` 或 `<-ctx.Done()` 分支

只有第 3 种是你可控的。凡是「可能没人接收」的发送，都必须配上取消分支。

## 陷阱二：不知道 goroutine 卡在哪

### 代码

```go
func demoStackDump() {
	fmt.Println("\n[3] 排查手段：打印所有 goroutine 的栈")

	buf := make([]byte, 1<<20)
	n := runtime.Stack(buf, true) // true = 抓取所有 goroutine
	stacks := string(buf[:n])

	blocked := strings.Count(stacks, "chan send")
	fmt.Printf("    快照 %d 字节，其中 %d 处卡在「chan send」（就是第 1 步留下的那批）\n", n, blocked)
	fmt.Println("    线上更省事的做法：curl 'http://localhost:6060/debug/pprof/goroutine?debug=2'")
}
```

### 解析

`runtime.Stack(buf, true)` 一次性抓取所有 goroutine 的调用栈，
输出里每个 goroutine 的状态方括号就是它的处境，例如：

```text
goroutine 24 [chan send]:
main.leak.func1()
	.../07_pitfalls/main.go:45 +0x38
created by main.leak in goroutine 1
```

状态名是最好的线索：

| 状态 | 含义 |
|---|---|
| `chan send` / `chan receive` | 卡在 channel 收发上 |
| `chan receive (nil chan)` | 用的 channel 是 nil（忘了初始化） |
| `select` | 卡在 select 上 |
| `sync.Mutex.Lock` | 在等锁（可能是死锁，也可能是高竞争） |
| `IO wait` | 在等网络/磁盘 |
| `semacquire` | 等信号量或 WaitGroup |

线上不重启进程的排查方式：

```bash
curl 'http://localhost:6060/debug/pprof/goroutine?debug=2'
```

实践建议：**先把泄漏 goroutine 的栈打出来对比两次**，两次都在同一行的，
就是泄漏点。

## 陷阱三：死锁

### 代码

安全版本（用超时兜底，不会真的卡死）：

```go
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
```

反例（`broken/deadlock.go`，两个 goroutine 以相反顺序加两把锁）：

```go
func main() {
	var (
		muA, muB sync.Mutex
		wg       sync.WaitGroup
	)

	wg.Add(2)

	go func() { // 顺序：A → B
		defer wg.Done()
		muA.Lock()
		defer muA.Unlock()

		time.Sleep(10 * time.Millisecond)

		muB.Lock()
		defer muB.Unlock()
		fmt.Println("goroutine 1 完成")
	}()

	go func() { // 顺序：B → A，正好相反
		defer wg.Done()
		muB.Lock()
		defer muB.Unlock()

		time.Sleep(10 * time.Millisecond)

		muA.Lock()
		defer muA.Unlock()
		fmt.Println("goroutine 2 完成")
	}()

	wg.Wait()
	fmt.Println("没崩溃说明这次恰好没死锁（时序问题，多跑几次就会出现）")
}
```

运行结果是：

```text
fatal error: all goroutines are asleep - deadlock!
```

### 解析

请注意运行时检测死锁的条件非常苛刻：**只有当所有 goroutine 都睡着时才报错**。
真实服务里有 HTTP 服务器、后台任务一直在跑，总有人醒着，
所以死锁通常表现为「部分请求永远挂起」——不报错、不崩溃、不会自动恢复。

预防手段比排查手段更重要：

- 全局约定加锁顺序（按 ID、按地址排序后再加锁）
- 临界区里不要调用可能回调你自己代码的函数
- 需要多个资源时，采用「一次只持有一把锁」的写法
- 关键路径加超时兜底（`select + time.After`），至少不会永久挂住

## 陷阱四：数据竞争

### 反例代码（`broken/race_counter.go`）

```go
func main() {
	var total int
	var wg sync.WaitGroup

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100_000; j++ {
				total++ // 读取、加一、写回，三步之间会被其它 goroutine 插进来
			}
		}()
	}
	wg.Wait()

	fmt.Println("结果:", total, "期望: 800000")
	fmt.Println("少掉的那些就是被覆盖掉的更新 —— 这就是数据竞争（data race）")
}
```

普通运行会打印一个小于 800000 的数字，而带 `-race` 运行会给出精确报告：

```text
==================
WARNING: DATA RACE
Read at 0x00c0000ba038 by goroutine 7:
  main.main.func1()
      .../broken/race_counter.go:26 +0x88

Previous write at 0x00c0000ba038 by goroutine 6:
  main.main.func1()
      .../broken/race_counter.go:26 +0x98
==================
```

### 解析

数据竞争最危险的地方是**它可能什么都不报**：程序照常运行，只是账对不上、
结果偶尔不同。所以 `-race` 不是「出问题才打开」的调试开关，而是日常命令：

```bash
go test -race ./...     # CI 里的标准做法
go run -race ./cmd/app  # 本地复现时
```

带 `-race` 的程序在检测到竞争时会打印报告并以退出码 66 结束。
它需要代码真的执行到那两处冲突的访问才能发现，所以测试覆盖率越高，效果越好。

## 陷阱五：并发写 map

### 反例代码（`broken/map_write.go`）

```go
func main() {
	m := map[int]int{}
	var wg sync.WaitGroup

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 10_000; j++ {
				m[i] = j // map 不是并发安全的
			}
		}(i)
	}
	wg.Wait()

	fmt.Println("map 大小:", len(m))
}
```

运行结果：

```text
fatal error: concurrent map writes
```

### 两种并发错误的区别

| 问题 | 表现 | 能否 recover |
|---|---|---|
| 数据竞争（data race） | 结果错误、偶发、可能什么都不报 | 能，但代码本身已经错了 |
| 并发写 map | 运行时立即 fatal error | 不能，进程直接终止 |

`concurrent map writes` 是运行时主动终止进程，`recover` 救不回来。
所以 map 的并发访问必须提前设计好：

| 方案 | 适用场景 |
|---|---|
| `map + sync.RWMutex` | 大多数情况，锁粒度可控 |
| `sync.Map` | 读多写少、key 集合基本固定（比如缓存元数据） |
| 分片 map + 分片锁 | 高并发写入、key 分散 |
| 让每个 goroutine 持有自己的 map，最后合并 | 可以批量合并的场景，最快 |

## 陷阱六：WaitGroup 时机与关闭顺序

### 代码

```go
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
```

### 解析

这两条规则在第 1、2、5 篇都出现过，这里再强调一次，因为它们是最高频的踩坑点：

1. `wg.Add(1)` 必须在 `go` 之前调用，且必须在同一个 goroutine 里调用。
2. `close` 必须发生在所有发送者结束之后。保险写法是「等 WaitGroup 归零的 goroutine 负责关」：

```go
go func() {
	wg.Wait()
	close(results)
}()
```

3. `WaitGroup` 的计数器不允许出现负数（`Done` 多于 `Add`），会直接 panic。

## 陷阱七：定时器与 context 的资源泄漏

### 代码

```go
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
```

### 解析

这一类的「资源泄漏」不会直接崩，但会慢慢拖垮进程：

| 泄漏对象 | 原因 | 修复 |
|---|---|---|
| timer / ticker | 忘了 `Stop`，或循环里反复 `time.After` | `defer t.Stop()`，复用定时器 |
| context | 创建了 `WithCancel` 却不调用 cancel | `defer cancel()`，`go vet` 会提示 |
| HTTP body / rows | 忘了 `Close`，连接池被占满 | `defer resp.Body.Close()` |
| goroutine | 没有退出路径 | `select` 里加 `<-ctx.Done()` |

它们往往同时出现：一个忘了 `Close` 的请求，会让连接池耗尽，
进而让所有等待连接的 goroutine 堆积——最终表现成「服务突然卡死」。

## 在测试里断言「没有泄漏」

```go
// TestNoGoroutineLeak 演示如何在测试里断言「没有泄漏」。
//
// 生产项目推荐直接用 go.uber.org/goleak，它会忽略运行时自身产生的 goroutine，
// 断言更严格也更稳定：
//
//	func TestMain(m *testing.M) {
//	    goleak.VerifyTestMain(m)
//	}
func TestNoGoroutineLeak(t *testing.T) {
	base := runtime.NumGoroutine()

	done := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())

	go func() { // 有明确退出路径的 worker
		defer close(done)
		<-ctx.Done()
	}()

	cancel()
	<-done

	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > base && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if got := runtime.NumGoroutine(); got > base {
		t.Fatalf("疑似 goroutine 泄漏: 基线 %d, 现在 %d", base, got)
	}
}
```

手写断言的缺点是容易受测试框架自身 goroutine 的干扰，
所以生产项目建议直接用 `go.uber.org/goleak`：

```go
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m) // 每个测试结束后自动检查
}
```

## 运行输出

```text
[1] goroutine 泄漏：没有退出路径的 goroutine
    基线 1 → 泄漏后 51（这 50 个 goroutine 会一直活着）
    它们各自带着栈内存，这就是「内存缓慢增长」的常见来源

[2] 修复：给每个 goroutine 一条退路
    同样启动 50 个带退路的 goroutine：51 → 51（全部正常退出，没有增加）
    规范：写下 go 之前，先回答「这个 goroutine 什么时候、由谁负责结束」

[3] 排查手段：打印所有 goroutine 的栈
    快照 13690 字节，其中 50 处卡在「chan send」（就是第 1 步留下的那批）
    线上更省事的做法：curl 'http://localhost:6060/debug/pprof/goroutine?debug=2'

[4] 死锁：运行时只在「所有 goroutine 都睡死」时报错
    收到: 1

[5] WaitGroup 的 Add 必须在 Wait 之前完成
    正确写法的结果: [0 1 4] —— main 返回前所有任务都已结束

[6] 关闭顺序：等发送者都退出后再关 channel
    结果总和: 60

[7] 定时器：不用了要 Stop
    心跳 1
    心跳 2
    心跳 3
    退出，定时器已 Stop

把 -race 加进日常命令：go test -race ./...，出问题前先让它报出来
```

## 排查流程速查

遇到问题时按这个顺序走，绝大多数并发问题都能定位：

| 症状 | 第一步做什么 |
|---|---|
| 内存只涨不降 | `curl .../debug/pprof/goroutine?debug=2`，看 goroutine 数是否持续增长 |
| 服务挂起、请求不返回 | 打印所有 goroutine 栈，找 `sync.Mutex.Lock` / `select` 的堆积点 |
| 结果偶发不对 | `go test -race ./...` |
| CPU 100% | CPU profile + 检查是否有 `for { select { default } }` 忙轮询 |
| 延迟毛刺 | mutex profile + `runtime/trace` |
| 进程直接退出 | 看 fatal error 类型：`concurrent map writes`、`deadlock`、`send on closed channel` |

## 练习

1. 给你自己的项目加上 `goleak.VerifyTestMain(m)`，看能不能抓出隐藏的泄漏。
2. 故意写一个「goroutine 数只增不减」的服务，然后用 pprof 找到泄漏点并修复。
3. 把 `broken/deadlock.go` 改成「按固定顺序加锁」的版本，验证死锁消失。

## 小结

- 每个 goroutine 都必须有可到达的退出路径，否则就是泄漏。
- `runtime.Stack` / pprof 的 goroutine profile 是排查泄漏的第一工具。
- 死锁只在「所有 goroutine 都睡死」时才报错，真实服务里往往表现为挂起。
- 数据竞争不一定报错，所以 `-race` 要常开。
- 并发写 map 是 fatal error，必须提前设计好同步方案。
- `Add` 在 `go` 之前，`close` 在所有发送者结束之后。
- 定时器、context、body、goroutine 都要能回答「谁负责释放」。

下一篇是最后一篇：把诊断工具系统化地用起来。
