# 第 2 部分：流程控制（if / for / switch）

这部分和 JS 很像，但 Go 做了一些简化，写法上也更严格。

## 1. if

```go
age := 20

if age >= 18 {
    fmt.Println("成年")
} else if age >= 12 {
    fmt.Println("青少年")
} else {
    fmt.Println("儿童")
}
```

和 JS 的区别：

- 条件**不用写括号** `()`，写了也能运行，但格式化工具会自动删掉
- 大括号 `{}` **必须写**，就算只有一行也不能省略
- `else` 必须和前面的 `}` 写在**同一行**

```go
// ❌ JS 风格的写法在 Go 里会报错
if (age >= 18) fmt.Println("成年")
```

### 条件必须是 bool 类型

不能像 JS 那样用"真值 / 假值"：

```go
count := 0
// if count { }        // ❌ 报错：int 不能当 bool 用
if count != 0 { }       // ✅ 必须写清楚
```

### 逻辑运算符

和 JS 完全一样：`&&`（并且）、`||`（或者）、`!`（取反）。

### if 的特殊写法：先执行一句，再判断

```go
if score := 85; score >= 60 {
    fmt.Println("及格，分数：", score)
}
// 这里不能再用 score，它只在 if/else 块里有效
```

分号前面是初始化语句，后面是条件。这种写法在错误处理里会**大量用到**。

**Go 没有三元运算符** `a ? b : c`，只能写 if-else。

---

## 2. for：Go 唯一的循环

Go 没有 `while`，也没有 `do...while`，**所有循环都用 `for`**。

### 写法一：经典三段式

和 JS 一样，只是不加括号。三段之间用**分号**分隔：

```go
for i := 0; i < 5; i++ {
    fmt.Println(i)   // 0 1 2 3 4
}
```

### 写法二：只写条件，相当于 while

```go
n := 1
for n < 100 {
    n *= 2
}
fmt.Println(n)   // 128
```

### 写法三：什么都不写，就是死循环

要靠 `break` 跳出：

```go
count := 0
for {
    count++
    if count == 3 {
        break
    }
}
```

### break 和 continue

```go
for i := 1; i <= 6; i++ {
    if i%2 == 0 {
        continue   // 跳过偶数，直接进入下一轮
    }
    fmt.Println(i)   // 1 3 5
}
```

### 写法四：for range 遍历

`range` 用来遍历字符串、数组、切片、map 等，类似 JS 的 `for...of`，但会**同时给出索引和值**。

```go
for i, ch := range "Go语言" {
    fmt.Printf("索引 %d：%c\n", i, ch)
}
```

输出：

```
索引 0：G
索引 1：o
索引 2：语
索引 5：言     ← 注意是 5 不是 3！
```

`range` 会**按字符遍历**（中文不会被拆坏），但索引 `i` 是**字节位置**。"语"占了 2、3、4 三个字节，所以"言"从 5 开始。`%c` 用来按字符打印 rune。

### 用 `_` 丢掉不需要的值

```go
for _, ch := range "abc" {   // 不要索引
    fmt.Printf("%c ", ch)
}
```

因为 Go 规定**声明了不用的变量会报错**，`_` 就是专门用来占位、表示"我不要这个值"的。

### 遍历整数（Go 1.22+）

```go
for i := range 3 {
    fmt.Println(i)   // 0 1 2
}
```

---

## 3. switch

Go 的 switch 最大的区别是：**每个 case 执行完会自动结束，不需要写 `break`**。

```go
day := 3

switch day {
case 1:
    fmt.Println("周一")
case 2:
    fmt.Println("周二")
case 3:
    fmt.Println("周三")    // 只打印这一行，不会继续往下执行
default:
    fmt.Println("其他")
}
```

JS 里忘写 `break` 会"穿透"到下一个 case，这是个经典 bug。Go 把默认行为改成了不穿透。

### 一个 case 匹配多个值

```go
switch day {
case 1, 2, 3, 4, 5:
    fmt.Println("工作日")
case 6, 7:
    fmt.Println("周末")
}
```

### 不带表达式的 switch：代替长串 if-else

```go
score := 78

switch {
case score >= 90:
    fmt.Println("优秀")
case score >= 75:
    fmt.Println("良好")    // 输出这个
case score >= 60:
    fmt.Println("及格")
default:
    fmt.Println("不及格")
}
```

`switch` 后面什么都不写，等于 `switch true`，会从上往下找**第一个为 true 的 case**。

### case 的格式

整个 switch 只有**一对**大括号，每个 case 用**冒号**结尾，不需要自己加 `{}`：

```go
} case i%3 == 0 {     // ❌ case 后面不用大括号
case i%3 == 0:        // ✅
    fmt.Println("Fizz")
```

### fallthrough：强制穿透

很少用，知道有就行：

```go
switch 1 {
case 1:
    fmt.Println("一")
    fallthrough        // 强制继续执行下一个 case
case 2:
    fmt.Println("二")
}
// 输出：一 二
```

---

## 📝 练习

### 练习 1：预测输出

代码在 [range-switch/main.go](range-switch/main.go)：

```go
package main

import "fmt"

func main() {
    for i, ch := range "Hi你好" {
        if ch == 'i' {
            continue
        }
        fmt.Print(i, " ")
    }
    fmt.Println()

    x := 15
    switch {
    case x > 10:
        fmt.Println("A")
    case x > 5:
        fmt.Println("B")
    }
}
```

<details>
<summary>答案</summary>

```
0 2 5
A
```

- `H` 索引 0，打印；`i` 索引 1，被 `continue` 跳过；`你` 从字节位置 2 开始；`好` 从 5 开始（`你` 占了 2、3、4）
- x = 15 同时满足两个 case，但 switch 只执行**第一个**为 true 的 case，且不会穿透

</details>

### 练习 2：FizzBuzz

打印 1 到 15：

- 能被 3 和 5 同时整除，打印 `FizzBuzz`
- 只能被 3 整除，打印 `Fizz`
- 只能被 5 整除，打印 `Buzz`
- 其他情况打印数字本身

提示：用 `for` 加上不带表达式的 `switch`，判断整除用 `%`。注意 case 的顺序会影响结果。

<details>
<summary>答案</summary>

见 [fizzbuzz/main.go](fizzbuzz/main.go)：

```go
for i := 1; i <= 15; i++ {
    switch {
    case i%3 == 0 && i%5 == 0:
        fmt.Println("FizzBuzz")
    case i%3 == 0:
        fmt.Println("Fizz")
    case i%5 == 0:
        fmt.Println("Buzz")
    default:
        fmt.Println(i)
    }
}
```

**常见错误：**

| 错误写法 | 正确写法 | 原因 |
|---|---|---|
| `for i=0,i<=15,i++` | `for i := 1; i <= 15; i++` | 用分号分隔；新变量用 `:=`；从 1 开始（`0` 会多打印一行 FizzBuzz） |
| `switch i%3==0, i%5==0 {` | `switch {` + `case i%3 == 0 && i%5 == 0:` | "同时满足"要用 `&&`，写在 case 里 |
| `} case i%3 == 0 {` | `case i%3 == 0:` | case 用冒号，不加大括号 |
| 没写 `default` | `default: fmt.Println(i)` | 其他情况要打印数字本身 |

FizzBuzz 的判断必须放在第一个，否则 15 会先匹配到 `i%3 == 0`，打印成 Fizz。

</details>

---

**本部分核心：if 和 for 都不加括号；Go 只有 for 一种循环；switch 不需要 break；range 遍历字符串时索引是字节位置。**
