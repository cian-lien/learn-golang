# 第 6 部分：指针和并发

前半段讲**指针**，后半段讲**并发**。并发是 Go 最出名的特性，也是很多公司用 Go 写后端的主要原因。

# 一、指针

## 1. 指针是什么

每个变量都存放在内存里的某个位置，叫**地址**。**指针就是一个存放地址的变量**。

| 符号 | 作用 | 读法 |
|---|---|---|
| `&x` | 取出 x 的地址 | "x 的地址" |
| `*p` | 通过地址找到原来的值 | "p 指向的值" |

```go
x := 10
p := &x              // p 存的是 x 的地址

fmt.Printf("%T\n", p)  // *int    ← 类型是"指向 int 的指针"
fmt.Println(*p)        // 10

*p = 20              // 通过 p 修改 x
fmt.Println(x)       // 20
```

打个比方：`x` 是一间房子，`&x` 是门牌号，`p` 是一张写着门牌号的纸条。`*p = 20` 就是"按纸条上的地址找到房子，把里面的东西换成 20"。

`*` 有两种用法：

- 写在**类型**前面，比如 `*int`、`*User`：表示"指针类型"
- 写在**变量**前面，比如 `*p`：表示"取出指向的值"

## 2. 为什么需要指针

**Go 传参时会复制一份**，函数里直接修改参数，外面看不到：

```go
func doubleWrong(n int) {
    n = n * 2        // 改的是副本
}

func double(n *int) {
    *n = *n * 2      // 通过地址改原来的值
}

x := 10
doubleWrong(x)
fmt.Println(x)   // 10   ← 没变
double(&x)       // 传入 x 的地址
fmt.Println(x)   // 20
```

这就是第 5 部分"指针接收者"的原理。

注意区分：

- `*b = ...`：改的是**指针指向的值**，外面会受影响
- `b = ...`：改的是**指针变量本身**（函数里的副本），外面不受影响

**用指针的三种常见情况：**

1. 函数需要**修改**传进来的值
2. 结构体很大，不想每次复制
3. 需要表示"没有值"：指针的零值是 `nil`

## 3. 结构体指针

```go
u := &User{Name: "小明", Age: 20}   // u 的类型是 *User
u.Age++                            // 直接用点号，不需要写 (*u).Age
fmt.Println(u.Age)                 // 21
```

## 4. ⚠️ nil 指针会导致崩溃

```go
var u *User              // 零值是 nil
fmt.Println(u == nil)    // true
fmt.Println(u.Name)      // ❌ panic: invalid memory address or nil pointer dereference
```

这是 Go 程序**最常见的崩溃原因之一**，类似 JS 的 `Cannot read properties of null`。不确定指针是否为 nil 时，先判断 `if u != nil`。

## 5. 什么时候不需要指针

**切片和 map 本身就能在函数里修改元素**，不需要再传指针：

```go
func setFirst(s []int) {
    s[0] = 100    // 外面的切片也会被改到
}
```

Go 的指针**不能做加减运算**（C 语言可以），比 C 安全很多。

---

# 二、并发

JS 是**单线程**的。`async/await` 只是让你在等待时不卡住，同一时刻真正在执行的代码只有一段。

Go 可以**真正地同时执行**多个任务，充分利用多核 CPU。

## 1. goroutine：用 `go` 开启并发任务

```go
func sayHello() {
    fmt.Println("你好")
}

func main() {
    go sayHello()            // 开启一个 goroutine
    fmt.Println("main 结束")
}
```

很可能**只打印 "main 结束"**。因为 **main 函数结束，整个程序就退出了**，不会等其他 goroutine。

goroutine 非常轻量，开几万个都没问题。

## 2. WaitGroup：等所有 goroutine 完成

```go
import "sync"

func main() {
    var wg sync.WaitGroup

    for i := 1; i <= 3; i++ {
        wg.Add(1)              // 计数 +1
        go func() {
            defer wg.Done()    // 完成时计数 -1
            fmt.Println("goroutine", i)
        }()
    }

    wg.Wait()                  // 等计数归零
    fmt.Println("全部完成")
}
```

- `go func() { ... }()`：开启一个匿名函数的 goroutine，末尾 `()` 表示立即调用
- **执行顺序是不确定的**，每次运行都可能不一样
- Go 1.22 起每轮循环都会创建新的 `i`，直接在 goroutine 里用没问题。旧代码常见 `go func(n int) { ... }(i)` 这种传参写法，是为了避开老版本的坑

## 3. channel：goroutine 之间传递数据

channel 像一根管道：一边放数据，另一边取出来。

```go
ch := make(chan string)    // 传递 string 的通道

go func() {
    ch <- "你好"            // 发送
}()

msg := <-ch                // 接收（拿不到会一直等）
fmt.Println(msg)           // 你好
```

- `ch <- 值`：数据流**进**通道（发送）
- `<-ch`：数据从通道流**出**（接收）

**接收会等待**，类似 JS 的 `await`：`msg := <-ch` ≈ `const msg = await promise`。

### 例子：多个任务同时执行，收集结果

```go
func worker(id int, results chan string) {
    time.Sleep(time.Duration(id*100) * time.Millisecond)  // 模拟耗时操作
    results <- fmt.Sprintf("任务 %d 完成", id)
}

func main() {
    results := make(chan string)

    for i := 1; i <= 3; i++ {
        go worker(i, results)
    }

    for i := 0; i < 3; i++ {
        fmt.Println(<-results)
    }
}
```

3 个任务同时跑，总共约 0.3 秒，而不是 0.6 秒。收了 3 次结果，channel 本身就起到了"等待"的作用。

### ⚠️ 死锁

```go
func main() {
    ch := make(chan int)
    ch <- 1              // 等待有人接收……永远没人来
    fmt.Println(<-ch)
}
// fatal error: all goroutines are asleep - deadlock!
```

普通 channel，**发送方要等到有人接收才能继续**。看到 `deadlock!`，就检查发送和接收是否配对。

### 带缓冲的 channel 和 close

```go
ch := make(chan int, 3)   // 缓冲区能放 3 个，放满之前发送不等待

go func() {
    for i := 1; i <= 3; i++ {
        ch <- i * 10
    }
    close(ch)             // 发完了，关闭通道
}()

for v := range ch {       // 一直接收，直到通道关闭
    fmt.Println(v)        // 10 20 30
}
```

**只有发送方才能 close**，关闭后再发送会 panic。

## 4. select：同时等待多个 channel

哪个先有数据就执行哪个，常用来设置超时：

```go
result := make(chan string)

go func() {
    time.Sleep(2 * time.Second)
    result <- "结果"
}()

select {
case r := <-result:
    fmt.Println(r)
case <-time.After(500 * time.Millisecond):
    fmt.Println("超时了")
}
```

和 JS 的 `Promise.race([fetchData(), timeout(500)])` 是一个思路。

## 5. ⚠️ 数据竞争和互斥锁

多个 goroutine **同时修改同一个变量**，结果会出错：

```go
count := 0
var wg sync.WaitGroup

for i := 0; i < 1000; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        count++
    }()
}
wg.Wait()
fmt.Println(count)   // 期望 1000，实际可能是 997、985……
```

`count++` 分三步：读出、加 1、写回。两个 goroutine 可能同时读到 500，都写回 501，少加了一次。

用**互斥锁（Mutex）**保证同一时刻只有一个 goroutine 能修改：

```go
var mu sync.Mutex

go func() {
    defer wg.Done()
    mu.Lock()       // 上锁，其他 goroutine 排队
    count++
    mu.Unlock()     // 解锁
}()
// 结果稳定是 1000
```

用 `go run -race main.go` 可以检测数据竞争。

---

## 📝 练习

### 练习 1：预测输出

代码在 [pointer/main.go](pointer/main.go)：

```go
func change(a int, b *int) {
    a = 100
    *b = 100
}

func main() {
    x, y := 1, 1
    change(x, &y)
    p := &x
    *p = 5
    fmt.Println(x, y)
}
```

追问：如果把 `*b = 100` 改成 `b = nil`，y 会是多少？

<details>
<summary>答案</summary>

```
5 100
```

- **y = 100**：传进去的是 `&y`，`*b = 100` 通过地址改到了外面的 y
- **x 在函数里没变**：传进去的是副本，`a = 100` 只改了副本
- **x 最后是 5**：`p := &x`，`*p = 5` 通过地址把 x 改成 5

追问答案：**y 是 1**。`b` 本身是复制进函数的"纸条"，`b = nil` 只是擦掉了函数里的纸条，没有通过地址动 y。

</details>

### 练习 2：并发计算平方和

用 goroutine 和 channel，**同时**计算 1 到 5 每个数的平方，在 main 里把结果加起来并打印。正确答案是 **55**。

<details>
<summary>答案</summary>

见 [squares/main.go](squares/main.go)：

```go
ch := make(chan int)

for i := 1; i <= 5; i++ {
    go func(n int) {
        ch <- n * n
    }(i)
}

sum := 0
for i := 0; i < 5; i++ {
    sum += <-ch
}
fmt.Println(sum)   // 55
```

**常见错误：边界差一（off-by-one）**

| 循环写法 | 实际发送 | 结果 |
|---|---|---|
| `i := 0; i < 5` | 0、1、4、9、16 | 30 ❌ |
| `i := 1; i < 5` | 1、4、9、16 | 30 ❌ |
| `i := 1; i <= 5` | 1、4、9、16、25 | 55 ✅ |

检查循环时，把第一个值和最后一个值代进去想一遍。

</details>

---

**本部分核心：`&` 取地址、`*` 取值；传指针才能在函数里修改原来的值；小心 nil 指针；`go` 开启 goroutine，main 结束程序就退出；用 channel 传递数据，接收会等待；多个 goroutine 改同一个变量要加锁。**
