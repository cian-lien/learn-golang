# Go 语言基础复习

`base/` 目录既是复习正文，也是一套可运行的示例代码：`01_vars` 到 `11_testing` 每个目录一个主题，文中出现的输出都是在模块根目录（`go/`）实际运行的结果（Go 1.27.1 / darwin arm64）。

```bash
cd ~/development/learning/go   # 模块根目录（base/ 的上一级）
go run ./base/01_vars               # 逐个主题运行
go vet ./...                   # 静态检查
gofmt -l .                     # 列出格式不规范的文件（无输出即全部合规）
go test ./base/11_testing -v        # 跑测试
```

## 目录

1. [变量、常量与零值](#1-变量常量与零值)
2. [类型系统与字符串](#2-类型系统与字符串)
3. [控制流](#3-控制流)
4. [函数](#4-函数)
5. [数组、切片与 map](#5-数组切片与-map)
6. [结构体与方法](#6-结构体与方法)
7. [接口、类型断言与错误处理](#7-接口类型断言与错误处理)
8. [指针](#8-指针)
9. [并发](#9-并发)
10. [泛型](#10-泛型)
11. [常用标准库](#11-常用标准库)
12. [测试与工具链](#12-测试与工具链)
13. [易错点速查](#13-易错点速查)
14. [自测清单](#14-自测清单)

## 1. 变量、常量与零值

```go
var a int = 10        // 显式类型
var b = 20            // 类型推断
var c, d = 1, "x"     // 一行多个
e := 30               // 短声明，只能用在函数内

var i int             // 零值 0
var s string          // 零值 ""
var sl []int          // 零值 nil，len(sl) == 0
var m map[string]int  // 零值 nil
```

要点：

- 短声明 `:=` **只能在函数体内**使用，且左侧至少要有一个新变量；`=` 用于已有变量赋值。
- 变量声明后一定有意义明确的零值：数值 `0`、字符串 `""`、布尔 `false`、指针/接口/slice/map/channel/函数 `nil`。Go 里不存在"未初始化变量"。
- `nil` slice 可以直接 `append`、`len`、`range`；`nil` map 可以读，但**写入会 panic**（见第 5 节）。
- 包级变量不能用 `:=`。

常量：

```go
const Pi = 3.14159

const (
    StatusOK       = iota // 0
    StatusNotFound        // 1
    StatusError           // 2
)

const big = 1 << 40
var x float64 = big        // 无类型常量可以按上下文转换，不报错
```

- `iota` 在每个 `const` 块内从 0 开始，每行 +1；省略右侧表达式时沿用上一行的表达式。
- 常量是**无类型的**，只要值能精确表示就能赋给任意兼容类型；`const` 必须有编译期可确定的值。

实际输出（`go run ./base/01_vars`）：

```
10 20 1 x 30 hello true
零值: int=0 string="" pointer=<nil> slice=[] map=map[]
func 零值是不是 nil: true
空 slice 是否 nil: true, len=0
iota: 0 1 2 Pi = 3.14159
无类型常量赋给 float64: 1.099511627776e+12
交换后: 20 10
```

## 2. 类型系统与字符串

Go 是**静态强类型**语言，不同类型之间不会自动转换，必须显式写 `T(v)`。

| 类别 | 常用类型 |
| --- | --- |
| 整数 | `int` `int8/16/32/64` `uint` `uint8` `uintptr` |
| 浮点/复数 | `float32` `float64` `complex64` `complex128` |
| 其他 | `bool` `string` `byte`(= `uint8`) `rune`(= `int32`) |

```go
var i int = 42
var f float64 = float64(i) // 必须显式转换
var u uint = uint(f)

var i8 int8 = 127
i8++ // -128：运行期回绕，不 panic
```

- `int` 的宽度与平台相关（64 位机上 64 位），跨平台存储/序列化时用明确宽度的类型。
- 常量参与运算时会在编译期检查溢出，变量运算则是运行期回绕。
- 整数除法会截断：`1/2 == 0`，要小数结果至少让一个操作数是浮点（`1.0/2`）。

字符串：

- `string` 是**只读字节序列**，按 UTF-8 编码存储，`len(s)` 是字节数而不是字符数。
- 索引 `s[i]` 得到 `byte`；`for i, r := range s` 按 UTF-8 解码得到 `rune`，`i` 是字节下标（中文会跳号）。
- 需要按字符处理时转成 `[]rune`；反引号 `` `...` `` 是原始字符串，不处理转义。

实际输出（`go run ./base/02_types`）：

```
int8 溢出回绕: -128 | uint8: 255
int=42 float64=42 uint=42
字节长度: 8 | 字符个数: 4
按字节的十六进制: 47 6f e8 af ad e8 a8 80
  index=0 rune=G code=U+0047
  index=1 rune=o code=U+006F
  index=2 rune=语 code=U+8BED
  index=5 rune=言 code=U+8A00
原始字符串: 第一行\n第二行
rune 切片: [71 111 35821 35328] | 还原: Go语言
拼接: Go-语言
```

## 3. 控制流

Go 只有 `for` 一个循环关键字，`if` / `switch` / `for` 都可以带初始化语句，`switch` 默认 **不穿透**。

```go
if v := 10; v > 5 { ... }              // v 只在 if/else 内可见

for i := 0; i < 5; i++ { }              // 三段式
for n < 3 { n++ }                       // 相当于 while
for { if c == 3 { break } }             // 无限循环
for i := range 3 { }                    // Go 1.22+：range over int（0,1,2）
for i, v := range slice { }             // range 切片/map/字符串/channel

switch x := 7; {                        // 无表达式 switch，等价于 if/else 链
case x < 5:
default:
}

switch 3 {
case 1, 2:       // 一个 case 可匹配多个值
case 3:
    fallthrough  // 显式穿透到下一个 case
case 4:
}
```

- `continue` / `break` 可以配合标签跳出多层循环：`outer: for {... continue outer }`。
- `range` 遍历 map 的顺序是随机的，需要顺序输出就自己排序（示例见第 5 节）。
- `range` 字符串返回的是 `rune`，`range` 数组/切片返回的是索引和元素副本。
- Go 1.22 起每次迭代都是新的循环变量，闭包里捕获 `i` 不再需要 `i := i`（老代码里常见这个写法）。

实际输出（`go run ./base/03_control`）：

```
if 初始化语句: 10
经典 for 求和: 10
while 风格: 3
无限循环 + break: 3
0 1 2
range 0=a range 1=b
switch 无表达式: 中
命中 3，接着 fallthrough
fallthrough 落到这里
0 0 1 0 2 0
goto 结束: 2
```

## 4. 函数

```go
func divmod(a, b int) (q, r int) { // 多返回值 + 命名返回值
    q, r = a/b, a%b
    return // 裸 return
}

func sum(nums ...int) int          // 变参，调用时 sum(1,2,3) 或 sum(s...)

func apply(x int, f func(int) int) int  // 函数是一等公民，可作参数/返回值

func counter() func() int {        // 闭包：捕获并延长外部变量生命周期
    n := 0
    return func() int { n++; return n }
}
```

`defer` 是 Go 的资源清理主力：

- **LIFO**：多个 `defer` 逆序执行；`defer f(x)` 的参数在**注册时**求值，函数体在返回时执行。
- 常用来 `defer file.Close()`、`defer mu.Unlock()`、`defer wg.Done()`。
- `defer` 里可以修改**命名返回值**，因此常见"给 error 补上下文"的写法：`defer func(){ if err != nil { err = fmt.Errorf("...: %w", err) } }()`。

实际输出（`go run ./base/04_funcs`）：

```
divmod: 2 1
变参: 6 0 9
函数作为参数: 25
闭包计数: 1 2 3
deferDemo 函数体结束
  defer 2
  defer 1
  defer 0
命名返回值 + defer: 包装后: 原始错误
立即执行函数: 来自匿名函数
```

## 5. 数组、切片与 map

三者区别：

| | 长度 | 赋值语义 | 可否比较 |
| --- | --- | --- | --- |
| 数组 `[3]int` | 类型的一部分 | 整体复制 | 可（元素可比较时） |
| 切片 `[]int` | 动态 | 复制 slice header，**共享底层数组** | 只能与 `nil` 比较 |
| map `map[K]V` | 动态 | 引用语义 | 只能与 `nil` 比较 |

切片三件套：

```go
s := make([]int, 0, 4) // len=0, cap=4
s = append(s, 1)       // append 返回新切片，必须接住
s[1:3] / s[:2] / s[3:] // 切片表达式共享底层数组
copy(dst, src)         // 独立副本
```

两个高频陷阱：

1. **视图共享底层数组**：`view := base[1:3]; view[0] = 99` 会改到 `base`。想要独立副本就 `make` + `copy`。
2. **切片传参**：函数内 `in[0] = -1` 会影响调用方（同一底层数组），但 `in = append(in, x)` 一旦扩容就换了数组，调用方看不到新元素。需要"追加结果"就让函数**返回**新切片。

map 要点：

```go
v, ok := m["k"]    // comma-ok 判断 key 是否存在
delete(m, "k")     // 删除
for k := range m   // 遍历顺序随机
```

- `nil` map 读安全（返回零值），写会 panic；`nil` map 也能 `len`、`range`、`delete`。
- 取不存在的 key 得到的是 value 类型的零值，所以"值可能是零值"的场景必须用 comma-ok。
- map 不是并发安全的，多 goroutine 读写需要 `sync.Mutex` / `sync.Map`。

实际输出（`go run ./base/05_slice_map`）：

```
数组: [0 0 0] [1 2 3] [4 5] | 复制不影响原数组: 1
初始 len/cap: 0 4
  append 0 → len=1 cap=4
  ...（省略 append 1/2/3）
  append 4 → len=5 cap=8
切片: [0 1 2 3 4] | s[1:3]: [1 2] | s[:2]: [0 1] | s[3:]: [3 4]
共享底层数组 → base: [1 99 3 4] view: [99 3]
copy 之后互不影响 → base: [1 99 3 4] safe: [0 3]
切片传参 → xs: [-1 2 3]
nil map 读取: 0 | len: 0 | 是否 nil: true
map: map[a:2 b:2] | len: 2
comma-ok → 0 false
delete 之后: map[b:2]
```

> `append` 的扩容策略：容量不足时通常翻倍（小切片），所以示例里 `cap` 从 4 直接到 8。

## 6. 结构体与方法

```go
type Point struct{ X, Y int }

func (p Point) MoveByValue(dx int) Point { p.X += dx; return p } // 副本
func (p *Point) MoveByPointer(dx int)    { p.X += dx }           // 改原件

p := Point{1, 2}
p2 := p.MoveByValue(10)  // p 不变，p2 = {11 2}
p.MoveByPointer(10)      // p = {11 2}
```

- **值接收者 vs 指针接收者**：需要修改接收者、或结构体较大时用指针接收者；同一类型的方法集尽量统一风格。
- 值类型变量调用指针方法时 Go 会自动取地址（`p.MoveByPointer()`），指针调用值方法会自动解引用。
- 结构体字段都可比较时才能用 `==`（含 slice/map/func 字段的结构体不可比较）。
- **嵌入**（匿名字段）实现组合与方法提升，外层同名方法会覆盖被嵌入类型的方法：

```go
type Dog struct {
    Animal        // 匿名字段，d.Name、d.Speak() 直接可用
    Breed string
}
d.Animal.Speak() // 显式调用被提升/被覆盖的方法
```

- 字段标签用于序列化（`omitempty` 表示零值时省略该字段）：

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email,omitempty"`
}
```

- 结构体字面量推荐写字段名（`Point{X: 7}`），新增字段时不易错位。

实际输出（`go run ./base/06_struct_method`）：

```
值接收者: {1 2} {11 2}
指针接收者: {11 2}
结构体可比较: true false
嵌入字段提升: 旺财
方法提升: 旺财 发出声音 | 外层同名方法覆盖: 旺财 汪汪
组合: 旺财(柴犬)
指针访问: {50 100} | pp.X = 50
字面量: {7 0} {0 0}
```

## 7. 接口、类型断言与错误处理

接口是**隐式实现**的：只要方法集满足，类型就实现了接口，不需要 `implements` 声明。

```go
type Shape interface{ Area() float64 }

var s Shape = Rect{3, 4}    // 赋值即隐式实现
fmt.Printf("%T", s)         // 动态类型 main.Rect

v, ok := any1.(string)      // 类型断言，ok=false 时 v 是零值

switch t := v.(type) {      // type switch
case int:  ...
case nil:  ...
default:   ...
}
```

- 接口变量 = (动态类型, 动态值)。接口方法集由动态类型决定，`%T` 打印的是动态类型。
- `any`（= `interface{}`）可装任何值，取出来必须断言或 switch。
- **nil 陷阱**：`var p *MyErr; var e error = p` 时 `e != nil`，因为接口里存了类型信息。函数返回接口时不要返回具体类型的 nil 指针。

错误处理：

```go
var ErrNotFound = errors.New("not found")          // 哨兵错误
fmt.Errorf("非法 id %d: %w", id, ErrNotFound)      // %w 包装，保留错误链
errors.Is(err, ErrNotFound)                        // 链上是否有这个错误
errors.As(err, &myErr)                             // 链上能否取到某具体类型
errors.Unwrap(err)                                 // 拆一层

if err != nil { return fmt.Errorf("读配置: %w", err) } // 始终补上下文
```

- `%w` 包装，`%v` 只拼字符串（断链）。
- 约定：可预期的失败用 `error` 返回；`panic` 只用于程序逻辑上"不该发生"的情况，跨包边界不要用 panic 传递错误。
- `recover` 只在 `defer` 中有效，一般用于顶层兜底（如 HTTP 中间件）。

实际输出（`go run ./base/07_interface_error`，节选）：

```
main.Rect 面积=12.00
main.Circle 面积=3.14
断言成功: hello
断言失败 → n = 0 , ok = false
  type switch: 其他 []int [1]
哨兵错误: not found | errors.Is: true
包装错误: 非法 id -5: not found
  errors.Is 仍然命中: true | Unwrap: not found
errors.As 取出 code: 404
接口 nil 判断 → e == nil: true | error(nilPtr) == nil: false
```

## 8. 指针

```go
p := Point{1, 2}
pp := &p           // 取地址
pp.Y = 100         // 自动解引用，等价于 (*pp).Y = 100
(*pp).X = 50

var np *int        // 零值 nil
fmt.Println(np == nil)
```

- Go 有指针但**没有指针运算**，不能靠 `p++` 跨类型访问内存。
- 函数参数一律按值传递：想修改调用方的变量就传指针（`*T`）。
- 结构体方法用指针接收者时，值变量调用会自动取地址；但接口的方法集不同——值类型的方法集只含值方法，`*T` 的方法集含值方法 + 指针方法。**把只有指针方法的类型的值赋给接口会编译报错**。
- 用指针表达"可选"（`*string` 能区分"没给"和"给了空串"），但注意 JSON 里只有 `nil` 指针才会被 `omitempty` 省略。
- 使用前判空，尤其在调用可能返回 `nil` 的函数之后。

## 9. 并发

Go 的并发模型：**goroutine + channel**（"不要通过共享内存来通信，而要通过通信来共享内存"）。

```go
go doWork()                       // 启动 goroutine

ch := make(chan int)              // 无缓冲：收发必须配对，会阻塞
buf := make(chan int, 3)          // 有缓冲：满/空才阻塞
ch <- 1                           // 发送
v, ok := <-ch                     // 接收；closed 后 ok=false，v 是零值
close(ch)                         // 由发送方关闭，接收方用 range 读完

select {
case v := <-ch:
case <-time.After(time.Second):    // 超时
default:                           // 无阻塞时的兜底
}
```

同步原语：

```go
var wg sync.WaitGroup             // 等一组 goroutine 结束
wg.Add(1); go func(){ defer wg.Done(); ... }(); wg.Wait()

var mu sync.Mutex                 // 保护共享变量
mu.Lock(); total++; mu.Unlock()

var once sync.Once                // 只执行一次
once.Do(func(){ ... })

ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel()                    // 超时/取消/传递请求作用域
<-ctx.Done(); ctx.Err()
```

必记的坑：

- **数据竞争**：多 goroutine 同时读写同一变量必须加锁或用 channel，用 `go test -race` / `go run -race` 检查。
- **channel 只由发送方关闭**，向已关闭的 channel 发送会 panic，重复 close 也会 panic。
- **goroutine 泄漏**：向无人接收的 channel 发送会永久阻塞；用 `context` 或 `select + done` 保证能退出。
- 无缓冲 channel 的收发是同步点，用它做"握手"、做限流都可以，但要避免互相等待造成死锁。
- `sync.WaitGroup` 的 `Add` 要在 `go` 语句之前调用，否则可能 `Wait` 早于 `Add` 返回。

实际输出（`go run ./base/08_concurrency`）：

```
WaitGroup + Mutex 求和: 15
无缓冲 channel: 来自 goroutine
关闭后 range 读到: 1 2 3
关闭后继续接收 → 零值 0 ok = false
worker pool 平方和: 55
select 超时分支
select 选中已关闭的 channel
context 结束原因: context canceled
sync.Once 只打印一次
```

## 10. 泛型

Go 1.18 引入类型参数，1.21 起标准库有 `slices` / `maps` / `cmp`。

```go
type Number interface {          // 联合约束；~ 表示底层类型匹配即算
    ~int | ~int64 | ~float64
}

func Sum[T Number](nums []T) T { // 类型参数列表
    var total T
    for _, n := range nums { total += n }
    return total
}

type Stack[T any] struct{ items []T }        // 泛型类型
func (s *Stack[T]) Push(v T) { ... }

func Max[T cmp.Ordered](a, b T) T { ... }    // 用标准库约束

fmt.Println(Sum([]int{1,2,3}))               // 类型推断，通常不用写 [int]
```

- 常见约束：`any`、`comparable`、`cmp.Ordered`、自定义联合类型。
- `~T` 表示"底层类型是 T 的自定义类型也算"，写库时很重要。
- 泛型适合容器、算法类代码；只有一两个类型时就别硬套泛型，接口 + 具体实现往往更清楚。

实际输出（`go run ./base/09_generics`）：

```
Sum[int]: 6 | Sum[float64]: 4
Stack 弹出: b a | 空栈 ok = false
Max: 7 z 1.5
Map 求长度: [2 4 1]
标准库 slices: [2 5 9] | Max: 9 | Contains: true
```

## 11. 常用标准库

| 包 | 常用 API |
| --- | --- |
| `fmt` | `Println` `Printf` `Sprintf` `Errorf` |
| `strings` | `ToUpper` `Contains` `Split` `Join` `TrimSpace` `ReplaceAll` `HasPrefix` `Repeat` |
| `strconv` | `Atoi` `Itoa` `ParseFloat` `FormatFloat` `Quote` |
| `sort` / `slices` | `sort.Ints` `sort.Slice` `slices.Sort` `slices.Contains` `slices.Max` |
| `time` | `time.Now` `time.Date` `Add` `Format` `Parse` `time.DateOnly` `time.After` |
| `encoding/json` | `json.Marshal` `MarshalIndent` `Unmarshal` |
| `errors` | `New` `Is` `As` `Unwrap` |
| `context` | `Background` `WithTimeout` `WithCancel` `Done` `Err` |
| `sync` | `WaitGroup` `Mutex` `RWMutex` `Once` |
| `net/http` | `http.Get` `http.HandleFunc` `http.ListenAndServe` |

记忆点：

- `fmt` 常用动词：`%v` 通用、`%+v` 带字段名、`%T` 类型、`%q` 加引号字符串、`%d` `%f` `%s` `%t` `%p`（指针）、`%w`（仅 `Errorf`）。
- `time` 的格式模板是**参考时间** `2006-01-02 15:04:05`，不是 `yyyy-MM-dd`。常量 `time.DateOnly` = `2006-01-02`，`time.RFC3339` 常用于日志/接口。
- JSON 标签控制字段名与省略：`json:"name"`、`json:"email,omitempty"`；未匹配的字段默认忽略，`json.Unmarshal` 需要传指针。

实际输出（`go run ./base/10_stdlib`，节选）：

```
GO true [a b c]
Atoi: 42 <nil>
Atoi 失败时 err != nil: true
Itoa: 7 | ParseFloat: 3.14 | FormatFloat: 0.333
Quote: "a\nb"          （双引号内是反斜杠 + n 两个字符，已转义）
排序: [1 3 5 9]
格式化: 2026-09-30 10:00:00
加 48 小时: 2026-10-02 | 时长字面量: 2h30m0s
Marshal: {"name":"小明","age":18,"created_at":"2026-09-30T10:00:00Z"}
Unmarshal: 小红 20 | 未知字段默认忽略, err = <nil>
包装错误: 读取用户失败: EOF | errors.Is(err, ErrEOF): true
```

## 12. 测试与工具链

表驱动测试 + 子测试是 Go 的标准写法（见 `base/11_testing/`）：

```go
func TestAdd(t *testing.T) {
    cases := []struct {
        name string
        a, b int
        want int
    }{
        {"正数", 1, 2, 3},
        {"负数", -1, -2, -3},
    }
    for _, c := range cases {
        t.Run(c.name, func(t *testing.T) {
            if got := Add(c.a, c.b); got != c.want {
                t.Errorf("Add(%d, %d) = %d, 想要 %d", c.a, c.b, got, c.want)
            }
        })
    }
}
```

- 测试文件与被测代码同包，文件名 `_test.go`；`t.Errorf` 继续执行，`t.Fatalf` 立即结束。
- 断言错误用 `errors.Is`，不要比字符串。
- 基准测试：`func BenchmarkAdd(b *testing.B)` + `b.N` 循环。

常用命令：

| 命令 | 作用 |
| --- | --- |
| `go run ./base/01_vars` | 编译并运行 |
| `go build ./...` | 编译全部包 |
| `go test ./... -v` | 跑测试（`-run` 过滤，`-race` 查数据竞争） |
| `go test -bench . -benchmem` | 基准测试与内存分配 |
| `go test -cover ./...` | 覆盖率 |
| `go vet ./...` | 静态检查，找可疑代码 |
| `gofmt -l .` / `gofmt -w .` | 检查 / 修复格式（提交前必跑） |
| `go mod init` / `go mod tidy` | 初始化模块 / 整理依赖 |

实际运行结果（`go test ./base/11_testing -v`）：

```
=== RUN   TestAdd
=== RUN   TestAdd/正数
=== RUN   TestAdd/负数
=== RUN   TestAdd/零
--- PASS: TestAdd (0.00s)
=== RUN   TestDivide
--- PASS: TestDivide (0.00s)
PASS
ok      gobasics/base/11_testing   0.459s
```

## 13. 易错点速查

| 现象 | 原因 / 正确做法 |
| --- | --- |
| `i` 在闭包里永远是最后一个值 | Go ≤1.21 的循环变量复用；升级到 1.22+ 或写 `i := i` |
| 改切片元素影响"另一个"切片 | 切片表达式共享底层数组；用 `copy` 建独立副本 |
| 函数内 `append` 结果丢失 | `append` 返回新切片；让它返回给调用方 |
| `nil map` 写入 panic | 用 `make(map[K]V)` 或字面量初始化 |
| `s[i]` 取中文乱码 | 按字节索引；用 `[]rune(s)` 或 `range` |
| 接口 `!= nil` 但里面是 nil 指针 | 接口非空判断要看动态类型；别返回具体类型的 nil 指针 |
| 浮点金额算不准 | 用 `decimal` 或整数分单位存储 |
| goroutine 里 panic 会整个进程挂掉 | 顶层加 `recover`，把错误传回主流程 |
| `defer` 里的循环变量值不对 | `defer` 参数注册时求值；必要时用闭包捕获 |
| 忘记 `wg.Add(1)` | `Add` 必须在 `go` 之前 |
| JSON 字段导出为空 | 只有**导出字段**（首字母大写）才会被序列化 |
| 时间格式化输出 `yyyy-MM-dd` | 模板要用参考时间 `2006-01-02` |

## 14. 自测清单

合上文档，能自己写出来就算过关：

**基础**

- [ ] 说清 `var` / `:=` 的区别与使用场景，列出各类型的零值。
- [ ] 用 `iota` 定义一组状态常量。
- [ ] 解释 `len("中文")` 与 `utf8.RuneCountInString` 的差别。
- [ ] 手写 `for` 的三种形态，以及带标签的 `break/continue`。

**数据结构**

- [ ] 说明数组与切片的区别，以及 `len` 与 `cap` 的关系。
- [ ] 写一段代码演示切片共享底层数组，并用 `copy` 修复。
- [ ] 用 comma-ok 判断 map key 是否存在，并按 key 排序输出。

**抽象**

- [ ] 写出一个接口和两个实现，并解释为什么不需要显式声明实现。
- [ ] 用 `errors.Is` / `errors.As` / `%w` 处理一个带包装的错误。
- [ ] 解释值接收者与指针接收者分别什么时候用。

**并发**

- [ ] 用 `WaitGroup` + `Mutex` 安全地对共享计数器自增。
- [ ] 用已关闭的 channel + `range` 实现 worker pool。
- [ ] 用 `select` + `time.After` 实现超时，用 `context` 实现取消。

**工程**

- [ ] 写一个表驱动测试并跑通 `go test -v`。
- [ ] 跑 `gofmt -l .`、`go vet ./...`、`go test -race ./...` 三项检查。

---

目录结构：基础示例都放在 `base/` 下，`base/01_vars` … `base/10_stdlib` 是可运行示例（`go run ./base/目录名`），`base/11_testing` 用 `go test` 运行；模块名 `gobasics`，`go.mod` 声明 `go 1.25`（`12_observability` 依赖的 OpenTelemetry SDK 要求 1.25+，这些基础示例本身在更低的 Go 版本上也能跑）。
