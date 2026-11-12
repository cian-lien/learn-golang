# 第 1 部分：程序结构、变量、常量、基本类型

## 1. 第一个 Go 程序

```go
package main  // 每个文件都必须属于一个包，main 包表示"可执行程序"

import "fmt"  // 导入标准库的 fmt 包（format，负责输入输出）

func main() {  // 程序入口，程序从这里开始执行
    fmt.Println("Hello, Go!")
}
```

运行方式：保存为 `main.go`，在终端执行 `go run main.go`。

几个要点：

- `package main` 和 `func main()` 两个都要有，程序才能直接运行。
- 语句末尾**不用写分号**，编译器会自动加。
- `{` **必须和 `func` 写在同一行**，换行写会报错。Go 强制统一代码风格。
- 首字母**大写**的名字（比如 `Println`）是对外公开的，相当于 JS 的 `export`；小写的只能在包内使用。

---

## 2. 变量

Go 是**静态类型**语言，每个变量的类型在编译时就确定了，之后不能改。这一点和 JS 很不一样。

```go
package main

import "fmt"

func main() {
    // 写法一：var 变量名 类型 = 值
    var name string = "小明"

    // 写法二：省略类型，编译器根据值自动推断
    var age = 25          // 推断为 int

    // 写法三：短变量声明 :=（最常用，只能在函数内部用）
    city := "广州"         // 等价于 var city = "广州"

    // 一次声明多个
    a, b := 1, 2

    fmt.Println(name, age, city, a, b)
    // 输出：小明 25 广州 1 2
}
```

### 零值（Zero Value）

只声明不赋值的变量会自动得到"零值"，不会像 JS 那样变成 `undefined`。

```go
var i int      // 0
var f float64  // 0
var s string   // ""（空字符串）
var ok bool    // false
```

### ⚠️ 声明了不用会报错

```go
func main() {
    x := 10
    // 下面什么都不写
}
// 编译报错：declared and not used: x
```

Go 规定**声明了但没用到的变量会编译失败**，没用到的 import 也一样。

### `=` 和 `:=` 的区别

```go
x := 1   // 声明并赋值（新变量）
x = 2    // 给已存在的变量重新赋值
x := 3   // ❌ 报错：x 已经声明过了
```

---

## 3. 常量

```go
const Pi = 3.14159
const AppName = "MyApp"

// 一组常量
const (
    StatusOK    = 200
    StatusError = 500
)
```

常量的值在编译时确定，之后不能修改，相当于 JS 的 `const`。区别是 Go 的 const 只能用于数字、字符串、布尔这类基本值，不能用于对象。

### `iota`：自动递增计数器

常用来做枚举：

```go
const (
    Monday = iota  // 0
    Tuesday        // 1（自动沿用 iota，每行 +1）
    Wednesday      // 2
)
```

---

## 4. 基本类型

| 类型 | 说明 | 例子 |
|---|---|---|
| `int` | 整数（最常用） | `42` |
| `int64`、`int32` 等 | 指定位数的整数 | |
| `float64` | 小数（最常用） | `3.14` |
| `string` | 字符串 | `"hello"` |
| `bool` | 布尔 | `true` / `false` |
| `byte` | 一个字节（等同 `uint8`） | |
| `rune` | 一个 Unicode 字符（等同 `int32`） | `'中'` |

### 重点：Go 没有隐式类型转换

```go
a := 10        // int
b := 3.5       // float64

// c := a + b          // ❌ 编译报错：int 和 float64 不能直接相加
c := float64(a) + b    // ✅ 必须手动转换
fmt.Println(c)         // 13.5
```

在 JS 里 `1 + "2"` 会得到 `"12"`，这种事在 Go 里不会发生，类型不一致直接报错。

### 整数除法会丢掉小数

**两个整数相除，结果还是整数**，小数部分直接截断（不是四舍五入）：

```go
fmt.Println(7 / 2)    // 3
fmt.Println(9 / 10)   // 0
fmt.Println(-7 / 2)   // -3  ← 向 0 截断，不是 -4
```

JS 里 `7 / 2` 是 `3.5`，因为 JS 的数字都是浮点数。想在 Go 里得到 3.5，至少要让一边是浮点数：

```go
fmt.Println(7.0 / 2)                  // 3.5
a, b := 7, 2
fmt.Println(float64(a) / float64(b))  // 3.5  变量要手动转换
```

取余数用 `%`：

```go
total := 125
fmt.Println(total/60, "分", total%60, "秒")  // 2 分 5 秒
```

### 字符串长度：len 返回的是字节数

Go 的字符串用 **UTF-8** 编码，不同字符占的字节数不一样：

| 字符类型 | 例子 | 占用字节 |
|---|---|---|
| 英文字母、数字、英文标点、空格 | `a` `Z` `7` `,` `!` | 1 |
| 中文汉字 | `你` `语` | 3 |
| 中文标点 | `，` `。` `！` | 3 |
| Emoji | `😀` | 4 |

```go
fmt.Println(len("Go"))       // 2   → 2 个英文 × 1
fmt.Println(len("语言"))      // 6   → 2 个汉字 × 3
fmt.Println(len("Go语言"))    // 8   → 2 + 6
fmt.Println(len("Hi，你好"))  // 11  → 2 + 3（中文逗号）+ 6

// 要数"看得见的字符个数"
fmt.Println(len([]rune("Go语言")))  // 4
```

中文逗号 `，` 和英文逗号 `,` 看起来很像，但前者占 3 字节，后者只占 1 字节。

另外，双引号 `"a"` 是字符串，单引号 `'a'` 是 rune（单个字符），两者不能混用。

---

## 5. 格式化输出

```go
name := "小明"
age := 25
score := 95.5

fmt.Println(name, age)                       // 小明 25（自动加空格和换行）
fmt.Printf("姓名：%s，年龄：%d\n", name, age) // 按格式输出
fmt.Printf("分数：%.1f\n", score)             // 保留 1 位小数：95.5
fmt.Printf("%v | %T\n", age, age)            // 25 | int
```

常用占位符：

- `%v`：任意值（万能，不确定就用它）
- `%T`：打印**类型**（调试时很好用）
- `%d`：整数
- `%s`：字符串
- `%f`：小数，`%.2f` 保留 2 位小数

---

## 📝 练习：预测输出

先在心里想一下下面代码的输出，再运行验证（代码在 [exercise/main.go](exercise/main.go)）：

```go
package main

import "fmt"

const (
    A = iota
    B
    C
)

func main() {
    var x int
    y := "Go语言"
    z := 7 / 2
    fmt.Println(x, C, len(y), z)
}
```

提示：`7 / 2` 两边都是整数，结果会怎样？

<details>
<summary>答案</summary>

```
0 2 8 3
```

- `x` = **0**：声明了 int 但没赋值，是零值
- `C` = **2**：iota 从 0 开始，A=0、B=1、C=2
- `len(y)` = **8**："Go" 占 2 字节，"语言" 占 6 字节
- `z` = **3**（不是 3.5）：两个整数相除，小数部分被截断

</details>

---

**本部分核心：`:=` 最常用、变量有零值、类型不会自动转换。**
