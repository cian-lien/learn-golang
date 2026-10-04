# select、超时与取消

并发代码里最危险的一句话是「我在这里等它回来」。只要对方不回来，
goroutine 就永远卡住——它不会超时，也不会报错，只是安静地占着内存。

这篇的目标是让**每一个等待都有退出路径**，也是下一篇 `context` 的铺垫。

**配套代码**：`examples/03_select_timeout/main.go`

```bash
cd go
go run ./base/08_concurrency/examples/03_select_timeout
```

---

## 一、select：同时等待多个 channel

### 代码

```go
// demoSelectBasics 演示 select 最基本的行为：谁先就绪就执行谁。
func demoSelectBasics() {
	fmt.Println("[1] select：同时等待多个 channel")
	fast := make(chan string)
	slow := make(chan string)

	go func() { time.Sleep(20 * time.Millisecond); fast <- "fast 的结果" }()
	go func() { time.Sleep(150 * time.Millisecond); slow <- "slow 的结果" }()

	for i := 0; i < 2; i++ {
		select {
		case v := <-fast:
			fmt.Println("    收到:", v)
		case v := <-slow:
			fmt.Println("    收到:", v)
		}
	}
	fmt.Println("    第二次 select 会一直阻塞，直到 slow 也把结果送过来")
}
```

### 解析

`select` 是「多路等待」：所有 case 都没就绪时它会阻塞，谁先就绪就执行谁。
它不是一个轮询循环，不会消耗 CPU。

上例第一次循环 20ms 后拿到 fast 的结果，第二次就要一直等到 150ms 的 slow 返回。
这解释了它的常见用途：

- 同时等多个数据源，谁先回来用谁（例如同时查缓存和数据库）
- 等待结果，同时监听取消信号

## 二、多个 case 同时就绪时，随机选择

### 代码

```go
// demoSelectRandom 演示多个 case 同时就绪时 select 的随机性。
func demoSelectRandom() {
	fmt.Println("\n[2] 多个 case 同时就绪：随机选一个")

	a := make(chan int, 1)
	b := make(chan int, 1)
	counts := map[string]int{}

	for i := 0; i < 200; i++ {
		a <- 1
		b <- 1
		select {
		case <-a:
			counts["a"]++
		case <-b:
			counts["b"]++
		}
		// 清空残留，保证下一轮两个 case 都是就绪状态
		for len(a) > 0 {
			<-a
		}
		for len(b) > 0 {
			<-b
		}
	}

	fmt.Printf("    200 次里 a=%d、b=%d\n", counts["a"], counts["b"])
	fmt.Println("    随机是为了避免某个分支长期被饿死，所以不要依赖 select 的顺序")
}
```

### 解析

输出大概是 `a=95、b=105`，每次运行都不一样，但长期接近五五开。

Go 故意把选择做成随机的，目的是**避免饥饿**：如果固定按书写顺序优先，
排在后面的分支在高压下可能永远拿不到机会。

实践结论：**永远不要依赖 select 的先后顺序**。如果你需要优先级，
必须显式写成「先探测高优先级，再进入无优先级的 select」。

## 三、超时：给「等不到结果」一个出口

### 代码

```go
// demoTimeout 演示超时退出，以及超时之后上游 goroutine 的处境。
func demoTimeout() {
	fmt.Println("\n[3] 超时：给「等不到结果」一个出口")

	// 关键：缓冲为 1。即使调用方超时走了，发送方也不会永久阻塞。
	slow := make(chan string, 1)
	go func() {
		time.Sleep(200 * time.Millisecond)
		slow <- "迟到的结果"
	}()

	select {
	case v := <-slow:
		fmt.Println("    正常收到:", v)
	case <-time.After(50 * time.Millisecond):
		fmt.Println("    50ms 内没有结果 → 超时返回")
	}

	time.Sleep(200 * time.Millisecond)
	fmt.Println("    超时只是「调用方不等了」，goroutine 还在跑；")
	fmt.Println("    如果它往无缓冲 channel 发送，那次发送会永久阻塞 —— 这就是典型的 goroutine 泄漏")
}
```

### 解析

这段代码里最值钱的是注释提到的那一点：**`make(chan string, 1)` 的 1**。

如果这里写成无缓冲 channel，超时 50ms 后调用方就走了；150ms 后那个 goroutine
执行 `slow <- "迟到的结果"` 时没有接收者，于是永久阻塞。它不会 panic，不会报错，
只是从此占着栈内存活着——这就是一个 goroutine 泄漏。

给 channel 加一个缓冲位，等于给迟到者留了张桌子，让它能放下结果然后正常退出。

另一个要注意的是 `time.After` 本身：它在超时触发前不会被回收，
所以**不要放在循环里反复调用**（见下一节）。

## 四、定时器复用：循环里的正确写法

### 代码

```go
// waitIdle 反复读取事件，超过 idle 没有新事件就退出。
// 全程只用一个定时器：循环里反复 time.After 会不断产生新的定时器。
func waitIdle(ch <-chan int, idle time.Duration) {
	timer := time.NewTimer(idle)
	defer timer.Stop() // 一定要 Stop，避免未触发的定时器一直占着资源

	for {
		select {
		case v, ok := <-ch:
			if !ok {
				fmt.Println("    上游关闭，退出")
				return
			}
			fmt.Println("    收到事件:", v)
			if !timer.Stop() { // 重置前先停掉并排空，避免读到过期事件
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(idle)
		case <-timer.C:
			fmt.Printf("    %v 没有新事件，空闲退出\n", idle)
			return
		}
	}
}
```

配套的演示：每 40ms 产生一个事件，idle 设为 100ms，所以三个事件之后空闲退出。

```go
func demoReusableTimer() {
	fmt.Println("\n[4] 定时器复用")

	events := make(chan int, 8)
	go func() {
		for i := 1; i <= 3; i++ {
			time.Sleep(40 * time.Millisecond)
			events <- i
		}
		// 故意不 close：让 waitIdle 走「空闲超时」分支
	}()

	waitIdle(events, 100*time.Millisecond)

	fmt.Println("    反例：for { select { case <-ch: case <-time.After(5 * time.Second): } }")
	fmt.Println("    每轮都新建定时器，5 秒内循环 1000 次就会留下最多 1000 个尚未触发的定时器")
}
```

### 解析

`waitIdle` 是「空闲超时」的通用模板，常用于：连接保活、批量聚合窗口、
心跳检测、消费端自动退出。三个细节值得记住：

1. `defer timer.Stop()` 必须写。未触发的定时器在 Go 1.23 之前不会被 GC 回收。
2. 重置之前先 `Stop()` 并**排空 channel**，否则可能读到上一次已经过期的信号，
   导致「提前超时」这种很难查的 bug。
3. `case v, ok := <-ch` 里的 `ok` 判断不能省，否则上游关闭时会一直读到零值，
   变成死循环。

## 五、default：非阻塞操作与它的陷阱

### 代码

```go
// demoNonBlocking 演示 default 分支带来的非阻塞语义。
func demoNonBlocking() {
	fmt.Println("\n[5] default：非阻塞操作")

	ch := make(chan int, 1)

	select {
	case v := <-ch:
		fmt.Println("    读到", v)
	default:
		fmt.Println("    通道里没有数据 → 不等待，立刻执行 default")
	}

	ch <- 7
	select {
	case v := <-ch:
		fmt.Println("    读到", v)
	default:
		fmt.Println("    通道里没有数据")
	}

	fmt.Println("    警告：在 for 循环里用 default 会退化成 100% CPU 的忙轮询")
}
```

### 解析

`default` 把 select 从「阻塞等待」变成「试一次就走」，适合：

- 非阻塞发送：缓冲满了就丢弃（metrics、日志）
- 非阻塞接收：批量处理时先把已到达的数据全部取走（第 2 篇的 `drain` 就是这么写的）

但如果写成下面这样，CPU 会直接跑满：

```go
// 反例：空忙轮询
for {
	select {
	case v := <-ch:
		handle(v)
	default:
	}
}
```

需要「等待」的时候就不该有 `default`；需要「不空转地等一小会儿」，
正确做法是加 `time.After`，而不是空转。

## 六、取消模式：done channel 与 orDone

### 代码

```go
// orDone 把「可取消」包装进一个只读 channel。
// 无论上游被取消，还是消费者提前退出，它都不会留下阻塞的 goroutine。
func orDone(done <-chan struct{}, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- v:
				case <-done: // 下游已经走了，不要在这里死等
					return
				}
			}
		}
	}()
	return out
}

// countUp 每 interval 产生一个递增数字，直到 done 被关闭。
// 注意两次 select：等待节拍时会响应取消，发送时也会响应取消。
func countUp(done <-chan struct{}, interval time.Duration) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		i := 0
		for {
			select {
			case <-done:
				fmt.Println("    上游 goroutine 收到取消，已退出")
				return
			case <-ticker.C:
				i++
				select {
				case out <- i:
				case <-done:
					fmt.Println("    上游 goroutine 在发送时收到取消，已退出")
					return
				}
			}
		}
	}()
	return out
}

// demoOrDone 演示 done channel 的取消传播：消费者收工，上游立刻全部退出。
func demoOrDone() {
	fmt.Println("\n[6] 取消模式：done channel + orDone")

	done := make(chan struct{})
	numbers := orDone(done, countUp(done, 30*time.Millisecond))

	count := 0
	for v := range numbers {
		fmt.Println("    消费:", v)
		count++
		if count == 3 {
			fmt.Println("    消费者决定收工 → close(done)")
			close(done)
		}
	}

	time.Sleep(20 * time.Millisecond) // 等上游打印退出日志，仅为了让输出整齐
	fmt.Println("    range 正常结束：done 关闭后，上游所有 goroutine 都已退出")
}
```

### 解析

`orDone` 只有二十几行，却是这套教程里最值得抄走的一段代码。它解决两个方向的泄漏：

1. **消费者先走**：`out <- v` 会阻塞在无人接收的 channel 上。
   第二个 `select` 里的 `case <-done` 让发送方有机会放弃。
2. **上游先结束**：`in` 关闭时 `ok == false`，`return` 并关闭 `out`，
   下游的 `range` 才会正常结束。

少了任何一层，都会有 goroutine 卡住。你可以把第二个 select 删掉跑一次，
观察 `countUp` 是否还能打印退出日志——这是最直观的验证方法。

另外注意 `countUp` 里的 ticker 用了 `defer ticker.Stop()`：
即使是被取消退出，定时器也会被释放。

## 把所有演示串起来

```go
func main() {
	demoSelectBasics()
	demoSelectRandom()
	demoTimeout()
	demoReusableTimer()
	demoNonBlocking()
	demoOrDone()
	fmt.Println("\n[7] select{} 不带任何 case 会永久阻塞（适合让 main 不退出，不要在需要退出的地方写）")
}
```

## 运行输出

```text
[1] select：同时等待多个 channel
    收到: fast 的结果
    收到: slow 的结果
    第二次 select 会一直阻塞，直到 slow 也把结果送过来

[2] 多个 case 同时就绪：随机选一个
    200 次里 a=95、b=105
    随机是为了避免某个分支长期被饿死，所以不要依赖 select 的顺序

[3] 超时：给「等不到结果」一个出口
    50ms 内没有结果 → 超时返回
    超时只是「调用方不等了」，goroutine 还在跑；
    如果它往无缓冲 channel 发送，那次发送会永久阻塞 —— 这就是典型的 goroutine 泄漏

[4] 定时器复用
    收到事件: 1
    收到事件: 2
    收到事件: 3
    100ms 没有新事件，空闲退出
    反例：for { select { case <-ch: case <-time.After(5 * time.Second): } }
    每轮都新建定时器，5 秒内循环 1000 次就会留下最多 1000 个尚未触发的定时器

[5] default：非阻塞操作
    通道里没有数据 → 不等待，立刻执行 default
    读到 7
    警告：在 for 循环里用 default 会退化成 100% CPU 的忙轮询

[6] 取消模式：done channel + orDone
    消费: 1
    消费: 2
    消费: 3
    消费者决定收工 → close(done)
    上游 goroutine 收到取消，已退出
    range 正常结束：done 关闭后，上游所有 goroutine 都已退出

[7] select{} 不带任何 case 会永久阻塞（适合让 main 不退出，不要在需要退出的地方写）
```

## 常见错误

| 写法 | 问题 |
|---|---|
| 超时后用无缓冲 channel 接收结果 | 上游 goroutine 永久阻塞（泄漏） |
| 循环里 `time.After` | 累积大量未触发的定时器 |
| `Reset` 前不 `Stop` + 排空 | 读到过期信号，提前超时 |
| 在 `for` 里用 `default` | 100% CPU 忙轮询 |
| 依赖 select 分支顺序 | 顺序是随机的 |
| `select {}` 写在需要退出的路径上 | 永远阻塞（它确实是合法写法，但只用于让 main 挂着） |

## 练习

1. 实现 `waitFor(ch <-chan int, timeout time.Duration) (int, error)`：
   拿到值就返回，超时返回 `context.DeadlineExceeded`。
2. 把 `orDone` 里第二个 `select` 删掉，运行并解释观察到的现象。
3. 用 `waitIdle` 的思路实现「批量聚合」：收集 100 个事件或空闲 50ms 就立即处理一批。

## 小结

- `select` 多路等待，多个就绪时随机选择。
- `default` 让它变成非阻塞，但不要在循环里空转。
- 超时是给调用方的退路，上游还得靠取消信号才能真正退出。
- 循环里复用定时器，不要反复 `time.After`。
- `orDone` 是处理「取消 + 发送」的标准写法。

下一篇进入共享内存：锁与原子操作该怎么选。
