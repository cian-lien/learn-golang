# goroutine 与 WaitGroup：并发的地基

这是整套教程的第一篇。我们会用一个可运行的程序回答三个问题：
`go` 语句到底做了什么、谁来等 goroutine 结束、主 goroutine 退出时会发生什么。

**配套代码**：`examples/01_goroutine/main.go`

```bash
cd go
go run ./base/08_concurrency/examples/01_goroutine
```

---

## 一、先量化收益：串行 vs 并发

### 代码

```go
const jobDelay = 100 * time.Millisecond

// work 模拟一次耗时的外部调用：网络、磁盘、RPC 都可以抽象成这样。
func work(d time.Duration) { time.Sleep(d) }

// serial 串行执行 n 个任务。
func serial(n int) time.Duration {
	start := time.Now()
	for i := 1; i <= n; i++ {
		work(jobDelay)
	}
	return time.Since(start)
}

// sumOfSquares 作为并发结果的校验基准。
func sumOfSquares(n int) int {
	total := 0
	for i := 1; i <= n; i++ {
		total += i * i
	}
	return total
}
```z

### 解析

`work` 用 `time.Sleep` 模拟「等待」——网络请求、磁盘 IO、下游 RPC 本质上都是等待。
这类任务的特点是：**goroutine 在睡觉，不占用 CPU**。这正是并发最擅长的场景。

`sumOfSquares` 不参与计时，它的作用是当裁判：并发求和的结果必须和它一致，
否则说明我们写出了数据竞争。

## 二、用 WaitGroup 并发做完同一批任务

### 代码

```go
// parallel 并发执行 n 个任务，用 WaitGroup 等它们全部结束。
func parallel(n int) time.Duration {
	start := time.Now()

	var wg sync.WaitGroup
	squares := make([]int, n) // 预先分配，每个 goroutine 只写自己那一格

	for i := 1; i <= n; i++ {
		wg.Add(1) // 必须在 go 之前 +1，否则 Wait 可能提前返回
		go func(i int) {
			defer wg.Done() // 任何返回路径都会正确计数
			work(jobDelay)
			squares[i-1] = i * i // 写不同下标，不构成数据竞争
		}(i)
	}

	wg.Wait() // 阻塞到计数器归零

	total := 0
	for _, v := range squares {
		total += v
	}
	fmt.Printf("    并发求和 = %d，校验通过: %v\n", total, total == sumOfSquares(n))
	return time.Since(start)
}
```

### 解析

`sync.WaitGroup` 是一个计数器：`Add(1)` 加一，`Done()` 减一，`Wait()` 阻塞到归零。
三个使用要点：

1. **`Add` 必须在 `go` 之前调用。** 写在 goroutine 内部就晚了——`Wait()` 可能在计数增加前
   就返回，main 直接结束，任务半途而废且不会报错。
2. **`Done` 用 `defer` 保证执行。** 函数中途 `return` 或 panic 都能正确计数。
3. **不要依赖输出顺序。** 5 个 goroutine 的完成顺序每次都不一样，所以这里先把结果写进
   切片、等 `Wait()` 之后再统一处理。

`squares[i-1] = i * i` 是这段代码的关键设计：**每个 goroutine 只写自己那一格**，
彼此不重叠，因此不需要加锁。按索引切分数据是消除竞争最便宜的手段。

最后不要忘了 `parallel` 是在 `main` 里被调用的——`Wait()` 只能保证这 5 个 goroutine
都结束了，不能保证主 goroutine 不会更早退出（这个问题在第五节）。

## 三、`go` 语句本身几乎不花时间

### 代码

```go
// demoGoIsNotBlocking 说明 go 语句本身几乎立刻返回，主 goroutine 不会被拖住。
func demoGoIsNotBlocking() {
	start := time.Now()
	result := make(chan int, 1) // 有缓冲：发送方不必等接收方

	go func() {
		work(120 * time.Millisecond)
		result <- 42
	}()

	fmt.Printf("    go 语句之后主 goroutine 继续往下走，只花了 %v\n",
		time.Since(start).Round(time.Millisecond))
	v := <-result
	fmt.Printf("    %v 之后才从 channel 里拿到结果: %d\n",
		time.Since(start).Round(time.Millisecond), v)
}
```

### 解析

输出里第一行是 `0s`，第二行是 `121ms`：说明「启动 goroutine」几乎不花时间，
真正的等待发生在 `<-result`。

这就是并发的核心价值：**把「等」交给别的 goroutine，自己先去做别的事**。
如果这里没有 `<-result`，`result` 只在缓冲里躺一下就随程序结束了，
不会有任何报错提醒你「结果丢了」。

## 四、循环变量捕获：Go 1.22 的分水岭

### 代码

```go
// demoPerIterationVar 展示 Go 1.22 起 for 循环变量每轮独立。
func demoPerIterationVar() {
	values := make([]int, 4)

	var wg sync.WaitGroup
	for i := range 4 { // Go 1.22+ 支持 range over int
		wg.Add(1)
		go func() {
			defer wg.Done()
			values[i] = i // 捕获的是本轮独有的 i
		}()
	}
	wg.Wait()
	fmt.Println("    闭包直接捕获循环变量:", values)
	fmt.Println("    Go 1.22 之前这里往往是 [4 4 4 4]，当时必须写 i := i 或作为参数传入")
}
```

### 解析

在 Go 1.22 之前，`for` 的循环变量是整个循环共享的一个变量，闭包捕获的是它本身，
所以 goroutine 真正执行时看到的往往是循环结束后的值——并发代码里这就是
「所有任务都用了最后一个参数」的经典 bug。

1.22 起每轮迭代都有独立的变量，所以下面的写法是正确的（本模块的 `go.mod` 声明
`go 1.22`，可以直接这么写）：

```go
for i := range 4 {
	go func() { values[i] = i }() // 安全
}
```

读老代码时看到下面两种写法，不要以为是多余的，那是在兼容旧语义：

```go
for i := 0; i < 4; i++ {
	i := i                                       // 影子变量
	go func() { ... }()
}

for i := 0; i < 4; i++ {
	go func(i int) { ... }(i)                    // 参数传值
}
```

第二种（参数传值）在今天的项目里依然值得推荐：意图更明显，
也不依赖读者对语言版本变化的记忆。

## 五、GOMAXPROCS：并行度不等于并发度

### 代码

```go
// demoRuntimeInfo 打印与调度相关的运行期信息。
func demoRuntimeInfo() {
	fmt.Printf("    逻辑 CPU = %d，GOMAXPROCS = %d，当前 goroutine 数 = %d\n",
		runtime.NumCPU(), runtime.GOMAXPROCS(0), runtime.NumGoroutine())
	fmt.Println("    GOMAXPROCS 限制的是同时执行 Go 代码的线程数（并行度）")
	fmt.Println("    它不限制 goroutine 的数量：几万个 goroutine 可以只用 8 个线程跑（并发度）")
}
```

### 解析

- `runtime.NumCPU()`：机器的逻辑 CPU 数。
- `runtime.GOMAXPROCS(0)`：查询当前的 P 数量（不修改），默认等于 CPU 数。
  它决定**同时执行 Go 代码的线程数上限**，即并行度。
- `runtime.NumGoroutine()`：当前活着的 goroutine 数量，排查泄漏时最常用的一行代码。

关键区别：`GOMAXPROCS` 不限制 goroutine 的数量。几万个 goroutine 完全可以
只在 8 个线程上轮流跑——因为绝大多数 goroutine 都在等 IO，不是在算。

## 六、主 goroutine 退出 = 进程退出

### 代码

```go
func main() {
	fmt.Println("[1] 串行 vs 并发")
	const n = 5
	fmt.Printf("    串行执行 %d 个 %v 的任务: %v\n", n, jobDelay, serial(n).Round(time.Millisecond))
	fmt.Printf("    并发执行同样的任务: %v\n", parallel(n).Round(time.Millisecond))

	fmt.Println("\n[2] go 语句立刻返回")
	demoGoIsNotBlocking()

	fmt.Println("\n[3] 循环变量捕获")
	demoPerIterationVar()

	fmt.Println("\n[4] 运行期信息")
	demoRuntimeInfo()

	fmt.Println("\n[5] 主 goroutine 退出 = 进程退出")
	fmt.Println("    下面这个 goroutine 要 200ms 后才打印，但 main 马上就要返回了：")
	go func() {
		time.Sleep(200 * time.Millisecond)
		fmt.Println("    这一行永远看不到：进程已退出，goroutine 被静默终止")
	}()
	// main 返回 → 进程退出 → 其他 goroutine 直接消失，不会有任何清理或提示。
}
```

### 解析

这是全篇最重要的一条规则：**Go 程序不会等任何 goroutine，`main` 一返回进程就结束。**
没有「优雅等待」，没有警告，没有 `defer` 被执行——所有 goroutine 连同它们的栈内存一起消失。

所以「任务丢了」「日志没打出来」「数据没写进去」这类问题的第一反应应该是：
**这个 goroutine 的生命周期是不是没人管？**

## 运行输出

```text
[1] 串行 vs 并发
    串行执行 5 个 100ms 的任务: 504ms
    并发求和 = 55，校验通过: true
    并发执行同样的任务: 101ms

[2] go 语句立刻返回
    go 语句之后主 goroutine 继续往下走，只花了 0s
    121ms 之后才从 channel 里拿到结果: 42

[3] 循环变量捕获
    闭包直接捕获循环变量: [0 1 2 3]
    Go 1.22 之前这里往往是 [4 4 4 4]，当时必须写 i := i 或作为参数传入

[4] 运行期信息
    逻辑 CPU = 18，GOMAXPROCS = 18，当前 goroutine 数 = 1
    GOMAXPROCS 限制的是同时执行 Go 代码的线程数（并行度）
    它不限制 goroutine 的数量：几万个 goroutine 可以只用 8 个线程跑（并发度）

[5] 主 goroutine 退出 = 进程退出
    下面这个 goroutine 要 200ms 后才打印，但 main 马上就要返回了：
```

注意输出到这里就结束了：那个要 200ms 才打印的 goroutine 永远不会执行。

## 常见错误

| 写法 | 后果 |
|---|---|
| 忘记 `Wait` | main 提前退出，任务丢失且不报错 |
| `wg.Add(1)` 写在 goroutine 内部 | `Wait` 可能提前返回 |
| 多个 goroutine 写同一个变量 | 数据竞争，结果随机错误 |
| 用 `Println` 顺序判断执行顺序 | 结论只在「这一次」成立 |
| 认为「并发一定更快」 | CPU 密集且已占满核时，反而增加调度开销 |

## 练习

1. 把 `parallel` 改成「最多 3 个任务同时执行」，其余排队等待。
2. 让每个任务有 20% 概率失败，收集所有错误并用 `errors.Join` 汇总返回。
3. 故意删掉 `wg.Wait()`，观察输出变化，并解释为什么「校验通过」这行有时仍然打印 true。

## 小结

- `go f()` 启动一个 goroutine，它几乎立刻返回。
- `sync.WaitGroup` 负责等它结束，`Add` 在 `go` 之前、`Done` 用 `defer`。
- Go 1.22 起循环变量每轮独立，闭包可以放心捕获。
- `GOMAXPROCS` 限制并行度，不限制 goroutine 数量。
- `main` 返回时进程立即结束，没有任何 goroutine 会被等待。

下一篇我们会解决「结果怎么传回来」的问题：channel 的五条规则。
