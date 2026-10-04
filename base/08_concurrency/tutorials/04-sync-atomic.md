# 锁、原子操作与其它同步原语

Go 宣扬「用通信代替共享内存」，但真实项目里该加锁的地方还是得加锁。
这篇把 `Mutex`、`RWMutex`、`atomic`、`Once`、`Pool` 和分片锁放在一起讲清楚：
每一种适合什么形状的问题，以及为什么选错会出事故。

**配套代码**：`examples/04_sync_atomic/main.go`

```bash
cd go
go run ./base/08_concurrency/examples/04_sync_atomic
go run -race ./base/08_concurrency/examples/04_sync_atomic   # 本文代码应该是干净的
```

---

## 一、Mutex：保护的是临界区，不是变量

### 代码

```go
const (
	workers   = 8
	perWorker = 50_000
)

// counterMutex 用互斥锁保护共享变量。
func counterMutex() int64 {
	var (
		mu    sync.Mutex
		total int64
		wg    sync.WaitGroup
	)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				mu.Lock()
				total++ // 临界区只放必需的这一行
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	return total
}

func demoMutex() {
	fmt.Println("[1] Mutex：保护临界区")
	start := time.Now()
	got, want := counterMutex(), int64(workers*perWorker)
	fmt.Printf("    结果 = %d，期望 = %d，正确 = %v，耗时 %v\n",
		got, want, got == want, time.Since(start).Round(time.Millisecond))
	fmt.Println("    对照实验（会报 data race）：go run -race ./base/08_concurrency/examples/07_pitfalls/broken/race_counter.go")
	fmt.Println("    锁的粒度要点：临界区只放必需的几行，绝不在锁里做 IO 或调用可能阻塞的函数")
}
```

### 解析

`total++` 看着是一行，实际是「读 → 加一 → 写回」三步。8 个 goroutine 同时执行时，
彼此的中间结果会互相覆盖——这就是数据竞争，也是为什么必须加锁。

真正要建立的观念是：**锁保护的是临界区，不是变量**。
规则是：所有访问这块共享状态的地方都必须持有同一把锁。
漏掉任何一处（比如某个函数忘了加锁就读 `total`），锁就白加了。

临界区要「足够小」也要「足够全」：

- 小：只包住必要的那几行，绝对不要在锁里做网络 IO、查数据库、写日志文件
- 全：所有读写路径都覆盖到

## 二、RWMutex：读多写少

### 代码

```go
// cache 是读多写少的场景，用 RWMutex 允许多个读并发进行。
type cache struct {
	mu sync.RWMutex
	m  map[string]string
}

func newCache() *cache { return &cache{m: make(map[string]string)} }

func (c *cache) Get(k string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.m[k] // 多个 goroutine 可以同时进入读锁
	return v, ok
}

func (c *cache) Set(k, v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[k] = v
}
```

配套的读并发演示：

```go
func demoRWMutex() {
	fmt.Println("\n[2] RWMutex：读多写少")

	c := newCache()
	c.Set("go", "goroutine")

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = c.Get("go") // 5 个读者可以同时在读锁里
		}()
	}
	wg.Wait()

	v, _ := c.Get("go")
	fmt.Println("    读到的值:", v)
	fmt.Println("    写锁要等所有读锁释放；读锁里做耗时操作会拖慢写，写非常频繁时 RWMutex 可能比 Mutex 更慢")
}
```

### 解析

`RWMutex` 允许任意多个读者同时进入，写者必须独占。它适合「读远多于写」的场景，
比如配置缓存、路由表、元数据查询。

三个容易踩的点：

1. **读锁里不要做耗时操作。** 写锁要等所有读者释放，一个慢读者会拖住整个写路径。
2. **写很频繁时它反而更慢。** `RWMutex` 自身的记账开销比 `Mutex` 大，
   读写比例接近 1:1 时，用 `Mutex` 往往更好。
3. **不要嵌套或递归加锁**，`RLock` 之后再 `Lock` 会直接死锁。

## 三、atomic：单个变量的无锁更新

### 代码

```go
// counterAtomic 用原子变量做同样的计数。
func counterAtomic() int64 {
	var (
		total atomic.Int64
		wg    sync.WaitGroup
	)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				total.Add(1)
			}
		}()
	}

	wg.Wait()
	return total.Load()
}

func demoAtomic() {
	fmt.Println("\n[3] atomic：单个变量的无锁更新")
	got, want := counterAtomic(), int64(workers*perWorker)
	fmt.Printf("    结果 = %d，期望 = %d，正确 = %v\n", got, want, got == want)
	fmt.Println("    适合：计数器、开关标志、状态机、指针整体替换（atomic.Pointer）")
	fmt.Println("    不适合：需要多个变量保持一致（那还是得用锁）")
}
```

### 解析

`atomic.Int64` 这类类型（Go 1.19+ 提供）把原子操作包装成方法，比裸的
`atomic.AddInt64(&n, 1)` 更安全也更易读。常用成员：`Add`、`Load`、`Store`、
`CompareAndSwap`、`Swap`。

它的适用形状非常明确：**一个变量，一个操作**。计数器、开关、状态枚举、
`atomic.Pointer[T]` 整体替换配置对象，都是它的主场。

一旦你说出「这两个字段必须一起更新」，`atomic` 就不适用了——两个独立的原子变量之间
没有任何一致性保证，观察者可能看到只更新了一半的状态。

## 四、CAS：读-改-写必须原子完成

### 代码

```go
// takeQuota 用 CAS 循环实现名额扣减：无锁，且绝不会超发。
func takeQuota(quota *atomic.Int64) (left int64, ok bool) {
	for {
		cur := quota.Load()
		if cur <= 0 {
			return 0, false
		}
		if quota.CompareAndSwap(cur, cur-1) { // 期间没人改过，才算扣减成功
			return cur - 1, true
		}
		// 失败说明被别人抢先改了，重读当前值后重试
	}
}

func demoCAS() {
	fmt.Println("\n[4] CAS：读-改-写必须原子完成的场景")

	var quota atomic.Int64
	quota.Store(3) // 一共只有 3 个名额

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results []string
	)

	for i := 1; i <= 6; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			left, ok := takeQuota(&quota)

			mu.Lock()
			defer mu.Unlock()
			if ok {
				results = append(results, fmt.Sprintf("请求%d 拿到名额(剩余%d)", i, left))
			} else {
				results = append(results, fmt.Sprintf("请求%d 被拒", i))
			}
		}(i)
	}
	wg.Wait()

	sort.Strings(results)
	fmt.Println("    " + strings.Join(results, "，"))
	fmt.Println("    6 个并发请求，只有 3 个成功，没有超发")
}
```

### 解析

「读当前值 → 判断 → 写回」这个序列必须整体原子，否则 6 个请求可能同时读到 3，
然后都扣减成功，超发到 -3。

CAS（Compare-And-Swap）的语义是：**只有当值仍然等于我读到的那个，才写入新值**。
失败了说明期间被别人改过，于是重读、重算、重试。整个模式就是一个乐观锁循环：

```go
for {
	cur := quota.Load()                          // 读快照
	if cur <= 0 { ... }                          // 基于快照做判断
	if quota.CompareAndSwap(cur, cur-1) { ... }  // 提交；失败就重来
}
```

它适合「单变量 + 冲突可接受重试」的场景。冲突非常激烈时，
大量 goroutine 会一直空转重试，反而比加锁更慢，此时应该回到 `Mutex`。

## 五、sync.Once：只做一次，且并发安全

### 代码

```go
type config struct{ name string }

var (
	configOnce sync.Once
	configVal  *config
)

// loadConfig 无论被多少个 goroutine 同时调用，初始化只执行一次。
func loadConfig() *config {
	configOnce.Do(func() {
		time.Sleep(20 * time.Millisecond) // 模拟昂贵初始化：读配置、建连接池
		configVal = &config{name: "prod"}
	})
	return configVal
}

func demoOnce() {
	fmt.Println("\n[5] sync.Once：并发调用也只执行一次")

	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = loadConfig()
		}()
	}
	wg.Wait()

	fmt.Printf("    5 个 goroutine 并发获取，总耗时 %v（≈ 单次初始化耗时）\n",
		time.Since(start).Round(5*time.Millisecond))
	fmt.Println("    其余 goroutine 会阻塞在 Once.Do 上，等初始化完成后拿到同一个指针")
}
```

### 解析

5 个 goroutine 并发调用 `loadConfig`，总耗时约等于一次初始化（20ms），
说明其余 4 个都阻塞在 `Once.Do` 上搭了顺风车。

三个使用要点：

1. **`Do` 内部 panic 会被记住。** 之后每次调用都会重新 panic，
   所以别把可能失败的操作放进去却不做处理。
2. **`Once` 不返回结果。** 需要「只执行一次并返回 (值, error)」时用
   `sync.OnceValues` / `sync.OnceFunc`（Go 1.21+）。
3. **`Once` 保护的是「执行一次」，不是数据安全。** 初始化完成后，
   `configVal` 仍可能被其它代码修改——如果它会变，就该改用锁或 `atomic.Pointer`。

## 六、sync.Pool：复用临时对象

### 代码

```go
func demoPool() {
	fmt.Println("\n[6] sync.Pool：复用临时对象，降低 GC 压力")

	pool := &sync.Pool{New: func() any { return make([]byte, 0, 4096) }}

	buf := pool.Get().([]byte)
	buf = append(buf, "hello"...)
	fmt.Printf("    取出的缓冲: %q（cap = %d）\n", buf, cap(buf))
	pool.Put(buf[:0]) // 归还前把长度清零，避免把内容带给下一个使用者

	buf2 := pool.Get().([]byte)
	fmt.Printf("    再次取出: len = %d，cap = %d（内容已清理）\n", len(buf2), cap(buf2))
	fmt.Println("    注意：池里的对象随时可能被 GC 回收，不能用来保存状态")
}
```

### 解析

`sync.Pool` 适合高频创建又高频丢弃的临时对象：`[]byte` 缓冲、`bytes.Buffer`、
JSON 编码器、protobuf 消息。它降低的是**分配次数和 GC 压力**。

必须记住的两条：

1. **池里的对象随时可能被 GC 清空。** 它不保证「放进去还能取出来」，
   所以不能用来保存状态、缓存结果、复用长连接。
2. **归还前要清理。** 示例里用 `buf[:0]` 把长度归零，避免把上一次的内容
   （可能是别人的数据）带给下一个使用者。

## 七、分片锁：用空间换竞争

### 代码

```go
// shardedCounter 把一把大锁拆成 N 把小锁，用空间换竞争。
type shardedCounter struct {
	shards [16]struct {
		mu sync.Mutex
		n  int64
	}
}

func (c *shardedCounter) Inc(key int) {
	s := &c.shards[uint64(key)%uint64(len(c.shards))]
	s.mu.Lock()
	s.n++
	s.mu.Unlock()
}

func (c *shardedCounter) Total() int64 {
	var total int64
	for i := range c.shards {
		s := &c.shards[i]
		s.mu.Lock()
		total += s.n
		s.mu.Unlock()
	}
	return total
}
```

```go
func demoSharded() {
	fmt.Println("\n[7] 分片锁：用空间换竞争")

	var c shardedCounter
	var wg sync.WaitGroup

	start := time.Now()
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				c.Inc(id) // 不同 goroutine 落在不同分片，冲突大幅减少
			}
		}(i)
	}
	wg.Wait()

	fmt.Printf("    结果 = %d，期望 = %d，耗时 %v\n",
		c.Total(), int64(workers*perWorker), time.Since(start).Round(time.Millisecond))
}
```

### 解析

单把锁的本质瓶颈是「串行化」：8 个 goroutine 抢一把锁，无论多少核，
临界区都是排队执行的。分片把一把锁拆成 16 把，不同 key 落到不同分片，
冲突概率大幅下降。

这是很通用的优化手段，典型应用：分片计数器、分片连接池、分片缓存
（第 8 篇的分片 benchmark 就是它）。

代价也要说清楚：

- `Total()` 必须遍历所有分片，读汇总比单锁慢
- 内存占用变大，且热点 key 仍然会集中在同一个分片

所以它适合「写多、读汇总少、key 分散」的场景。

## 性能对比（本机实测）

```text
cpu: Apple M5 Pro
BenchmarkCounterMutex-18      	14578431	        75.08 ns/op	       0 B/op	       0 allocs/op
BenchmarkCounterAtomic-18     	94703822	        12.61 ns/op	       0 B/op	       0 allocs/op
BenchmarkCounterChannel-18    	12087625	        96.97 ns/op	       0 B/op	       0 allocs/op
BenchmarkCounterSharded-18    	38231577	        32.54 ns/op	       0 B/op	       0 allocs/op
```

对应到「8 个 goroutine 抢着累加一个计数器」这个场景：

| 方式 | 每次操作 | 相对单锁 |
|---|---|---|
| `atomic.Int64` | 12.6 ns | 快约 6 倍 |
| 分片锁（16 片） | 32.5 ns | 快约 2.3 倍 |
| 单把 `Mutex` | 75.1 ns | 基准 |
| 容量 1 的 channel | 97.0 ns | 慢约 30% |

**这些数字是用来建立量级感的，不是用来选型的。** 如果一件操作每秒只发生几百次，
75ns 和 13ns 的差别毫无意义，此时应该选最容易读、最不容易写错的那个。
只有热点路径上的竞争才值得优化，而且要先用 benchmark 或 mutex profile 证明它确实是热点。

## 把所有演示串起来

```go
func main() {
	demoMutex()
	demoRWMutex()
	demoAtomic()
	demoCAS()
	demoOnce()
	demoPool()
	demoSharded()
}
```

## 运行输出

```text
[1] Mutex：保护临界区
    结果 = 400000，期望 = 400000，正确 = true，耗时 51ms

[2] RWMutex：读多写少
    读到的值: goroutine
    写锁要等所有读锁释放；读锁里做耗时操作会拖慢写，写非常频繁时 RWMutex 可能比 Mutex 更慢

[3] atomic：单个变量的无锁更新
    结果 = 400000，期望 = 400000，正确 = true

[4] CAS：读-改-写必须原子完成的场景
    请求1 被拒，请求2 被拒，请求3 拿到名额(剩余0)，请求4 拿到名额(剩余1)，请求5 被拒，请求6 拿到名额(剩余2)
    6 个并发请求，只有 3 个成功，没有超发

[5] sync.Once：并发调用也只执行一次
    5 个 goroutine 并发获取，总耗时 20ms（≈ 单次初始化耗时）

[6] sync.Pool：复用临时对象，降低 GC 压力
    取出的缓冲: "hello"（cap = 4096）
    再次取出: len = 0，cap = 4096（内容已清理）

[7] 分片锁：用空间换竞争
    结果 = 400000，期望 = 400000，耗时 17ms
```

## 常见错误

### 复制含锁的结构体

这是最容易犯又最隐蔽的错误——锁被复制成两份，保护立刻失效：

```go
type Counter struct {
	mu sync.Mutex
	n  int
}

func (c Counter) Inc() { ... }   // 反例：值接收者，复制了锁
func (c *Counter) Inc() { ... }  // 正确：指针接收者
```

`go vet` 的 `copylocks` 检查能抓到这种情况，所以 `go vet ./...` 应该进入日常流程。

### 其它高频问题

| 写法 | 后果 |
|---|---|
| `RLock` 里做耗时操作 | 写锁被长时间饿死 |
| `atomic` 和普通读写混用 | 竞争依旧存在，race detector 会报 |
| 忘记 `Unlock` / `RUnlock` | 死锁；用 `defer` 就不会忘 |
| 用两个 atomic 变量维护「必须一致」的状态 | 观察者能看到中间状态 |
| `Pool` 里存状态 | 对象随时被 GC 回收，状态丢失 |

## 练习

1. 实现并发安全的 LRU 缓存：`map` + `container/list` + `sync.Mutex`，
   `Get` 命中时把元素移到链表头部。
2. 把 LRU 改造成 16 分片版本，和单锁版本做 benchmark 对比。
3. 用 `atomic.Pointer[Config]` 实现热更新配置：写方整体替换指针，读方永远读到完整配置。

## 小结

- `Mutex` 保护临界区，粒度要小且要覆盖完整。
- `RWMutex` 适合读多写少，写多时不划算。
- `atomic` 只适合单变量，`CAS` 是它的读-改-写版本。
- `Once` 保证只执行一次，`Pool` 只用来减少分配。
- 分片锁用空间换竞争，适合热点写入。

下一篇把这些原语组装成可以复用的结构：五种并发模式。
