# 诊断与调优工具箱

并发问题之所以难，是因为它们通常不可见。这一篇把 Go 自带的四件工具系统化地用一遍：
`-race`、`pprof`、`runtime/trace`、基准测试。

**配套代码**：`examples/08_diagnostics/main.go`、`examples/08_diagnostics/bench_test.go`

```bash
cd go
go run ./base/08_concurrency/examples/08_diagnostics
go test -bench . -benchmem ./base/08_concurrency/examples/08_diagnostics
```

---

## 一、pprof HTTP 端点

### 代码

```go
import (
	"net/http"
	_ "net/http/pprof" // 通过副作用注册 /debug/pprof/* 到 DefaultServeMux
)

func demoPprofServer() {
	fmt.Println("\n[1] pprof HTTP 端点")

	go func() {
		_ = http.ListenAndServe("127.0.0.1:6060", nil) // 端口被占用时会静默失败，不影响其它演示
	}()

	fmt.Println("    已尝试启动 http://127.0.0.1:6060/debug/pprof/")
	fmt.Println("    go tool pprof http://127.0.0.1:6060/debug/pprof/profile?seconds=5   # CPU")
	fmt.Println("    go tool pprof http://127.0.0.1:6060/debug/pprof/heap              # 内存")
	fmt.Println("    curl 'http://127.0.0.1:6060/debug/pprof/goroutine?debug=2'        # 所有 goroutine 的栈")
}
```

### 解析

关键就是那一行下划线 import：`_ "net/http/pprof"`。
它的 `init()` 会把 `/debug/pprof/*` 注册到 `http.DefaultServeMux`，
所以只要在服务里再起一个 HTTP server（或者复用现有 server 的 mux），
就能拿到运行时的全部诊断数据。

真实项目里的写法通常是**单独开一个端口**，避免暴露在主站入口：

```go
// 主服务监听 8080，诊断端口监听 127.0.0.1:6060（仅本机或内网可访问）
go func() {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/heap", pprof.Handler("heap").ServeHTTP)
	mux.HandleFunc("/debug/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)
	http.ListenAndServe("127.0.0.1:6060", mux)
}()
```

安全提醒：pprof 端点在 Go 1.22+ 默认只允许 localhost 访问（Host 头校验），
但仍然不要把 6060 直接暴露到公网——它能读到你的调用栈和内存数据。

## 二、CPU profile：找热点函数

### 代码

```go
// busyWork 制造 CPU 负载，让 profile 里有内容可看。
func busyWork(d time.Duration) {
	deadline := time.Now().Add(d)
	sum := 0
	for time.Now().Before(deadline) {
		for i := 0; i < 10_000; i++ {
			sum += i % 7
		}
	}
	_ = sum
}

func demoCPUProfile() {
	fmt.Println("\n[2] CPU profile：找热点函数")

	f, err := os.CreateTemp("", "cpu-*.pprof")
	if err != nil {
		fmt.Println("    创建临时文件失败:", err)
		return
	}
	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		fmt.Println("    启动失败:", err)
		return
	}
	busyWork(300 * time.Millisecond)
	pprof.StopCPUProfile()

	fmt.Printf("    profile 已写入 %s\n", f.Name())
	fmt.Printf("    go tool pprof -http=:8080 %s\n", f.Name())
}
```

### 解析

CPU profile 回答的问题是「**时间花在哪个函数上**」。两种采集方式：

| 方式 | 用法 |
|---|---|
| 代码里手动采集 | `pprof.StartCPUProfile(f)` / `StopCPUProfile()`，适合测试和 CLI 程序 |
| HTTP 端点 | `go tool pprof http://host/debug/pprof/profile?seconds=30`，适合线上服务 |

采集完用 `go tool pprof -http=:8080 <file>` 打开浏览器，最常用的三个视图：

- **Top**：按 CPU 时间排序的函数列表，找热点从这里开始
- **Flame Graph**：火焰图，看调用链上的耗时分布
- **Source**：源码级的热点行，能看到具体哪一行最耗时

注意：CPU profile 采样的是**正在消耗 CPU 的代码**。如果程序卡在等待
（等锁、等网络、等 channel），CPU profile 会是空的——那种情况要看 block/mutex
profile 和 trace（下面会讲）。

## 三、goroutine profile：排查泄漏

### 代码

```go
func demoGoroutineProfile() {
	fmt.Println("\n[3] goroutine profile：排查泄漏最直接的手段")

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(200 * time.Millisecond)
		}()
	}
	time.Sleep(20 * time.Millisecond) // 让它们进入 sleep 状态

	var buf bytes.Buffer
	if err := pprof.Lookup("goroutine").WriteTo(&buf, 1); err != nil { // 1 = 带调用栈
		fmt.Println("    写入失败:", err)
		wg.Wait()
		return
	}

	firstLine := strings.SplitN(buf.String(), "\n", 2)[0]
	fmt.Printf("    当前 goroutine 数 = %d，profile 大小 = %d 字节\n", runtime.NumGoroutine(), buf.Len())
	fmt.Println("    profile 的第一行:", firstLine)

	wg.Wait()
}
```

### 解析

`pprof.Lookup("goroutine").WriteTo(&buf, 1)` 里的 `1` 表示「带调用栈」，
写 `0` 只输出汇总计数。第一行是总数：

```text
goroutine profile: total 7
```

排查泄漏的标准动作是**隔一段时间抓两次，对比**：

```bash
# 抓两次快照
curl -s 'http://localhost:6060/debug/pprof/goroutine?debug=2' > /tmp/g1.txt
sleep 30
curl -s 'http://localhost:6060/debug/pprof/goroutine?debug=2' > /tmp/g2.txt

# 找出两次都停在同一行的 goroutine 数量变化
diff <(grep -o 'main\..*\.go:[0-9]*' /tmp/g1.txt | sort | uniq -c | sort -rn) \
     <(grep -o 'main\..*\.go:[0-9]*' /tmp/g2.txt | sort | uniq -c | sort -rn)
```

两次快照里都在增长、且位置相同的那批 goroutine，就是泄漏源。

## 四、runtime/trace：看时间线

### 代码

```go
// concurrentWork 制造可观测的阻塞与唤醒，方便在 trace 里看到调度。
func concurrentWork() {
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 3; j++ {
				time.Sleep(10 * time.Millisecond)
			}
		}()
	}
	wg.Wait()
}

func demoTrace() {
	fmt.Println("\n[4] runtime/trace：调度、GC、阻塞的时间线")

	f, err := os.CreateTemp("", "trace-*.out")
	if err != nil {
		fmt.Println("    创建临时文件失败:", err)
		return
	}
	defer f.Close()

	if err := trace.Start(f); err != nil {
		fmt.Println("    启动失败:", err)
		return
	}
	concurrentWork()
	trace.Stop()

	fmt.Printf("    trace 已写入 %s\n", f.Name())
	fmt.Printf("    go tool trace %s\n", f.Name())
}
```

### 解析

pprof 告诉你「哪个函数花的时间多」，trace 告诉你「**这段代码在时间轴上到底在干什么**」。
它是唯一能回答下面这些问题的工具：

- 这个 goroutine 为什么等了 2ms 才被调度？
- 这次延迟毛刺是不是 GC 造成的？
- P 的数量够不够？有没有 goroutine 在排队等 P？
- 系统调用阻塞了多久？

```bash
go tool trace /var/folders/.../trace-4282684880.out
```

浏览器里最常用的两个视图：**Goroutine analysis**（每个 goroutine 的时间分布）、
**Scheduler latency**（调度延迟）。采集时间一般 1~5 秒即可，
时间越长开销越大（会显著影响吞吐）。

经验法则：**延迟问题先用 trace，吞吐问题先用 pprof。**

## 五、运行期指标

### 代码

```go
func demoMemStats() {
	fmt.Println("\n[5] 运行期指标")

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Printf("    HeapAlloc = %d KiB，TotalAlloc = %d KiB，NumGC = %d，Goroutine = %d\n",
		m.HeapAlloc/1024, m.TotalAlloc/1024, m.NumGC, runtime.NumGoroutine())
	fmt.Println("    线上这些指标通常以 Prometheus 形式暴露：go_goroutines、go_memstats_*、go_gc_*")
	fmt.Println("    ↑ HeapAlloc 长期不降 + goroutine 数持续增长，就是需要看 profile 的信号")
}
```

### 解析

这几个指标足以做第一层判断：

| 指标 | 说明 | 异常信号 |
|---|---|---|
| `Goroutine` | 当前 goroutine 数 | 持续增长不回落 |
| `HeapAlloc` | 当前堆占用 | 长期不降，随请求线性增长 |
| `NumGC` | GC 次数 | 短时间内暴涨 |
| `TotalAlloc` | 累计分配量 | 增长快说明分配压力大 |

线上通常由 Prometheus 抓取 `go_goroutines`、`go_memstats_*`、`go_gc_*` 这些指标。
**看到异常再上 pprof**，不然容易陷入「为了排查而排查」。

注意 `runtime.MemStats` 每次读取都会 Stop-The-World 一小会儿，
高频调用有代价，生产里用 `metrics` 库的 `ReadMemStats` 采样即可（通常 15s 一次）。

## 六、基准测试：用数据说话

### 代码（`bench_test.go`）

```go
func BenchmarkCounterMutex(b *testing.B) {
	var (
		mu sync.Mutex
		n  int64
	)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			n++
			mu.Unlock()
		}
	})
	_ = n
}

func BenchmarkCounterAtomic(b *testing.B) {
	var n atomic.Int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n.Add(1)
		}
	})
	_ = n.Load()
}

func BenchmarkCounterChannel(b *testing.B) {
	// 用容量 1 的 channel 当锁：语义正确，但开销通常是最大的。
	ch := make(chan struct{}, 1)
	var n int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ch <- struct{}{}
			n++
			<-ch
		}
	})
	_ = n
}

func BenchmarkCounterSharded(b *testing.B) {
	var shards [16]struct {
		mu sync.Mutex
		n  int64
	}
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			s := &shards[i%len(shards)]
			s.mu.Lock()
			s.n++
			s.mu.Unlock()
			i++
		}
	})
}
```

### 真实输出（Apple M5 Pro，Go 1.27.1）

```text
goos: darwin
goarch: arm64
pkg: gobasics/base/08_concurrency/examples/08_diagnostics
cpu: Apple M5 Pro
BenchmarkCounterMutex-18      	14578431	        75.08 ns/op	       0 B/op	       0 allocs/op
BenchmarkCounterAtomic-18     	94703822	        12.61 ns/op	       0 B/op	       0 allocs/op
BenchmarkCounterChannel-18    	12087625	        96.97 ns/op	       0 B/op	       0 allocs/op
BenchmarkCounterSharded-18    	38231577	        32.54 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	gobasics/base/08_concurrency/examples/08_diagnostics	5.447s
```

### 怎么读这些数字

| 列 | 含义 |
|---|---|
| `-18` | `GOMAXPROCS` 值，即并行度 |
| `ns/op` | 每次操作平均耗时 |
| `B/op` | 每次操作分配的字节数 |
| `allocs/op` | 每次操作的分配次数 |

三个写法要点：

1. **`b.RunParallel`**：用 GOMAXPROCS 个 goroutine 并发调用，测的是竞争下的表现。
   只写普通 `for` 循环测不到锁竞争。
2. **关注 `allocs/op`**：分配次数往往比纳秒更能解释性能问题，
   因为它最终会变成 GC 压力。
3. **对比时用 `benchstat`**，不要肉眼比两次输出：

```bash
go test -bench . -benchmem -count 10 ./... > old.txt
# 修改代码
go test -bench . -benchmem -count 10 ./... > new.txt
benchstat old.txt new.txt      # 会给出统计显著性判断
```

注意：这些数字是**量级参考**，不是选型依据。每秒只调用几百次的地方，
75ns 和 13ns 的差别毫无意义，可读性和正确性才是第一优先级。

## 把所有演示串起来

```go
func main() {
	fmt.Printf("Go %s，GOMAXPROCS = %d，逻辑 CPU = %d\n",
		runtime.Version(), runtime.GOMAXPROCS(0), runtime.NumCPU())

	demoPprofServer()
	demoCPUProfile()
	demoGoroutineProfile()
	demoTrace()
	demoMemStats()
}
```

## 运行输出

```text
Go go1.27.1，GOMAXPROCS = 18，逻辑 CPU = 18

[1] pprof HTTP 端点
    已尝试启动 http://127.0.0.1:6060/debug/pprof/
    go tool pprof http://127.0.0.1:6060/debug/pprof/profile?seconds=5   # CPU
    go tool pprof http://127.0.0.1:6060/debug/pprof/heap              # 内存
    curl 'http://127.0.0.1:6060/debug/pprof/goroutine?debug=2'        # 所有 goroutine 的栈

[2] CPU profile：找热点函数
    profile 已写入 /var/folders/.../cpu-3109424314.pprof
    go tool pprof -http=:8080 /var/folders/.../cpu-3109424314.pprof

[3] goroutine profile：排查泄漏最直接的手段
    当前 goroutine 数 = 7，profile 大小 = 2637 字节
    profile 的第一行: goroutine profile: total 7

[4] runtime/trace：调度、GC、阻塞的时间线
    trace 已写入 /var/folders/.../trace-4282684880.out
    go tool trace /var/folders/.../trace-4282684880.out

[5] 运行期指标
    HeapAlloc = 2699 KiB，TotalAlloc = 4709 KiB，NumGC = 1，Goroutine = 2
    线上这些指标通常以 Prometheus 形式暴露：go_goroutines、go_memstats_*、go_gc_*
    ↑ HeapAlloc 长期不降 + goroutine 数持续增长，就是需要看 profile 的信号
```

## 四种 profile 分别看什么

| profile | 采集方式 | 回答的问题 |
|---|---|---|
| goroutine | 默认开启 | 每个 goroutine 卡在哪（泄漏首选） |
| heap | 默认开启 | 内存分配热点、疑似泄漏的对象 |
| block | 需 `runtime.SetBlockProfileRate(1)` | 谁在阻塞（channel、锁、select） |
| mutex | 需 `runtime.SetMutexProfileFraction(5)` | 哪把锁竞争最严重 |

后两个默认关闭，因为采样有开销。怀疑「锁竞争导致延迟毛刺」再打开：

```go
runtime.SetMutexProfileFraction(5)  // 采样 1/5 的锁竞争事件
runtime.SetBlockProfileRate(1)      // 采样全部阻塞事件（开销大，短时间用）
```

## 工具选择速查

| 你想知道 | 用什么 |
|---|---|
| 有没有数据竞争 | `go test -race ./...` |
| 哪个函数在烧 CPU | CPU profile |
| 内存/goroutine 是否泄漏 | heap / goroutine profile |
| 谁在等锁、谁在阻塞 | mutex / block profile |
| 为什么这个请求慢了 3ms | `runtime/trace` |
| 这个优化到底有没有效果 | `testing.B` + `benchstat` |
| 服务当前健康吗 | `runtime.MemStats` / Prometheus 指标 |

## 练习

1. 给第 5 篇的 `workerPool` 写 benchmark，对比 worker 数为 1/4/16/64 时的吞吐。
2. 制造一个「锁竞争导致 P99 变高」的场景，用 mutex profile 定位，再用分片锁修复并验证。
3. 用 trace 观察一次 `time.Sleep` 与一次 CPU 密集循环的差别，解释为什么它们的
   CPU profile 表现完全不同。

## 小结

- `-race` 必须进入日常命令，它抓的是「不报错的错误」。
- pprof 有四类 profile：goroutine、heap、block、mutex。
- trace 看时间线，pprof 看总量。
- 跑基准测试用 `b.RunParallel`，对比用 `benchstat`。
- 先有指标异常，再上 profile；不要凭感觉优化。

## 这套教程到此结束

回头看一遍整体脉络：

1. goroutine 与 WaitGroup：谁启动，谁等它结束。
2. channel：握手、队列、广播、所有权。
3. select：让每个等待都有退出路径。
4. 锁与原子操作：先想清楚保护的是不是一个变量。
5. 五种并发模式：worker pool、pipeline、fan-in、限流、singleflight。
6. context：把取消、超时、生命周期贯穿到整条调用链。
7. 陷阱：泄漏、死锁、竞争都是可排查的。
8. 诊断：让问题可见。

接下来最有价值的动作是**读源码**：`sync` 包、`golang.org/x/sync/errgroup`、
`golang.org/x/sync/singleflight`，以及你正在用的框架里的并发实现。
