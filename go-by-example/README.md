# Go by Example 中文示例手册

按官网目录顺序收录 85 个主题。每节都有中文讲解、独立编写的完整 Go 代码、运行命令和可对照的输出。[Go by Example 原站](https://gobyexample.com/)

使用 Go 1.25 或更高版本。下文的命令都从本手册所在目录执行，采用 macOS / Linux shell 语法。示例 83 使用 Unix 进程接口。

可以先运行第一例。

```sh
go run ./01-hello-world
```

## 目录

| 编号 | 主题 | 示例 |
| --- | --- | --- |
| 01–10 | 语法与集合 | [你好，世界](#ex-01) · [基本值与运算](#ex-02) · [变量](#ex-03) · [常量](#ex-04) · [循环](#ex-05) · [条件分支](#ex-06) · [多路选择](#ex-07) · [数组](#ex-08) · [切片](#ex-09) · [映射](#ex-10) |
| 11–18 | 函数、指针与字符串 | [函数](#ex-11) · [多返回值](#ex-12) · [可变参数](#ex-13) · [闭包](#ex-14) · [递归](#ex-15) · [遍历内置类型](#ex-16) · [指针](#ex-17) · [字符串、字节与 rune](#ex-18) |
| 19–27 | 类型、泛型与错误 | [结构体](#ex-19) · [方法](#ex-20) · [接口](#ex-21) · [枚举式常量](#ex-22) · [结构体嵌入](#ex-23) · [泛型](#ex-24) · [遍历迭代器](#ex-25) · [错误值](#ex-26) · [自定义错误](#ex-27) |
| 28–37 | goroutine 与通道 | [启动 goroutine](#ex-28) · [通道](#ex-29) · [带缓冲的通道](#ex-30) · [用通道同步](#ex-31) · [单向通道](#ex-32) · [等待多个通道](#ex-33) · [超时](#ex-34) · [非阻塞通道操作](#ex-35) · [关闭通道](#ex-36) · [遍历通道](#ex-37) |
| 38–45 | 定时、任务调度与共享状态 | [一次性定时器](#ex-38) · [周期定时器](#ex-39) · [工作池](#ex-40) · [等待一组任务](#ex-41) · [限流](#ex-42) · [原子计数器](#ex-43) · [互斥锁](#ex-44) · [用 goroutine 管理状态](#ex-45) |
| 46–50 | 排序与异常处理 | [排序](#ex-46) · [自定义排序](#ex-47) · [触发 panic](#ex-48) · [延迟执行](#ex-49) · [恢复 panic](#ex-50) |
| 51–56 | 文本与数据格式 | [常用字符串函数](#ex-51) · [字符串格式化](#ex-52) · [文本模板](#ex-53) · [正则表达式](#ex-54) · [JSON 编解码](#ex-55) · [XML 编解码](#ex-56) |
| 57–64 | 时间、解析与编码 | [日期、时间与时长](#ex-57) · [Unix 时间戳](#ex-58) · [时间格式化与解析](#ex-59) · [随机数](#ex-60) · [解析数字](#ex-61) · [解析 URL](#ex-62) · [SHA-256 摘要](#ex-63) · [Base64 编解码](#ex-64) |
| 65–72 | 文件与测试 | [读取文件](#ex-65) · [写入文件](#ex-66) · [逐行过滤](#ex-67) · [文件路径](#ex-68) · [目录操作](#ex-69) · [临时文件与目录](#ex-70) · [嵌入文件](#ex-71) · [单元测试与基准测试](#ex-72) |
| 73–77 | 命令行与日志 | [命令行位置参数](#ex-73) · [命令行选项](#ex-74) · [命令行子命令](#ex-75) · [环境变量](#ex-76) · [日志](#ex-77) |
| 78–85 | 网络、上下文与进程 | [HTTP 客户端](#ex-78) · [HTTP 服务端](#ex-79) · [TCP 服务端](#ex-80) · [取消与截止时间](#ex-81) · [启动子进程](#ex-82) · [替换当前进程](#ex-83) · [接收系统信号](#ex-84) · [退出状态码](#ex-85) |

## 01–10　语法与集合

<a id="ex-01"></a>

### 01　你好，世界 · Hello World

一个可执行程序从 main 包的 main 函数开始。运行后，终端会打印一行问候。 [原文](https://gobyexample.com/hello-world) · [源码](01-hello-world/main.go)

```go
package main

import (
	"fmt"
)

func main() {
	fmt.Println("你好，Go！")
}
```

**运行**

```sh
go run ./01-hello-world
```

**输出**

```text
你好，Go！
```

<a id="ex-02"></a>

### 02　基本值与运算 · Values

试着预测四个结果。两个整数相除会舍去小数部分，把一个操作数写成浮点数就能得到小数。 [原文](https://gobyexample.com/values) · [源码](02-values/main.go)

```go
package main

import (
	"fmt"
)

func main() {
	fmt.Println("Go" + "语言")
	fmt.Println(7 / 2)
	fmt.Println(7.0 / 2)
	fmt.Println(3 < 5 && 5 < 8)
}
```

**运行**

```sh
go run ./02-values
```

**输出**

```text
Go语言
3
3.5
true
```

<a id="ex-03"></a>

### 03　变量 · Variables

var 可以显式指定类型并使用该类型的零值。函数内的短声明 := 会从右侧推断类型。 [原文](https://gobyexample.com/variables) · [源码](03-variables/main.go)

```go
package main

import (
	"fmt"
)

func main() {
	var count int
	label := "待读"
	fmt.Println(label, count)
	count = 2
	fmt.Println(label, count)
}
```

**运行**

```sh
go run ./03-variables
```

**输出**

```text
待读 0
待读 2
```

<a id="ex-04"></a>

### 04　常量 · Constants

给换算系数起一个名字。未显式指定类型的数值常量可以用在类型兼容的表达式中。 [原文](https://gobyexample.com/constants) · [源码](04-constants/main.go)

```go
package main

import (
	"fmt"
)

func main() {
	const secondsPerMinute = 60
	const minutes = 3
	fmt.Println(minutes * secondsPerMinute)
	fmt.Println(1.5 * secondsPerMinute)
}
```

**运行**

```sh
go run ./04-constants
```

**输出**

```text
180
90
```

<a id="ex-05"></a>

### 05　循环 · For

逐项累加 1 到 5 之间的奇数。continue 跳过本轮剩余语句，然后进入下一轮。 [原文](https://gobyexample.com/for) · [源码](05-for/main.go)

```go
package main

import (
	"fmt"
)

func main() {
	sum := 0
	for n := 1; n <= 5; n++ {
		if n%2 == 0 {
			continue
		}
		sum += n
	}
	fmt.Println(sum)
}
```

**运行**

```sh
go run ./05-for
```

**输出**

```text
9
```

<a id="ex-06"></a>

### 06　条件分支 · If/Else

可以在 if 的条件前声明变量。这里的 n 在 if 和 else 两个分支内都可用。 [原文](https://gobyexample.com/if-else) · [源码](06-if-else/main.go)

```go
package main

import (
	"fmt"
)

func main() {
	tasks := []string{"洗衣", "读书"}
	if n := len(tasks); n == 0 {
		fmt.Println("今天没有待办")
	} else {
		fmt.Printf("还有 %d 项待办\n", n)
	}
}
```

**运行**

```sh
go run ./06-if-else
```

**输出**

```text
还有 2 项待办
```

<a id="ex-07"></a>

### 07　多路选择 · Switch

省略 switch 后面的表达式，可以按顺序判断多个布尔条件。Go 默认只执行第一个匹配的分支。 [原文](https://gobyexample.com/switch) · [源码](07-switch/main.go)

```go
package main

import (
	"fmt"
)

func main() {
	score := 82
	switch {
	case score >= 90:
		fmt.Println("优秀")
	case score >= 60:
		fmt.Println("通过")
	default:
		fmt.Println("继续练习")
	}
}
```

**运行**

```sh
go run ./07-switch
```

**输出**

```text
通过
```

<a id="ex-08"></a>

### 08　数组 · Arrays

[3]int 的长度是类型的一部分。数组赋值会复制所有元素，修改副本后可以比较两者。 [原文](https://gobyexample.com/arrays) · [源码](08-arrays/main.go)

```go
package main

import (
	"fmt"
)

func main() {
	scores := [3]int{70, 80, 90}
	copied := scores
	copied[0] = 100
	fmt.Println(scores)
	fmt.Println(copied)
}
```

**运行**

```sh
go run ./08-arrays
```

**输出**

```text
[70 80 90]
[100 80 90]
```

<a id="ex-09"></a>

### 09　切片 · Slices

切片可以共享底层数组。修改切片元素会反映到原切片，追加元素时要接住 append 返回的新切片。 [原文](https://gobyexample.com/slices) · [源码](09-slices/main.go)

```go
package main

import (
	"fmt"
)

func main() {
	tasks := []string{"吃饭", "散步", "写代码"}
	firstTwo := tasks[:2]
	firstTwo[0] = "做饭"
	fmt.Println(tasks)
	tasks = append(tasks, "读书")
	fmt.Println(len(tasks), tasks)
}
```

**运行**

```sh
go run ./09-slices
```

**输出**

```text
[做饭 散步 写代码]
4 [做饭 散步 写代码 读书]
```

<a id="ex-10"></a>

### 10　映射 · Maps

map 按键存取数据。双返回值能区分“键存在且值为零”和“键不存在”，delete 可以删除键。 [原文](https://gobyexample.com/maps) · [源码](10-maps/main.go)

```go
package main

import (
	"fmt"
)

func main() {
	stock := map[string]int{"茶": 0}
	count, ok := stock["茶"]
	fmt.Println(count, ok)
	delete(stock, "茶")
	count, ok = stock["茶"]
	fmt.Println(count, ok)
}
```

**运行**

```sh
go run ./10-maps
```

**输出**

```text
0 true
0 false
```

## 11–18　函数、指针与字符串

<a id="ex-11"></a>

### 11　函数 · Functions

参数名后写类型，参数列表后写返回类型。调用 price 计算三本笔记本的总价。 [原文](https://gobyexample.com/functions) · [源码](11-functions/main.go)

```go
package main

import (
	"fmt"
)

func price(quantity, unitPrice int) int {
	return quantity * unitPrice
}

func main() {
	fmt.Println(price(3, 12))
}
```

**运行**

```sh
go run ./11-functions
```

**输出**

```text
36
```

<a id="ex-12"></a>

### 12　多返回值 · Multiple Return Values

一个函数可以同时返回查找位置和是否找到。下标 0 也是合法结果，所以用第二个返回值表达成功与否。 [原文](https://gobyexample.com/multiple-return-values) · [源码](12-multiple-return-values/main.go)

```go
package main

import (
	"fmt"
)

func indexOf(names []string, wanted string) (int, bool) {
	for i, name := range names {
		if name == wanted {
			return i, true
		}
	}
	return -1, false
}

func main() {
	names := []string{"阿青", "小林"}
	fmt.Println(indexOf(names, "阿青"))
	fmt.Println(indexOf(names, "小周"))
}
```

**运行**

```sh
go run ./12-multiple-return-values
```

**输出**

```text
0 true
-1 false
```

<a id="ex-13"></a>

### 13　可变参数 · Variadic Functions

...int 允许调用者传任意个整数。已有的切片可以用尾随的 ... 展开后传入。 [原文](https://gobyexample.com/variadic-functions) · [源码](13-variadic-functions/main.go)

```go
package main

import (
	"fmt"
)

func total(prices ...int) int {
	sum := 0
	for _, price := range prices {
		sum += price
	}
	return sum
}

func main() {
	fmt.Println(total(6, 8))
	cart := []int{3, 4, 5}
	fmt.Println(total(cart...))
}
```

**运行**

```sh
go run ./13-variadic-functions
```

**输出**

```text
14
12
```

<a id="ex-14"></a>

### 14　闭包 · Closures

返回的函数会记住外层的 n。每次调用 newCounter 都会创建独立的计数状态。 [原文](https://gobyexample.com/closures) · [源码](14-closures/main.go)

```go
package main

import (
	"fmt"
)

func newCounter() func() int {
	n := 0
	return func() int {
		n++
		return n
	}
}

func main() {
	a, b := newCounter(), newCounter()
	fmt.Println(a(), a(), b())
}
```

**运行**

```sh
go run ./14-closures
```

**输出**

```text
1 2 1
```

<a id="ex-15"></a>

### 15　递归 · Recursion

把“从 n 加到 1”缩小为“n 加上从 n-1 加到 1 的和”。n 到 0 时返回，这就是递归的终止条件。 [原文](https://gobyexample.com/recursion) · [源码](15-recursion/main.go)

```go
package main

import (
	"fmt"
)

func sumTo(n int) int {
	if n <= 0 {
		return 0
	}
	return n + sumTo(n-1)
}

func main() {
	fmt.Println(sumTo(4))
}
```

**运行**

```sh
go run ./15-recursion
```

**输出**

```text
10
```

<a id="ex-16"></a>

### 16　遍历内置类型 · Range over Built-in Types

遍历切片会得到下标和值。遍历整数 n 则依次得到 0 到 n-1，可以用来重复执行一段代码。 [原文](https://gobyexample.com/range-over-built-in-types) · [源码](16-range-over-built-in-types/main.go)

```go
package main

import (
	"fmt"
)

func main() {
	for i, task := range []string{"读书", "跑步"} {
		fmt.Println(i, task)
	}
	for i := range 3 {
		fmt.Println(i)
	}
}
```

**运行**

```sh
go run ./16-range-over-built-in-types
```

**输出**

```text
0 读书
1 跑步
0
1
2
```

<a id="ex-17"></a>

### 17　指针 · Pointers

& 取变量的地址，* 访问指针指向的值。把库存的地址传入函数，就能更新调用方持有的库存。 [原文](https://gobyexample.com/pointers) · [源码](17-pointers/main.go)

```go
package main

import (
	"fmt"
)

func restock(stock *int, quantity int) {
	*stock += quantity
}

func main() {
	stock := 2
	restock(&stock, 3)
	fmt.Println(stock)
}
```

**运行**

```sh
go run ./17-pointers
```

**输出**

```text
5
```

<a id="ex-18"></a>

### 18　字符串、字节与 rune · Strings and Runes

len 计算字节数。range 按 Unicode 码点解码，返回该码点起始的字节下标。观察最后一个字母的下标。 [原文](https://gobyexample.com/strings-and-runes) · [源码](18-strings-and-runes/main.go)

```go
package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	word := "A茶B"
	fmt.Println(len(word), utf8.RuneCountInString(word))
	for offset, r := range word {
		fmt.Printf("%d %c\n", offset, r)
	}
}
```

**运行**

```sh
go run ./18-strings-and-runes
```

**输出**

```text
5 3
0 A
1 茶
4 B
```

## 19–27　类型、泛型与错误

<a id="ex-19"></a>

### 19　结构体 · Structs

结构体把不同类型的字段组合成一个值。只给 Title 赋初值时，Done 会取 bool 的零值 false。 [原文](https://gobyexample.com/structs) · [源码](19-structs/main.go)

```go
package main

import (
	"fmt"
)

type Task struct {
	Title string
	Done  bool
}

func main() {
	task := Task{Title: "浇花"}
	fmt.Println(task.Title, task.Done)
	task.Done = true
	fmt.Println(task.Title, task.Done)
}
```

**运行**

```sh
go run ./19-structs
```

**输出**

```text
浇花 false
浇花 true
```

<a id="ex-20"></a>

### 20　方法 · Methods

接收者把方法关联到具体类型。指针接收者可以修改原值，调用时 Go 会为可寻址的变量自动取地址。 [原文](https://gobyexample.com/methods) · [源码](20-methods/main.go)

```go
package main

import (
	"fmt"
)

type Wallet struct{ Balance int }

func (w *Wallet) Deposit(amount int) {
	w.Balance += amount
}

func main() {
	w := Wallet{Balance: 10}
	w.Deposit(5)
	fmt.Println(w.Balance)
}
```

**运行**

```sh
go run ./20-methods
```

**输出**

```text
15
```

<a id="ex-21"></a>

### 21　接口 · Interfaces

接口描述需要哪些方法。Book 和 Film 都实现了 Title，可以传给同一个接收接口的函数。 [原文](https://gobyexample.com/interfaces) · [源码](21-interfaces/main.go)

```go
package main

import (
	"fmt"
)

type Titled interface{ Title() string }

type Book struct{ Name string }

func (b Book) Title() string { return b.Name }

type Film struct{ Name string }

func (f Film) Title() string { return f.Name }

func printTitle(item Titled) { fmt.Println(item.Title()) }

func main() {
	printTitle(Book{Name: "城南旧事"})
	printTitle(Film{Name: "千与千寻"})
}
```

**运行**

```sh
go run ./21-interfaces
```

**输出**

```text
城南旧事
千与千寻
```

<a id="ex-22"></a>

### 22　枚举式常量 · Enums

用自定义类型和 iota 定义一组相关常量。进行状态判断时，代码可以直接使用有含义的名称。 [原文](https://gobyexample.com/enums) · [源码](22-enums/main.go)

```go
package main

import (
	"fmt"
)

type Status int

const (
	Pending Status = iota
	Running
	Done
)

func main() {
	s := Running
	fmt.Println(Pending, Running, Done)
	fmt.Println(s == Running)
}
```

**运行**

```sh
go run ./22-enums
```

**输出**

```text
0 1 2
true
```

<a id="ex-23"></a>

### 23　结构体嵌入 · Struct Embedding

嵌入成员的方法可以提升到外层结构体。这里既能从 article 直接调用 Byline，也能显式访问 article.Author。 [原文](https://gobyexample.com/struct-embedding) · [源码](23-struct-embedding/main.go)

```go
package main

import (
	"fmt"
)

type Author struct{ Name string }

func (a Author) Byline() string { return "作者 " + a.Name }

type Article struct {
	Author
	Title string
}

func main() {
	article := Article{Author: Author{Name: "林"}, Title: "春天"}
	fmt.Println(article.Title, article.Byline())
	fmt.Println(article.Author.Name)
}
```

**运行**

```sh
go run ./23-struct-embedding
```

**输出**

```text
春天 作者 林
林
```

<a id="ex-24"></a>

### 24　泛型 · Generics

用类型参数 T 让同一个函数检查不同类型的切片。comparable 约束允许在函数里用 == 比较两个 T。 [原文](https://gobyexample.com/generics) · [源码](24-generics/main.go)

```go
package main

import (
	"fmt"
)

func contains[T comparable](items []T, target T) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func main() {
	fmt.Println(contains([]int{4, 8}, 8))
	fmt.Println(contains([]string{"茶", "水"}, "奶"))
}
```

**运行**

```sh
go run ./24-generics
```

**输出**

```text
true
false
```

<a id="ex-25"></a>

### 25　遍历迭代器 · Range over Iterators

迭代器通过 yield 逐项交出结果。外层循环 break 后，yield 会返回 false，迭代器应随即停止。 [原文](https://gobyexample.com/range-over-iterators) · [源码](25-range-over-iterators/main.go)

```go
package main

import (
	"fmt"
	"iter"
)

func countdown(from int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for n := from; n > 0; n-- {
			if !yield(n) {
				return
			}
		}
	}
}

func main() {
	for n := range countdown(5) {
		fmt.Println(n)
		if n == 3 {
			break
		}
	}
}
```

**运行**

```sh
go run ./25-range-over-iterators
```

**输出**

```text
5
4
3
```

<a id="ex-26"></a>

### 26　错误值 · Errors

用 %w 给错误补上操作背景。调用方可以打印完整消息，也可以用 errors.Is 检查它是否包装了某个已知错误。 [原文](https://gobyexample.com/errors) · [源码](26-errors/main.go)

```go
package main

import (
	"errors"
	"fmt"
)

var errOutOfStock = errors.New("库存不足")

func reserve(n int) error {
	if n > 3 {
		return fmt.Errorf("预订 %d 件商品: %w", n, errOutOfStock)
	}
	return nil
}

func main() {
	if err := reserve(5); err != nil {
		fmt.Println(err)
		fmt.Println(errors.Is(err, errOutOfStock))
	}
}
```

**运行**

```sh
go run ./26-errors
```

**输出**

```text
预订 5 件商品: 库存不足
true
```

<a id="ex-27"></a>

### 27　自定义错误 · Custom Errors

用结构体携带调用方需要的细节，再实现 Error 方法。errors.As 可以取出该类型的错误并读取字段。 [原文](https://gobyexample.com/custom-errors) · [源码](27-custom-errors/main.go)

```go
package main

import (
	"errors"
	"fmt"
)

type LimitError struct{ Limit int }

func (e *LimitError) Error() string {
	return fmt.Sprintf("每次最多借 %d 本", e.Limit)
}

func borrow(n int) error {
	if n > 2 {
		return &LimitError{Limit: 2}
	}
	return nil
}

func main() {
	err := borrow(3)
	var limit *LimitError
	if errors.As(err, &limit) {
		fmt.Println(limit.Error())
		fmt.Println("上限", limit.Limit)
	}
}
```

**运行**

```sh
go run ./27-custom-errors
```

**输出**

```text
每次最多借 2 本
上限 2
```

## 28–37　goroutine 与通道

<a id="ex-28"></a>

### 28　启动 goroutine · Goroutines

go 会让函数独立执行。main 如果提前返回，程序就会结束，所以这里先打印自己的消息，再等待后台任务完成。 [原文](https://gobyexample.com/goroutines) · [源码](28-goroutines/main.go)

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("后台任务已完成")
	}()
	fmt.Println("主函数已启动")
	wg.Wait()
}
```

**运行**

```sh
go run ./28-goroutines
```

**一种可能的输出，行的顺序可能变化**

```text
主函数已启动
后台任务已完成
```

<a id="ex-29"></a>

### 29　通道 · Channels

无缓冲通道把发送方和接收方配对。后台计算面积并发送结果，主函数从通道收到值后继续执行。 [原文](https://gobyexample.com/channels) · [源码](29-channels/main.go)

```go
package main

import (
	"fmt"
)

func main() {
	area := make(chan int)
	go func() {
		area <- 3 * 4
	}()
	result := <-area
	fmt.Println("面积", result)
}
```

**运行**

```sh
go run ./29-channels
```

**输出**

```text
面积 12
```

<a id="ex-30"></a>

### 30　带缓冲的通道 · Channel Buffering

容量为 2 的通道能暂存两条消息，后续接收会按发送顺序取出它们。缓冲区满时继续发送会等待。 [原文](https://gobyexample.com/channel-buffering) · [源码](30-channel-buffering/main.go)

```go
package main

import (
	"fmt"
)

func main() {
	queue := make(chan string, 2)
	queue <- "早餐"
	queue <- "午餐"
	fmt.Println(len(queue), cap(queue))
	fmt.Println(<-queue)
	fmt.Println(<-queue)
}
```

**运行**

```sh
go run ./30-channel-buffering
```

**输出**

```text
2 2
早餐
午餐
```

<a id="ex-31"></a>

### 31　用通道同步 · Channel Synchronization

后台任务完成后关闭 done。主函数在接收处等待，因此“可以退出”一定在“保存完成”之后打印。 [原文](https://gobyexample.com/channel-synchronization) · [源码](31-channel-synchronization/main.go)

```go
package main

import (
	"fmt"
)

func main() {
	done := make(chan struct{})
	go func() {
		fmt.Println("保存完成")
		close(done)
	}()
	<-done
	fmt.Println("可以退出")
}
```

**运行**

```sh
go run ./31-channel-synchronization
```

**输出**

```text
保存完成
可以退出
```

<a id="ex-32"></a>

### 32　单向通道 · Channel Directions

chan<- 约束函数只能发送，<-chan 约束函数只能接收。双向通道可以传给这两种参数。 [原文](https://gobyexample.com/channel-directions) · [源码](32-channel-directions/main.go)

```go
package main

import (
	"fmt"
)

func publish(out chan<- string) {
	out <- "新书到馆"
}

func receive(in <-chan string) string {
	return <-in
}

func main() {
	news := make(chan string, 1)
	publish(news)
	fmt.Println(receive(news))
}
```

**运行**

```sh
go run ./32-channel-directions
```

**输出**

```text
新书到馆
```

<a id="ex-33"></a>

### 33　等待多个通道 · Select

select 会选择一个已就绪的分支。当两个通道都有数据时，两种接收顺序都有可能，每条消息只会取一次。 [原文](https://gobyexample.com/select) · [源码](33-select/main.go)

```go
package main

import (
	"fmt"
)

func main() {
	mail := make(chan string, 1)
	chat := make(chan string, 1)
	mail <- "收到邮件"
	chat <- "收到消息"
	for range 2 {
		select {
		case msg := <-mail:
			fmt.Println(msg)
		case msg := <-chat:
			fmt.Println(msg)
		}
	}
}
```

**运行**

```sh
go run ./33-select
```

**一种可能的输出，行的顺序可能变化**

```text
收到邮件
收到消息
```

<a id="ex-34"></a>

### 34　超时 · Timeouts

同时等待结果和截止时间。这里让结果通道保持空闲，可以稳定地观察到 time.After 对应的分支。 [原文](https://gobyexample.com/timeouts) · [源码](34-timeouts/main.go)

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	reply := make(chan string)
	select {
	case value := <-reply:
		fmt.Println(value)
	case <-time.After(20 * time.Millisecond):
		fmt.Println("等待回复超时")
	}
}
```

**运行**

```sh
go run ./34-timeouts
```

**输出**

```text
等待回复超时
```

<a id="ex-35"></a>

### 35　非阻塞通道操作 · Non-Blocking Channel Operations

带 default 的 select 在其他分支都无法立即执行时返回。一个满的缓冲通道可以用来演示非阻塞发送。 [原文](https://gobyexample.com/non-blocking-channel-operations) · [源码](35-non-blocking-channel-operations/main.go)

```go
package main

import (
	"fmt"
)

func main() {
	queue := make(chan string, 1)
	queue <- "已排队"
	select {
	case queue <- "新任务":
		fmt.Println("加入成功")
	default:
		fmt.Println("队列已满")
	}
	fmt.Println(<-queue)
}
```

**运行**

```sh
go run ./35-non-blocking-channel-operations
```

**输出**

```text
队列已满
已排队
```

<a id="ex-36"></a>

### 36　关闭通道 · Closing Channels

发送方可以关闭通道。关闭后，接收方先取出已有数据，再次接收会得到该类型的零值以及 false。 [原文](https://gobyexample.com/closing-channels) · [源码](36-closing-channels/main.go)

```go
package main

import (
	"fmt"
)

func main() {
	queue := make(chan int, 1)
	queue <- 7
	close(queue)
	n, ok := <-queue
	fmt.Println(n, ok)
	n, ok = <-queue
	fmt.Println(n, ok)
}
```

**运行**

```sh
go run ./36-closing-channels
```

**输出**

```text
7 true
0 false
```

<a id="ex-37"></a>

### 37　遍历通道 · Range over Channels

for range 会持续接收，直到发送方关闭通道且数据已读完。接收方无需预先知道会收到几个数。 [原文](https://gobyexample.com/range-over-channels) · [源码](37-range-over-channels/main.go)

```go
package main

import (
	"fmt"
)

func main() {
	prices := make(chan int)
	go func() {
		defer close(prices)
		for _, price := range []int{6, 7, 8} {
			prices <- price
		}
	}()
	sum := 0
	for price := range prices {
		sum += price
	}
	fmt.Println(sum)
}
```

**运行**

```sh
go run ./37-range-over-channels
```

**输出**

```text
21
```

## 38–45　定时、任务调度与共享状态

<a id="ex-38"></a>

### 38　一次性定时器 · Timers

time.NewTimer 创建一次性定时器。读取 timer.C 会等到设定的时间过去。尚未触发的定时器也可以主动停止。 [原文](https://gobyexample.com/timers) · [源码](38-timers/main.go)

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	timer := time.NewTimer(10 * time.Millisecond)
	<-timer.C
	fmt.Println("休息结束")

	reminder := time.NewTimer(time.Hour)
	fmt.Println("取消下一次提醒", reminder.Stop())
}
```

**运行**

```sh
go run ./38-timers
```

**输出**

```text
休息结束
取消下一次提醒 true
```

<a id="ex-39"></a>

### 39　周期定时器 · Tickers

Ticker 会按设定的间隔产生通知。这里接收三次后停止，给循环一个明确的结束条件。 [原文](https://gobyexample.com/tickers) · [源码](39-tickers/main.go)

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for n := 1; n <= 3; n++ {
		<-ticker.C
		fmt.Println("检查", n)
	}
}
```

**运行**

```sh
go run ./39-tickers
```

**输出**

```text
检查 1
检查 2
检查 3
```

<a id="ex-40"></a>

### 40　工作池 · Worker Pools

两个 worker 从同一个 jobs 通道取数。发送方关闭 jobs 后，worker 会结束。等它们全部退出，就能关闭结果通道并汇总。 [原文](https://gobyexample.com/worker-pools) · [源码](40-worker-pools/main.go)

```go
package main

import (
	"fmt"
	"sync"
)

func squareWorker(jobs <-chan int, results chan<- int) {
	for n := range jobs {
		results <- n * n
	}
}

func main() {
	jobs := make(chan int)
	results := make(chan int)
	var workers sync.WaitGroup
	for range 2 {
		workers.Go(func() { squareWorker(jobs, results) })
	}
	go func() {
		for n := 1; n <= 4; n++ {
			jobs <- n
		}
		close(jobs)
		workers.Wait()
		close(results)
	}()
	sum := 0
	for n := range results {
		sum += n
	}
	fmt.Println("平方和", sum)
}
```

**运行**

```sh
go run ./40-worker-pools
```

**输出**

```text
平方和 30
```

<a id="ex-41"></a>

### 41　等待一组任务 · WaitGroups

WaitGroup.Go 登记并启动一个任务，Wait 等到所有任务返回。每个任务写入自己的切片位置，等待结束后再读取结果。 [原文](https://gobyexample.com/waitgroups) · [源码](41-waitgroups/main.go)

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	names := []string{"茶", "咖啡", "水"}
	labels := make([]string, len(names))
	var wg sync.WaitGroup
	for i, name := range names {
		wg.Go(func() { labels[i] = "饮品：" + name })
	}
	wg.Wait()
	fmt.Println(labels)
}
```

**运行**

```sh
go run ./41-waitgroups
```

**输出**

```text
[饮品：茶 饮品：咖啡 饮品：水]
```

<a id="ex-42"></a>

### 42　限流 · Rate Limiting

用 ticker 控制队列的处理速度。每收到一次时钟通知，才允许一项任务开始处理。 [原文](https://gobyexample.com/rate-limiting) · [源码](42-rate-limiting/main.go)

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	permits := time.NewTicker(10 * time.Millisecond)
	defer permits.Stop()
	for _, job := range []string{"甲", "乙", "丙"} {
		<-permits.C
		fmt.Println("开始处理", job)
	}
}
```

**运行**

```sh
go run ./42-rate-limiting
```

**输出**

```text
开始处理 甲
开始处理 乙
开始处理 丙
```

<a id="ex-43"></a>

### 43　原子计数器 · Atomic Counters

多个 goroutine 同时累加同一个整数时，atomic.Int64 的 Add 会原子地更新计数，Load 读取结果。 [原文](https://gobyexample.com/atomic-counters) · [源码](43-atomic-counters/main.go)

```go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	var visits atomic.Int64
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			for range 100 {
				visits.Add(1)
			}
		})
	}
	wg.Wait()
	fmt.Println(visits.Load())
}
```

**运行**

```sh
go run ./43-atomic-counters
```

**输出**

```text
400
```

<a id="ex-44"></a>

### 44　互斥锁 · Mutexes

给整个“读取库存、判断、扣减”操作加锁，三个 goroutine 竞争两张票时，只有两个能订到。 [原文](https://gobyexample.com/mutexes) · [源码](44-mutexes/main.go)

```go
package main

import (
	"fmt"
	"sync"
)

type Tickets struct {
	mu   sync.Mutex
	left int
}

func (t *Tickets) Book() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.left == 0 {
		return false
	}
	t.left--
	return true
}

func main() {
	tickets := Tickets{left: 2}
	var wg sync.WaitGroup
	booked := make([]bool, 3)
	for i := range booked {
		wg.Go(func() { booked[i] = tickets.Book() })
	}
	wg.Wait()
	count := 0
	for _, ok := range booked {
		if ok {
			count++
		}
	}
	fmt.Println("成功订票", count)
}
```

**运行**

```sh
go run ./44-mutexes
```

**输出**

```text
成功订票 2
```

<a id="ex-45"></a>

### 45　用 goroutine 管理状态 · Stateful Goroutines

让单个 goroutine 持有余额。调用方通过通道提交变更并接收新余额，每个请求都由这个 goroutine 顺序处理。 [原文](https://gobyexample.com/stateful-goroutines) · [源码](45-stateful-goroutines/main.go)

```go
package main

import (
	"fmt"
)

type Deposit struct {
	Amount int
	Reply  chan int
}

func ledger(requests <-chan Deposit, done chan<- struct{}) {
	defer close(done)
	balance := 0
	for req := range requests {
		balance += req.Amount
		req.Reply <- balance
	}
}

func main() {
	requests := make(chan Deposit)
	done := make(chan struct{})
	go ledger(requests, done)
	reply := make(chan int)
	for _, amount := range []int{10, 7} {
		requests <- Deposit{Amount: amount, Reply: reply}
		fmt.Println(<-reply)
	}
	close(requests)
	<-done
}
```

**运行**

```sh
go run ./45-stateful-goroutines
```

**输出**

```text
10
17
```

## 46–50　排序与异常处理

<a id="ex-46"></a>

### 46　排序 · Sorting

slices.Sort 直接修改切片中的元素。slices.IsSorted 可以检查当前顺序。 [原文](https://gobyexample.com/sorting) · [源码](46-sorting/main.go)

```go
package main

import (
	"fmt"
	"slices"
)

func main() {
	scores := []int{28, 7, 14}
	fmt.Println(slices.IsSorted(scores))
	slices.Sort(scores)
	fmt.Println(scores)
}
```

**运行**

```sh
go run ./46-sorting
```

**输出**

```text
false
[7 14 28]
```

<a id="ex-47"></a>

### 47　自定义排序 · Sorting by Functions

给 slices.SortFunc 提供比较函数。这里按时长升序排列，同样长的任务再按名称排序。 [原文](https://gobyexample.com/sorting-by-functions) · [源码](47-sorting-by-functions/main.go)

```go
package main

import (
	"cmp"
	"fmt"
	"slices"
)

type Task struct {
	Name    string
	Minutes int
}

func main() {
	tasks := []Task{{"read", 20}, {"walk", 10}, {"cook", 20}}
	slices.SortFunc(tasks, func(a, b Task) int {
		if order := cmp.Compare(a.Minutes, b.Minutes); order != 0 {
			return order
		}
		return cmp.Compare(a.Name, b.Name)
	})
	fmt.Println(tasks)
}
```

**运行**

```sh
go run ./47-sorting-by-functions
```

**输出**

```text
[{walk 10} {cook 20} {read 20}]
```

<a id="ex-48"></a>

### 48　触发 panic · Panic

这里故意传入无效的内部配置。panic 会展开当前 goroutine 的调用栈，运行已经注册的 defer。程序预期以非零状态退出。 [原文](https://gobyexample.com/panic) · [源码](48-panic/main.go)

```go
package main

import (
	"fmt"
)

func openBoard(width int) {
	if width <= 0 {
		panic("棋盘宽度必须为正数")
	}
	fmt.Println("创建棋盘", width)
}

func main() {
	defer fmt.Println("清理棋盘资源")
	openBoard(0)
}
```

**运行**

```sh
go run ./48-panic
```

**输出节选，之后还会打印调用栈**

```text
清理棋盘资源
panic: 棋盘宽度必须为正数
```

<a id="ex-49"></a>

### 49　延迟执行 · Defer

defer 注册的调用会在函数返回前逆序执行。用两层收尾动作，可以直接从输出看出顺序。 [原文](https://gobyexample.com/defer) · [源码](49-defer/main.go)

```go
package main

import (
	"fmt"
)

func main() {
	defer fmt.Println("关闭仓库")
	defer fmt.Println("收回钥匙")
	fmt.Println("完成盘点")
}
```

**运行**

```sh
go run ./49-defer
```

**输出**

```text
完成盘点
收回钥匙
关闭仓库
```

<a id="ex-50"></a>

### 50　恢复 panic · Recover

在同一个 goroutine 的延迟函数内直接调用 recover，可以取到 panic 的值。被中断的函数返回后，调用方从下一条语句继续。 [原文](https://gobyexample.com/recover) · [源码](50-recover/main.go)

```go
package main

import (
	"fmt"
)

func runJob() {
	defer func() {
		if reason := recover(); reason != nil {
			fmt.Println("任务失败", reason)
		}
	}()
	panic("缺少配置")
}

func main() {
	runJob()
	fmt.Println("开始下一项任务")
}
```

**运行**

```sh
go run ./50-recover
```

**输出**

```text
任务失败 缺少配置
开始下一项任务
```

## 51–56　文本与数据格式

<a id="ex-51"></a>

### 51　常用字符串函数 · String Functions

用 strings 把一行标签拆开，去掉两端空格，再用新的分隔符连接起来。字符串处理函数会返回新值。 [原文](https://gobyexample.com/string-functions) · [源码](51-string-functions/main.go)

```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	tags := strings.Split(" go, web ,cli ", ",")
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}
	fmt.Println(strings.Join(tags, " / "))
	fmt.Println(strings.HasPrefix(tags[0], "g"))
}
```

**运行**

```sh
go run ./51-string-functions
```

**输出**

```text
go / web / cli
true
```

<a id="ex-52"></a>

### 52　字符串格式化 · String Formatting

用 Sprintf 拼成一行商品信息。%03d 将编号补到三位，%.2f 保留两位小数，%q 会给字符串加引号。 [原文](https://gobyexample.com/string-formatting) · [源码](52-string-formatting/main.go)

```go
package main

import (
	"fmt"
)

func main() {
	label := fmt.Sprintf("商品 %03d | %q | %.2f 元", 7, "茶杯", 12.5)
	fmt.Println(label)
}
```

**运行**

```sh
go run ./52-string-formatting
```

**输出**

```text
商品 007 | "茶杯" | 12.50 元
```

<a id="ex-53"></a>

### 53　文本模板 · Text Templates

模板中的点号代表当前数据。range 让点号依次指向每个切片元素，按同一个格式输出清单。 [原文](https://gobyexample.com/text-templates) · [源码](53-text-templates/main.go)

```go
package main

import (
	"os"
	"text/template"
)

func main() {
	const source = "采购清单\n{{range .}}- {{.}}\n{{end}}"
	tmpl := template.Must(template.New("list").Parse(source))
	if err := tmpl.Execute(os.Stdout, []string{"米", "茶"}); err != nil {
		panic(err)
	}
}
```

**运行**

```sh
go run ./53-text-templates
```

**输出**

```text
采购清单
- 米
- 茶
```

<a id="ex-54"></a>

### 54　正则表达式 · Regular Expressions

先编译一个固定模式，再从文本中取出所有工单号。最后一个参数 -1 表示取出全部匹配。 [原文](https://gobyexample.com/regular-expressions) · [源码](54-regular-expressions/main.go)

```go
package main

import (
	"fmt"
	"regexp"
)

func main() {
	ticket := regexp.MustCompile(`\bT-[0-9]{3}\b`)
	line := "open T-015, close T-207, ignore T-8"
	fmt.Println(ticket.FindAllString(line, -1))
	fmt.Println(ticket.ReplaceAllString(line, "[工单]"))
}
```

**运行**

```sh
go run ./54-regular-expressions
```

**输出**

```text
[T-015 T-207]
open [工单], close [工单], ignore T-8
```

<a id="ex-55"></a>

### 55　JSON 编解码 · JSON

结构体标签指定 JSON 字段名。先把字符串解码成结构体，修改字段，再把修改后的值编码回 JSON。 [原文](https://gobyexample.com/json) · [源码](55-json/main.go)

```go
package main

import (
	"encoding/json"
	"fmt"
)

type Task struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

func main() {
	var task Task
	if err := json.Unmarshal([]byte(`{"title":"读书","done":false}`), &task); err != nil {
		panic(err)
	}
	task.Done = true
	data, err := json.Marshal(task)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}
```

**运行**

```sh
go run ./55-json
```

**输出**

```text
{"title":"读书","done":true}
```

<a id="ex-56"></a>

### 56　XML 编解码 · XML

xml.Name 指定根元素名。带 ,attr 的字段来自属性，普通字段来自子元素。调整数量后再编码，可以看到两者的位置。 [原文](https://gobyexample.com/xml) · [源码](56-xml/main.go)

```go
package main

import (
	"encoding/xml"
	"fmt"
)

type Order struct {
	XMLName xml.Name `xml:"order"`
	ID      string   `xml:"id,attr"`
	Count   int      `xml:"count"`
}

func main() {
	var order Order
	if err := xml.Unmarshal([]byte(`<order id="A7"><count>2</count></order>`), &order); err != nil {
		panic(err)
	}
	order.Count++
	out, err := xml.Marshal(order)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
}
```

**运行**

```sh
go run ./56-xml
```

**输出**

```text
<order id="A7"><count>3</count></order>
```

## 57–64　时间、解析与编码

<a id="ex-57"></a>

### 57　日期、时间与时长 · Time

用一个固定的 UTC 时间安排 45 分钟后的会议结束时间。Add 计算时间点，Sub 算出两者相隔的时长。 [原文](https://gobyexample.com/time) · [源码](57-time/main.go)

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Date(2026, time.April, 3, 9, 30, 0, 0, time.UTC)
	end := start.Add(45 * time.Minute)
	fmt.Println(end.Format("15:04"))
	fmt.Println(end.Sub(start))
}
```

**运行**

```sh
go run ./57-time
```

**输出**

```text
10:15
45m0s
```

<a id="ex-58"></a>

### 58　Unix 时间戳 · Epoch

Unix 时间戳以 1970 年 UTC 元旦为起点。比较同一时刻的秒数和毫秒数，再用 time.Unix 把秒数还原。 [原文](https://gobyexample.com/epoch) · [源码](58-epoch/main.go)

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	moment := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	fmt.Println(moment.Unix())
	fmt.Println(moment.UnixMilli())
	fmt.Println(time.Unix(moment.Unix(), 0).UTC().Format(time.RFC3339))
}
```

**运行**

```sh
go run ./58-epoch
```

**输出**

```text
1704067200
1704067200000
2024-01-01T00:00:00Z
```

<a id="ex-59"></a>

### 59　时间格式化与解析 · Time Formatting / Parsing

Go 以参考时间 2006-01-02 15:04:05 中的数字表示年月日和时分秒。使用 ParseInLocation 可以指定无时区输入的解释方式。 [原文](https://gobyexample.com/time-formatting-parsing) · [源码](59-time-formatting-parsing/main.go)

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	zone := time.FixedZone("UTC+8", 8*60*60)
	meeting, err := time.ParseInLocation("2006/01/02 15:04", "2026/04/03 09:30", zone)
	if err != nil {
		panic(err)
	}
	fmt.Println(meeting.Format(time.RFC3339))
	fmt.Println(meeting.UTC().Format(time.RFC3339))
}
```

**运行**

```sh
go run ./59-time-formatting-parsing
```

**输出**

```text
2026-04-03T09:30:00+08:00
2026-04-03T01:30:00Z
```

<a id="ex-60"></a>

### 60　随机数 · Random Numbers

给局部伪随机数生成器一个固定种子，可以重复得到相同序列，方便调试。IntN(6) 的取值范围是 0 到 5。 [原文](https://gobyexample.com/random-numbers) · [源码](60-random-numbers/main.go)

```go
package main

import (
	"fmt"
	"math/rand/v2"
)

func main() {
	dice := rand.New(rand.NewPCG(12, 34))
	for range 3 {
		fmt.Println(dice.IntN(6) + 1)
	}
}
```

**运行**

```sh
go run ./60-random-numbers
```

**输出**

```text
5
4
4
```

<a id="ex-61"></a>

### 61　解析数字 · Number Parsing

Atoi 读取十进制整数。ParseInt 的进制参数为 0 时会根据前缀识别进制。遇到非法字符时，可以检查返回的错误。 [原文](https://gobyexample.com/number-parsing) · [源码](61-number-parsing/main.go)

```go
package main

import (
	"fmt"
	"strconv"
)

func main() {
	n, err := strconv.Atoi("42")
	if err != nil {
		panic(err)
	}
	hex, err := strconv.ParseInt("0x2a", 0, 64)
	if err != nil {
		panic(err)
	}
	_, invalid := strconv.Atoi("4x")
	fmt.Println(n, hex, invalid != nil)
}
```

**运行**

```sh
go run ./61-number-parsing
```

**输出**

```text
42 42 true
```

<a id="ex-62"></a>

### 62　解析 URL · URL Parsing

url.Parse 把地址拆成各个部分。用 Hostname 去掉端口，用 Query 获取已解码的查询参数。 [原文](https://gobyexample.com/url-parsing) · [源码](62-url-parsing/main.go)

```go
package main

import (
	"fmt"
	"net/url"
)

func main() {
	u, err := url.Parse("https://example.org:8443/books?q=Go+Web")
	if err != nil {
		panic(err)
	}
	fmt.Println(u.Scheme, u.Hostname(), u.Port())
	fmt.Println(u.Path, u.Query().Get("q"))
}
```

**运行**

```sh
go run ./62-url-parsing
```

**输出**

```text
https example.org 8443
/books Go Web
```

<a id="ex-63"></a>

### 63　SHA-256 摘要 · SHA256 Hashes

Sum256 计算输入字节的固定长度摘要。用 %x 把 32 个字节打印成 64 个十六进制字符。 [原文](https://gobyexample.com/sha256-hashes) · [源码](63-sha256-hashes/main.go)

```go
package main

import (
	"crypto/sha256"
	"fmt"
)

func main() {
	data := []byte("go study notes")
	digest := sha256.Sum256(data)
	fmt.Println(len(digest))
	fmt.Printf("%x\n", digest)
}
```

**运行**

```sh
go run ./63-sha256-hashes
```

**输出**

```text
32
9d67d6aa8d3a2dc249a39295887caf46d9937ab289d2c77e475a437efc773fa6
```

<a id="ex-64"></a>

### 64　Base64 编解码 · Base64 Encoding

把 UTF-8 字节编码成可打印的文本，再解码回原字符串。Base64 可以直接还原原文。 [原文](https://gobyexample.com/base64-encoding) · [源码](64-base64-encoding/main.go)

```go
package main

import (
	"encoding/base64"
	"fmt"
)

func main() {
	encoded := base64.StdEncoding.EncodeToString([]byte("一杯茶"))
	fmt.Println(encoded)
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(decoded))
}
```

**运行**

```sh
go run ./64-base64-encoding
```

**输出**

```text
5LiA5p2v6Iy2
一杯茶
```

## 65–72　文件与测试

<a id="ex-65"></a>

### 65　读取文件 · Reading Files

os.ReadFile 适合一次读入小文件。示例附带一份书单，从该示例目录运行即可直接读取。 [原文](https://gobyexample.com/reading-files) · [源码](65-reading-files/main.go)

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	data, err := os.ReadFile("books.txt")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(data))
}
```

示例目录里的 [books.txt](65-reading-files/books.txt) 包含以下内容。

```text
Go 入门
并发笔记
```

**运行**

```sh
(cd 65-reading-files && go run .)
```

**输出**

```text
Go 入门
并发笔记
```

<a id="ex-66"></a>

### 66　写入文件 · Writing Files

os.WriteFile 可以创建或覆盖文件。这里写进新建的临时目录，读回内容作为检查，并在结束时清理。 [原文](https://gobyexample.com/writing-files) · [源码](66-writing-files/main.go)

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	dir, err := os.MkdirTemp("", "go-note-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "note.txt")
	if err := os.WriteFile(path, []byte("明天读十页\n"), 0600); err != nil {
		panic(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(data))
}
```

**运行**

```sh
go run ./66-writing-files
```

**输出**

```text
明天读十页
```

<a id="ex-67"></a>

### 67　逐行过滤 · Line Filters

Scanner 从标准输入逐行读取。只保留带 TODO 前缀的行，扫描结束后检查是否发生读取错误。 [原文](https://gobyexample.com/line-filters) · [源码](67-line-filters/main.go)

```go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "TODO ") {
			fmt.Println(strings.TrimPrefix(line, "TODO "))
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
```

**运行**

```sh
printf 'TODO 买茶\nDONE 洗碗\nTODO 读书\n' | go run ./67-line-filters
```

**输出**

```text
买茶
读书
```

<a id="ex-68"></a>

### 68　文件路径 · File Paths

filepath.Join 使用当前系统的路径规则拼接目录和文件名。Base 和 Ext 可以分别提取最后一段和扩展名。 [原文](https://gobyexample.com/file-paths) · [源码](68-file-paths/main.go)

```go
package main

import (
	"fmt"
	"path/filepath"
)

func main() {
	path := filepath.Join("notes", "go", "day1.md")
	fmt.Println(path)
	fmt.Println(filepath.Base(path))
	fmt.Println(filepath.Ext(path))
}
```

**运行**

```sh
go run ./68-file-paths
```

**macOS / Linux 上的输出**

```text
notes/go/day1.md
day1.md
.md
```

<a id="ex-69"></a>

### 69　目录操作 · Directories

用 MkdirAll 创建嵌套目录，再用 ReadDir 列出其直接子项。每项的 IsDir 可以区分目录和普通文件。 [原文](https://gobyexample.com/directories) · [源码](69-directories/main.go)

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	root, err := os.MkdirTemp("", "go-library-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(root)
	if err := os.MkdirAll(filepath.Join(root, "go", "examples"), 0700); err != nil {
		panic(err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		fmt.Println(entry.Name(), entry.IsDir())
	}
}
```

**运行**

```sh
go run ./69-directories
```

**输出**

```text
go true
```

<a id="ex-70"></a>

### 70　临时文件与目录 · Temporary Files and Directories

MkdirTemp 创建独立的工作目录，CreateTemp 在其中生成唯一文件名。关闭文件后，函数返回时会一并清理目录。 [原文](https://gobyexample.com/temporary-files-and-directories) · [源码](70-temporary-files-and-directories/main.go)

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	dir, err := os.MkdirTemp("", "go-draft-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	file, err := os.CreateTemp(dir, "draft-*.txt")
	if err != nil {
		panic(err)
	}
	if err := file.Close(); err != nil {
		panic(err)
	}
	info, err := os.Stat(file.Name())
	if err != nil {
		panic(err)
	}
	fmt.Println("临时文件已创建，字节数", info.Size())
}
```

**运行**

```sh
go run ./70-temporary-files-and-directories
```

**输出**

```text
临时文件已创建，字节数 0
```

<a id="ex-71"></a>

### 71　嵌入文件 · Embed Directive

//go:embed 让编译器把文件内容放进程序。这里把同目录的 menu.txt 嵌入字符串，运行时可以直接打印。 [原文](https://gobyexample.com/embed-directive) · [源码](71-embed-directive/main.go)

```go
package main

import (
	_ "embed"
	"fmt"
)

//go:embed menu.txt
var menu string

func main() {
	fmt.Print(menu)
}
```

示例目录里的 [menu.txt](71-embed-directive/menu.txt) 包含以下内容。

```text
今日菜单
- 茶
- 米饭
```

**运行**

```sh
go run ./71-embed-directive
```

**输出**

```text
今日菜单
- 茶
- 米饭
```

<a id="ex-72"></a>

### 72　单元测试与基准测试 · Testing and Benchmarking

TestTotal 用几个输入检查返回值，失败时会报告差异。BenchmarkTotal 用 b.Loop 重复计时，观察一次调用平均耗时。 [原文](https://gobyexample.com/testing-and-benchmarking) · [源码](72-testing-and-benchmarking/total_test.go)

```go
package main

import (
	"testing"
)

func total(prices []int) int {
	sum := 0
	for _, price := range prices {
		sum += price
	}
	return sum
}

func TestTotal(t *testing.T) {
	cases := []struct {
		prices []int
		want   int
	}{{nil, 0}, {[]int{2, 5}, 7}, {[]int{4, -1}, 3}}
	for _, tc := range cases {
		if got := total(tc.prices); got != tc.want {
			t.Errorf("total(%v) = %d, want %d", tc.prices, got, tc.want)
		}
	}
}

func BenchmarkTotal(b *testing.B) {
	prices := []int{2, 5, 8}
	for b.Loop() {
		total(prices)
	}
}
```

**运行**

```sh
go test -v -bench . -benchtime=100ms ./72-testing-and-benchmarking
```

**一次运行的节选，基准计时随机器和运行情况变化**

```text
--- PASS: TestTotal (0.00s)
BenchmarkTotal-18    	54133512	         1.992 ns/op
PASS
```

## 73–77　命令行与日志

<a id="ex-73"></a>

### 73　命令行位置参数 · Command-Line Arguments

os.Args[0] 是程序名，os.Args[1:] 是调用者提供的参数。试着把带空格的标题放进引号，观察它占几个位置。 [原文](https://gobyexample.com/command-line-arguments) · [源码](73-command-line-arguments/main.go)

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	for i, arg := range os.Args[1:] {
		fmt.Printf("第 %d 个参数 = %q\n", i+1, arg)
	}
}
```

**运行**

```sh
go run ./73-command-line-arguments add "Go 练习"
```

**输出**

```text
第 1 个参数 = "add"
第 2 个参数 = "Go 练习"
```

<a id="ex-74"></a>

### 74　命令行选项 · Command-Line Flags

flag 包提供选项解析和默认值。Parse 后读取指针指向的值，flag.Args 则给出剩余的位置参数。 [原文](https://gobyexample.com/command-line-flags) · [源码](74-command-line-flags/main.go)

```go
package main

import (
	"flag"
	"fmt"
)

func main() {
	count := flag.Int("n", 1, "打印次数")
	flag.Parse()
	for range *count {
		fmt.Println(flag.Args())
	}
}
```

**运行**

```sh
go run ./74-command-line-flags -n 2 读书
```

**输出**

```text
[读书]
[读书]
```

<a id="ex-75"></a>

### 75　命令行子命令 · Command-Line Subcommands

先判断第一个位置参数，再用独立的 FlagSet 解析该子命令的选项。示例支持 add 和 list 两个子命令。 [原文](https://gobyexample.com/command-line-subcommands) · [源码](75-command-line-subcommands/main.go)

```go
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "用法：tasks add -title 名称 | tasks list")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "add":
		add := flag.NewFlagSet("add", flag.ExitOnError)
		title := add.String("title", "", "待办名称")
		add.Parse(os.Args[2:])
		if *title == "" {
			fmt.Fprintln(os.Stderr, "请提供 -title")
			os.Exit(2)
		}
		fmt.Println("收到新任务", *title)
	case "list":
		fmt.Println("当前任务清单为空")
	default:
		fmt.Fprintln(os.Stderr, "未知子命令", os.Args[1])
		os.Exit(2)
	}
}
```

**运行**

```sh
go run ./75-command-line-subcommands add -title 写笔记
```

**输出**

```text
收到新任务 写笔记
```

<a id="ex-76"></a>

### 76　环境变量 · Environment Variables

LookupEnv 同时返回值和是否设置。命令中临时设置的 STUDY_TOPIC 只传给这次启动的程序。 [原文](https://gobyexample.com/environment-variables) · [源码](76-environment-variables/main.go)

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	topic, ok := os.LookupEnv("STUDY_TOPIC")
	if !ok {
		topic = "自由练习"
	}
	fmt.Println("今日主题", topic)
}
```

**运行**

```sh
STUDY_TOPIC=通道 go run ./76-environment-variables
```

**输出**

```text
今日主题 通道
```

<a id="ex-77"></a>

### 77　日志 · Logging

用标准库的 slog 给日志添加可查询的字段。移除处理器的默认时间字段，便于重复运行时核对示例输出。 [原文](https://gobyexample.com/logging) · [源码](77-logging/main.go)

```go
package main

import (
	"log/slog"
	"os"
)

func main() {
	options := &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if len(groups) == 0 && attr.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return attr
		},
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, options))
	logger.Info("借书完成", "book", "Go", "days", 7)
}
```

**运行**

```sh
go run ./77-logging
```

**输出**

```text
level=INFO msg=借书完成 book=Go days=7
```

## 78–85　网络、上下文与进程

<a id="ex-78"></a>

### 78　HTTP 客户端 · HTTP Client

启动一个本地测试服务，再用带超时的客户端请求它。拿到响应后要关闭响应体，HTTP 状态码可以从 resp.StatusCode 读取。 [原文](https://gobyexample.com/http-client) · [源码](78-http-client/main.go)

```go
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"
)

func main() {
	local := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "今日有茶")
	}))
	defer local.Close()

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(local.URL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.StatusCode, string(body))
}
```

**运行**

```sh
go run ./78-http-client
```

**输出**

```text
200 今日有茶
```

<a id="ex-79"></a>

### 79　HTTP 服务端 · HTTP Server

用 ServeMux 把特定 HTTP 方法和路径映射到处理函数。本地服务监听临时端口，示例自己发起请求，展示路径参数和响应内容。 [原文](https://gobyexample.com/http-server) · [源码](79-http-server/main.go)

```go
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /books/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "书籍编号 ", r.PathValue("id"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/books/42")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
		panic(err)
	}
	fmt.Println()
}
```

**运行**

```sh
go run ./79-http-server
```

**输出**

```text
书籍编号 42
```

<a id="ex-80"></a>

### 80　TCP 服务端 · TCP Server

监听本机随机端口，接受一个 TCP 连接。客户端发送一行文字，服务端用换行符识别消息边界，然后回送确认。 [原文](https://gobyexample.com/tcp-server) · [源码](80-tcp-server/main.go)

```go
package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func serveOne(listener net.Listener) error {
	conn, err := listener.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(conn, "收到", strings.TrimSpace(line))
	return err
}

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	done := make(chan error, 1)
	go func() { done <- serveOne(listener) }()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	if _, err := fmt.Fprintln(conn, "你好"); err != nil {
		panic(err)
	}
	reply, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		panic(err)
	}
	fmt.Print(reply)
	if err := <-done; err != nil {
		panic(err)
	}
}
```

**运行**

```sh
go run ./80-tcp-server
```

**输出**

```text
收到 你好
```

<a id="ex-81"></a>

### 81　取消与截止时间 · Context

调用方通过 Context 给工作设置截止时间。工作函数在等待输入的同时监听 ctx.Done()，并把取消原因传回调用方。 [原文](https://gobyexample.com/context) · [源码](81-context/main.go)

```go
package main

import (
	"context"
	"fmt"
	"time"
)

func take(ctx context.Context, jobs <-chan string) (string, error) {
	select {
	case job := <-jobs:
		return job, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := take(ctx, make(chan string))
	fmt.Println(err)
}
```

**运行**

```sh
go run ./81-context
```

**输出**

```text
context deadline exceeded
```

<a id="ex-82"></a>

### 82　启动子进程 · Spawning Processes

exec.Command 启动外部程序。Output 等它完成并捕获标准输出，之后当前 Go 程序会继续执行。 [原文](https://gobyexample.com/spawning-processes) · [源码](82-spawning-processes/main.go)

```go
package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func main() {
	out, err := exec.Command("go", "version").Output()
	if err != nil {
		panic(err)
	}
	fmt.Println(strings.TrimSpace(string(out)))
	fmt.Println("子进程执行完毕")
}
```

**运行**

```sh
go run ./82-spawning-processes
```

**本机输出，Go 版本和平台因环境而异**

```text
go version go1.27.3 darwin/arm64
子进程执行完毕
```

<a id="ex-83"></a>

### 83　替换当前进程 · Exec'ing Processes

在 macOS 或 Linux 上，syscall.Exec 可以用另一个程序替换当前进程。成功后新程序从自己的入口运行，旧程序的 defer 也不会执行。 [原文](https://gobyexample.com/execing-processes) · [源码](83-execing-processes/main.go)

```go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	path, err := exec.LookPath("echo")
	if err != nil {
		panic(err)
	}
	defer fmt.Println("旧程序执行收尾")
	if err := syscall.Exec(path, []string{"echo", "新程序开始运行"}, os.Environ()); err != nil {
		panic(err)
	}
}
```

**运行**

```sh
go run ./83-execing-processes
```

**输出**

```text
新程序开始运行
```

<a id="ex-84"></a>

### 84　接收系统信号 · Signals

运行后按一次 Ctrl+C。NotifyContext 收到中断信号会取消上下文，程序可以在退出前完成收尾。 [原文](https://gobyexample.com/signals) · [源码](84-signals/main.go)

```go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	fmt.Println("正在等待中断信号，请按 Ctrl+C")
	<-ctx.Done()
	fmt.Println("收到中断，清理完成")
}
```

**运行**

```sh
go run ./84-signals
```

**运行后在终端按一次 Ctrl+C**

```text
正在等待中断信号，请按 Ctrl+C
收到中断，清理完成
```

<a id="ex-85"></a>

### 85　退出状态码 · Exit

os.Exit 会立即以给定的状态码结束进程，已经注册的 defer 也不会运行。这个程序故意返回 3，方便用 shell 查看。 [原文](https://gobyexample.com/exit) · [源码](85-exit/main.go)

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	defer fmt.Println("执行延迟清理")
	fmt.Println("任务未完成")
	os.Exit(3)
}
```

**运行**

```sh
demo_dir=$(mktemp -d)
go build -o "$demo_dir/exit-demo" ./85-exit
"$demo_dir/exit-demo"
echo "$?"
rm -f "$demo_dir/exit-demo"
rmdir "$demo_dir"
```

**程序输出及 shell 读取的退出码**

```text
任务未完成
3
```
