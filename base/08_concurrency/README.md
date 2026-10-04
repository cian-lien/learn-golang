# Go 并发编程教程

九篇独立教程文章，每篇都自带完整代码示例、真实运行输出和逐段解析。
第 0 篇从零基础讲起，不需要任何并发经验；文章正文在 `tutorials/`，
可运行的代码在 `examples/`，两者一一对应。

## 文章目录

| # | 文章 | 配套代码 | 你会得到什么 |
|---|---|---|---|
| 0 | [从零开始：什么是并发](tutorials/00-zero-basics.md) | `examples/00_hello_concurrency` | 只会一点 Go 语法也能读懂：串行 vs 并发、`go` 关键字、为什么必须等 goroutine |
| 1 | [goroutine 与 WaitGroup：并发的地基](tutorials/01-goroutine.md) | `examples/01_goroutine` | 搞清 `go` 做了什么、谁负责等待、main 退出时会发生什么 |
| 2 | [channel 的五条规则](tutorials/02-channel.md) | `examples/02_channel` | 握手、队列、广播、所有权、nil channel |
| 3 | [select、超时与取消](tutorials/03-select-timeout.md) | `examples/03_select_timeout` | 让每个等待都有退出路径，写出 `orDone` |
| 4 | [锁、原子操作与其它同步原语](tutorials/04-sync-atomic.md) | `examples/04_sync_atomic` | Mutex / RWMutex / atomic / CAS / Once / Pool / 分片锁 |
| 5 | [五种并发模式](tutorials/05-patterns.md) | `examples/05_patterns` | worker pool、pipeline、fan-out/fan-in、限流、singleflight |
| 6 | [context、errgroup 与优雅退出](tutorials/06-context-errgroup.md) | `examples/06_context` | 取消传播、超时判断、errgroup、HTTP 优雅停机 |
| 7 | [七个最常见的并发陷阱](tutorials/07-pitfalls.md) | `examples/07_pitfalls` | 泄漏、死锁、数据竞争、并发写 map 的排查与预防 |
| 8 | [诊断与调优工具箱](tutorials/08-diagnostics.md) | `examples/08_diagnostics` | race、pprof、trace、benchmark 四件套 |

建议按顺序阅读：第 0 篇零基础入门，之后每一篇都会引用前面的约定，
比如「谁发送谁关闭」「Add 在 go 之前」。

## 怎么用

所有命令都在 `go/` 目录下执行（`go.mod` 所在目录）。

```bash
cd go

# 跟随某一篇文章跑对应示例
go run ./base/08_concurrency/examples/01_goroutine

# 并发代码建议一律带上 race detector
go run -race ./base/08_concurrency/examples/04_sync_atomic

# 跑测试（第 7 篇带一个 goroutine 泄漏断言）
go test -race ./base/08_concurrency/...

# 跑基准测试（第 4、8 篇的性能对比数据来自这里）
go test -bench . -benchmem ./base/08_concurrency/examples/08_diagnostics

# 故意写错的反例，需要单独指定文件运行
go run -race ./base/08_concurrency/examples/07_pitfalls/broken/race_counter.go
go run ./base/08_concurrency/examples/07_pitfalls/broken/deadlock.go
go run ./base/08_concurrency/examples/07_pitfalls/broken/map_write.go
```

`broken/` 下的文件带 `//go:build ignore`，不会参与 `go build ./...` 和 `go vet ./...`，
所以放心跑。

## 先建立三个直觉

**一、goroutine 不是线程。** 它由 Go 运行时调度，初始栈只有 2KB 左右，可以按需增长。
创建和切换的开销远小于系统线程，所以「一个连接一个 goroutine」在 Go 里是合理设计。
真正限制你的是内存和下游服务的承受能力，不是 goroutine 的数量。

**二、并发不等于并行。** 并发是「同时处理多件事」的结构，并行是「同一时刻真的在多个核上跑」。
`GOMAXPROCS` 决定并行度，但几万个 goroutine 完全可以只靠几个线程轮流跑——
因为绝大多数 goroutine 都在等待 IO，而不是在计算。

**三、Go 的选择顺序是：能不用共享就不用共享，能用简单原语就别上 channel。**

| 需求 | 首选 |
|---|---|
| 一个计数器、一个开关标志 | `sync/atomic` |
| 一小段代码读写共享数据结构 | `sync.Mutex` |
| 读多写少 | `sync.RWMutex` |
| 在 goroutine 之间传递数据、传递所有权 | `channel` |
| 只初始化一次 | `sync.Once` |
| 复用临时对象、降低 GC 压力 | `sync.Pool` |
| 取消、超时、跨层传递请求级数据 | `context` |

channel 不是「更高级的锁」。用 channel 实现计数器是典型的过度设计，
实测数据（`examples/08_diagnostics` 的 benchmark）：

| 实现方式 | 每次操作耗时 |
|---|---|
| `atomic.Int64` | 12.6 ns |
| 分片锁（16 片） | 32.5 ns |
| 单把 `sync.Mutex` | 75.1 ns |
| 容量 1 的 channel 当锁 | 97.0 ns |

## 综合练习

读完八篇之后，用这四道题把知识串起来。每题都给出了验收标准。

**1. 并发爬虫**

输入一组 URL，最多 8 个并发，支持 Ctrl-C 取消，失败的 URL 不影响其它 URL，
最后输出成功/失败清单和耗时。

验收：取消后 1 秒内进程退出（说明没有泄漏）；`go run -race` 干净。

**2. 有界任务队列**

生产者可以持续提交，队列有上限，满了生产者阻塞（背压）。
支持优雅关闭：关闭后处理完在途任务才退出。

验收：用 `runtime.NumGoroutine()` 验证关闭后没有多余 goroutine；
写测试覆盖「关闭时队列非空」的情况。

**3. 并发安全缓存**

带 TTL、容量上限、LRU 淘汰，并对同一个 key 的并发回源用 singleflight 合并。

验收：写一个并发测试，证明 100 个 goroutine 同时读同一个 key 时回源只发生一次。

**4. mini errgroup**

实现 `Group`：`Go(f func() error)`、`Wait() error`、`SetLimit(n)`、
只保留第一个错误、任一失败取消其余任务。

验收：测试覆盖「全部成功」「第一个失败后其余被取消」「并发上限生效」三种情况，
并用 `-race` 运行。

## 自测清单

不看资料能答出下面这些问题，说明内容已经掌握：

1. 无缓冲 channel 和有缓冲 channel 的阻塞点分别在哪？
2. `close(ch)` 之后，`range ch` 和 `v, ok := <-ch` 分别是什么行为？
3. 为什么 `wg.Add(1)` 不能写在 goroutine 内部？
4. 什么情况下应该用 `atomic` 而不是 `Mutex`？
5. `RWMutex` 在什么场景下反而比 `Mutex` 慢？
6. `select` 有多个 case 就绪时如何选择？加 `default` 改变了什么？
7. 循环里的 `time.After` 泄漏的是什么资源？
8. `ctx.Err()` 返回 `Canceled` 和 `DeadlineExceeded` 分别意味着什么？
9. 判断一个 goroutine 会不会泄漏，你要问自己哪个问题？
10. `-race`、pprof、trace 各解决什么问题？

## 速查表

| 你要做的事 | 用什么 |
|---|---|
| 等一组 goroutine 结束 | `sync.WaitGroup`（Add 在 go 之前） |
| 一组任务，任一失败就取消 | `errgroup.WithContext` + `SetLimit` |
| 限制同时执行的数量 | 带缓冲 channel 当信号量 |
| 限制每秒发起的次数 | `time.Ticker` 或 `golang.org/x/time/rate` |
| 给等待加一个上限 | `context.WithTimeout` / `select + time.After` |
| 让调用方可以取消 | `context` 一路传下去 |
| 只初始化一次 | `sync.Once` / `sync.OnceValues` |
| 保护一个计数器 | `sync/atomic` |
| 保护一段共享数据 | `sync.Mutex`（读多写少用 `RWMutex`） |
| 高并发写入的计数器/指标 | 分片锁 |
| 减少重复分配 | `sync.Pool` |
| 合并同一资源的并发请求 | `golang.org/x/sync/singleflight` |
| 优雅退出 | `signal.NotifyContext` + `Server.Shutdown` |
| 检查数据竞争 | `go test -race ./...` |
| 排查 goroutine 泄漏 | goroutine profile + `goleak` |
| 定位锁竞争 | mutex profile + 分片 |
| 定位延迟毛刺 | `runtime/trace` |

## 参考资料

- 《The Go Programming Language》（《Go 语言圣经》）第 8、9 章：并发原语与传统模式的对照
- 《Concurrency in Go》（《Go 语言并发之道》）：`orDone`、`tee`、`bridge` 等组合子的出处
- 《Learning Go》（《Go 语言学习指南》）第 11 章：工程实践视角
- Go 官方博客：*Share Memory By Communicating*、*Go Concurrency Patterns*、
  *Advanced Go Concurrency Patterns*、*The Go Memory Model*
- [Go 并发缺陷检测（race detector）](https://go.dev/doc/articles/race_detector)
- [Diagnostics 官方总览](https://go.dev/doc/diagnostics)：pprof、trace、godebug
- 源码阅读：`sync` 包、`golang.org/x/sync/errgroup`、`golang.org/x/sync/singleflight`
