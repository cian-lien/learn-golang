# channel 的五条规则

goroutine 解决了「怎么同时干活」，这篇解决「它们之间怎么交换数据」。
channel 的规则不多，但每一条都有对应的坑，所以这篇把语义讲透，再对比「什么时候不该用 channel」。

**配套代码**：`examples/02_channel/main.go`

```bash
cd go
go run ./base/08_concurrency/examples/02_channel
```

---

## 规则一：无缓冲 channel 是一次「握手」

### 代码

```go
// demoUnbuffered 演示无缓冲 channel 的「同步握手」语义。
func demoUnbuffered() {
	fmt.Println("[1] 无缓冲 channel：发送和接收必须碰面")
	ch := make(chan string)

	go func() {
		fmt.Println("    sender : 准备发送（此刻会阻塞，直到有人接收）")
		ch <- "ping"
		fmt.Println("    sender : 发送完成（说明接收方已经就位）")
	}()

	time.Sleep(50 * time.Millisecond) // 让 sender 先跑起来，方便观察顺序
	fmt.Println("    main   : 现在开始接收")
	fmt.Println("    main   : 收到", <-ch)
	time.Sleep(20 * time.Millisecond) // 等 sender 打印完最后一行
}
```

### 解析

运行输出是：

```text
    sender : 准备发送（此刻会阻塞，直到有人接收）
    main   : 现在开始接收
    main   : 收到 ping
    sender : 发送完成（说明接收方已经就位）
```

注意 `sender : 发送完成` 出现在 `main : 收到 ping` **之后**——`ch <- "ping"` 这一行
阻塞了 50ms 以上，直到接收方真的就位才继续。

所以无缓冲 channel 传递的不只是数据，还有「两个 goroutine 此刻交会」这个事实。
它天然适合做同步点：等待某个事件发生、等待一个结果就绪。

## 规则二：有缓冲 channel 是队列，容量就是水位线

### 代码

```go
// demoBuffered 演示缓冲区容量对阻塞点的影响。
func demoBuffered() {
	fmt.Println("\n[2] 有缓冲 channel：容量内发送不阻塞")
	ch := make(chan int, 2)
	ch <- 1
	ch <- 2
	fmt.Printf("    len=%d cap=%d：两个元素已经入队，发送方没有被阻塞\n", len(ch), cap(ch))

	// 用 select + default 探测「这一次发送会不会阻塞」，避免程序真的卡住
	select {
	case ch <- 3:
		fmt.Println("    第 3 次发送成功")
	default:
		fmt.Println("    缓冲已满：第 3 次发送会阻塞，被 default 挡下")
	}

	fmt.Println("    收走一个元素:", <-ch)
	select {
	case ch <- 3:
		fmt.Println("    腾出空间后，第 3 次发送成功")
	default:
		fmt.Println("    缓冲仍然是满的")
	}
	fmt.Println("    缓冲区剩余内容:", drain(ch))
}

// drain 读出 channel 中当前所有数据（不等待新的数据）。
func drain(ch <-chan int) []int {
	out := []int{}
	for {
		select {
		case v := <-ch:
			out = append(out, v)
		default:
			return out
		}
	}
}
```

### 解析

- `len(ch)` 是当前积压量，`cap(ch)` 是上限。`len=2 cap=2` 表示队列满了。
- 满了之后再发送会阻塞，直到有人取走一个。
- `select + default` 是**非阻塞探测**：有空间就发，没空间就走 `default`。
  这在调试和「丢旧数据」策略里很有用，但不要放在循环里无脑旋转。

`drain` 使用 `<-chan int`（只读 channel）作为参数类型——这是规则四的伏笔。

## 规则三：close 是广播，不是清空

### 代码

```go
// demoClose 演示 close 的三种效果：range 结束、零值 + ok、广播唤醒。
func demoClose() {
	fmt.Println("\n[3] close：关闭是广播，不是清空")

	ch := make(chan int, 3)
	for i := 1; i <= 3; i++ {
		ch <- i
	}
	close(ch)

	fmt.Print("    关闭后 range 仍能读出剩余数据:")
	for v := range ch { // 缓冲耗尽且已关闭时，range 自动结束
		fmt.Print(" ", v)
	}
	fmt.Println("  → range 正常结束")

	v, ok := <-ch
	fmt.Printf("    再接收一次: v=%d ok=%v（ok=false 表示 channel 已关闭且没有数据）\n", v, ok)

	// close 是一次广播：所有等待者同时被唤醒
	done := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-done // 三个 goroutine 一起等
		}()
	}
	time.Sleep(20 * time.Millisecond)
	close(done)
	wg.Wait()
	fmt.Println("    close(done) 一次性唤醒 3 个等待者（这就是取消信号的经典用法）")
}
```

### 解析

`close` 有三层含义，全部在输出里能看到：

| 行为 | 结果 |
|---|---|
| 关闭后继续 `range` | 缓冲里的数据照常读出，读空后循环结束 |
| 关闭后再接收 | 立刻返回零值，`ok == false` |
| 关闭时有人在等 | 所有等待者**同时**被唤醒 |

第三点最关键：`close` 是一次广播，而发送一个值只有一个接收者能拿到。
所以「通知所有 goroutine 停下来」必须用 `close(done)`，不能用 `done <- struct{}{}`。

顺带一提：`close` 不会清空缓冲区，也不会释放还没被读走的数据。

## 规则四：谁发送，谁关闭；只能关一次

### 代码

```go
// demoClosePanic 演示两种会 panic 的关闭/发送动作。
func demoClosePanic() {
	fmt.Println("\n[4] 关闭相关的两个 panic")

	ch := make(chan int, 1)
	ch <- 1
	close(ch)

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("    向已关闭的 channel 发送 →", r)
			}
		}()
		ch <- 2 // panic: send on closed channel
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("    重复 close →", r)
			}
		}()
		close(ch) // panic: close of closed channel
	}()

	fmt.Println("    结论：只有发送方能关闭自己拥有的 channel，且必须确保只关一次")
}
```

### 解析

两种 panic 都源于同一件事：**关闭状态是不可逆的，而且接收方无法判断「现在还能不能发」**。

所以社区形成了所有权约定：

```go
// produce 返回只读 channel。
//
// 约定：channel 的所有者是发送方，也由发送方负责 close。
// 返回 <-chan int 让调用方在编译期就无法往里面写。
func produce(nums []int) <-chan int {
	out := make(chan int, len(nums))
	go func() {
		defer close(out) // 只有发送方知道「没有更多数据了」
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}
```

注意函数签名：返回 `<-chan int`，参数里用 `chan<- int`。单向 channel 是编译期的约束，
比写在注释里的约定可靠得多。

调用方只负责读（`for v := range ch`），读到关闭自然结束：

```go
func demoDirectionAndOwnership() {
	fmt.Println("\n[6] 单向 channel 与所有权约定")

	ch := produce([]int{1, 2, 3, 4})
	sum := 0
	for v := range ch { // 调用方只管读，读到关闭自然结束
		sum += v
	}
	fmt.Println("    生产者 close、消费者 range:", sum)
}
```

## 规则五：nil channel 永远阻塞

### 代码

```go
// demoNilChannel 演示 nil channel 永远阻塞的特性。
func demoNilChannel() {
	fmt.Println("\n[5] nil channel：永远阻塞")

	var never chan int // 零值是 nil

	select {
	case v := <-never:
		fmt.Println("    永远不会执行", v)
	case <-time.After(30 * time.Millisecond):
		fmt.Println("    nil channel 的分支永不就绪，只能走超时分支")
	}

	fmt.Println("    实战技巧：在 select 里把暂时不关心的 channel 置为 nil，即可关闭该分支")
}
```

### 解析

channel 的零值是 nil，对 nil channel 的收发会**永久阻塞**（不是 panic）。
在 `select` 里，nil channel 对应的 case 永远不会被选中——这不是 bug，而是一个好用的开关：

```go
// 只关心 jobs，暂时不想处理 signals
signals := make(chan os.Signal)
select {
case job := <-jobs:
	handle(job)
case sig := <-signals:
	handleSignal(sig)
}

// 想临时禁用某个分支：
signals = nil
```

调试提示：如果 goroutine 的栈显示它卡在 `chan receive (nil chan)`，说明你忘了初始化 channel。

## 附：用 channel 做信号量

### 代码

```go
// demoSemaphore 用带缓冲 channel 当信号量，限制同时运行的任务数。
func demoSemaphore() {
	fmt.Println("\n[7] 用带缓冲 channel 做信号量：限制并发数")

	sem := make(chan struct{}, 2) // 最多 2 个任务同时在跑
	var wg sync.WaitGroup
	var mu sync.Mutex
	finished := []int{}

	start := time.Now()
	for i := 1; i <= 4; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}        // 取令牌：满了就在这里等
			defer func() { <-sem }() // 释放令牌
			time.Sleep(50 * time.Millisecond)
			mu.Lock()
			finished = append(finished, i)
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	fmt.Printf("    完成顺序: %v，总耗时 %v（4 个任务 × 50ms，限流 2 后约 100ms）\n",
		finished, time.Since(start).Round(10*time.Millisecond))
}
```

### 解析

带缓冲 channel 的容量天然表达「最多同时进行 N 个」：发送是取令牌，接收是还令牌。
这是 channel 最合适的用法之一。

注意 `finished` 这个共享切片需要 `mu` 保护——**信号量控制的是并发数量，
不负责数据安全**。两者是正交的。

## 把所有演示串起来

```go
func main() {
	demoUnbuffered()
	demoBuffered()
	demoClose()
	demoClosePanic()
	demoNilChannel()
	demoDirectionAndOwnership()
	demoSemaphore()
}
```

## 运行输出

```text
[1] 无缓冲 channel：发送和接收必须碰面
    sender : 准备发送（此刻会阻塞，直到有人接收）
    main   : 现在开始接收
    main   : 收到 ping
    sender : 发送完成（说明接收方已经就位）

[2] 有缓冲 channel：容量内发送不阻塞
    len=2 cap=2：两个元素已经入队，发送方没有被阻塞
    缓冲已满：第 3 次发送会阻塞，被 default 挡下
    收走一个元素: 1
    腾出空间后，第 3 次发送成功
    缓冲区剩余内容: [2 3]

[3] close：关闭是广播，不是清空
    关闭后 range 仍能读出剩余数据: 1 2 3  → range 正常结束
    再接收一次: v=0 ok=false（ok=false 表示 channel 已关闭且没有数据）
    close(done) 一次性唤醒 3 个等待者（这就是取消信号的经典用法）

[4] 关闭相关的两个 panic
    向已关闭的 channel 发送 → send on closed channel
    重复 close → close of closed channel
    结论：只有发送方能关闭自己拥有的 channel，且必须确保只关一次

[5] nil channel：永远阻塞
    nil channel 的分支永不就绪，只能走超时分支
    实战技巧：在 select 里把暂时不关心的 channel 置为 nil，即可关闭该分支

[6] 单向 channel 与所有权约定
    生产者 close、消费者 range: 10

[7] 用带缓冲 channel 做信号量：限制并发数
    完成顺序: [4 1 3 2]，总耗时 100ms（4 个任务 × 50ms，限流 2 后约 100ms）
```

## 什么时候不该用 channel

channel 不等于「更高级的锁」。判断标准：

| 你的需求 | 用什么 |
|---|---|
| 表达「同时最多 N 个」「传递数据所有权」「事件通知」 | channel |
| 保护一段临界区、几个字段必须一起更新 | `sync.Mutex` |
| 一个计数器的自增 | `sync/atomic` |

用 channel 实现计数器的开销大约是 `atomic` 的 8 倍、单把锁的 1.3 倍
（第 8 篇有实测数据），而且代码更难读。**能一句话说清用锁的地方，就用锁。**

## 练习

1. 实现一个有界任务队列：生产者可以关闭队列，消费者 `range` 到关闭后自然退出，不得 panic。
2. 为什么 `produce` 返回 `<-chan int` 而不是 `chan int`？把返回值改成 `chan int`，
   看看调用方引出了什么问题。
3. 用 `close(done)` 实现「同时通知 100 个 goroutine 退出」，再用 `runtime.NumGoroutine()`
   验证没有泄漏。

## 小结

1. 无缓冲 = 握手，有缓冲 = 队列。
2. `len`/`cap` 描述水位，满了就阻塞。
3. `close` 是广播，关闭后仍能读完剩余数据。
4. 谁发送谁关闭，只关一次；用单向 channel 把它写进类型。
5. nil channel 永远阻塞，在 select 里可以当开关。

下一篇解决「等待不能无限期」的问题：select、超时与取消。
