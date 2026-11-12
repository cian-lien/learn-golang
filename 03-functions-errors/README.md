# 第 3 部分：函数、多返回值、错误处理

这部分是 Go 最有特色的地方。特别是**错误处理**，Go 的做法和 JS 完全不同，写后端代码时几乎每个函数都会用到。

## 1. 函数的基本写法

```go
func add(a int, b int) int {
    return a + b
}

func main() {
    result := add(3, 5)
    fmt.Println(result)   // 8
}
```

结构：`func 函数名(参数 类型) 返回类型 { ... }`

和 TypeScript 对比：

```ts
// TypeScript
function add(a: number, b: number): number { return a + b }
```

```go
// Go
func add(a int, b int) int { return a + b }
```

Go 的类型**直接写在名字后面**，不用冒号。

相同类型的参数可以合并写：

```go
func add(a, b int) int {
    return a + b
}
```

没有返回值的函数，直接不写返回类型：

```go
func sayHi(name string) {
    fmt.Println("你好，" + name)
}
```

---

## 2. 多返回值

**一个函数可以返回多个值**，多个返回类型要用括号包起来：

```go
// 同时返回商和余数
func divide(a, b int) (int, int) {
    return a / b, a % b
}

func main() {
    q, r := divide(17, 5)
    fmt.Println(q, r)   // 3 2
}
```

JS 里想返回多个值，得先包成数组再解构：`const [q, r] = divide(17, 5)`。Go 是语言本身直接支持。

只要其中一个值，用 `_` 丢掉另一个：

```go
q, _ := divide(17, 5)    // 只要商
_, r := divide(17, 5)    // 只要余数
```

---

## 3. 错误处理（重点）

**Go 没有 try/catch。** 可能出错的函数，**把错误当作最后一个返回值返回**，调用的人自己检查。

标准库的例子：`strconv.Atoi` 把字符串转成整数，失败时返回错误：

```go
package main

import (
    "fmt"
    "strconv"
)

func main() {
    n, err := strconv.Atoi("123")
    fmt.Println(n, err)   // 123 <nil>

    n, err = strconv.Atoi("abc")
    fmt.Println(n, err)   // 0 strconv.Atoi: parsing "abc": invalid syntax
}
```

导入多个包时，用括号把它们写在一起。

两个新概念：

- **`error`**：Go 内置的错误类型
- **`nil`**：表示"空、没有"，类似 JS 的 `null`。**`err` 是 `nil` 就说明没出错**

### 标准写法：`if err != nil`

```go
n, err := strconv.Atoi("abc")
if err != nil {
    fmt.Println("转换失败：", err)
    return    // 出错就提前结束
}
fmt.Println("转换成功：", n)
```

这种写法在 Go 代码里**到处**都是。好处是：**哪里可能出错，代码里一眼就能看出来**，不会像 try/catch 那样把错误抛到不知道哪里去。

### 自己写一个会返回错误的函数

```go
import (
    "errors"
    "fmt"
)

func safeDivide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("除数不能为 0")   // 出错：返回零值 + 错误
    }
    return a / b, nil                          // 成功：返回结果 + nil
}

func main() {
    result, err := safeDivide(10, 0)
    if err != nil {
        fmt.Println("出错了：", err)   // 出错了： 除数不能为 0
        return
    }
    fmt.Println(result)
}
```

约定俗成的规则：

- `error` 放在**最后一个**返回值
- 出错时，其他返回值一般给**零值**
- 成功时，error 返回 **`nil`**

需要在错误信息里带上变量时，用 `fmt.Errorf`（用法和 `Printf` 一样）。不需要带变量时，用 `errors.New` 就够了：

```go
return 0, fmt.Errorf("无法计算 %d 除以 %d", a, b)
```

### 结合 if 初始化语句

```go
if n, err := strconv.Atoi("42"); err != nil {
    fmt.Println("出错：", err)
} else {
    fmt.Println("结果：", n)
}
```

### `:=` 左边只要有一个新变量就行

```go
a, err := strconv.Atoi("1")
b, err := strconv.Atoi("2")    // ✅ 可以！
```

`:=` 的规则是：**左边至少有一个新变量**，已经存在的变量（这里是 `err`）会被直接赋值。所以连续写多个 `xxx, err :=` 不会报错。

---

## 4. defer：延迟执行

`defer` 后面的语句等到**当前函数结束时**才执行：

```go
func main() {
    defer fmt.Println("最后执行")
    fmt.Println("先执行")
}
// 先执行
// 最后执行
```

多个 defer 按**后进先出**的顺序执行（像叠盘子）：

```go
func main() {
    defer fmt.Println("1")
    defer fmt.Println("2")
    defer fmt.Println("3")
}
// 3 2 1
```

主要用来做"收尾工作"，比如打开文件之后马上写 `defer file.Close()`，不管函数在哪里 return，文件最后都一定会被关闭。作用类似 JS 里 `try...finally` 的 `finally`。

---

## 5. 匿名函数和闭包

Go 的函数也可以赋值给变量、当作参数传递：

```go
// 类似 JS 的 const square = (x) => x * x
square := func(x int) int {
    return x * x
}
fmt.Println(square(4))   // 16
```

闭包：函数可以"记住"外部的变量：

```go
func makeCounter() func() int {   // 返回值类型是"一个返回 int 的函数"
    count := 0
    return func() int {
        count++
        return count
    }
}

func main() {
    counter := makeCounter()
    fmt.Println(counter())   // 1
    fmt.Println(counter())   // 2
    fmt.Println(counter())   // 3
}
```

原理和 JS 的闭包一样，只是要把类型写出来。

---

## 📝 练习

### 练习 1：预测输出

代码在 [defer-calc/main.go](defer-calc/main.go)：

```go
package main

import (
    "errors"
    "fmt"
)

func calc(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("除数不能为0")
    }
    return a / b, nil
}

func main() {
    defer fmt.Println("结束")

    r, err := calc(9, 2)
    fmt.Println(r, err)

    r, err = calc(1, 0)
    if err != nil {
        fmt.Println("错误：", err)
        return
    }
    fmt.Println("这行会打印吗？")
}
```

<details>
<summary>答案</summary>

```
4 <nil>
错误： 除数不能为0
结束
```

- `9 / 2` 是整数除法，得到 4；没出错时 err 是 nil，打印出来是 `<nil>`
- 第二次除数为 0，打印错误后 `return`，所以"这行会打印吗？"**不会**打印
- `return` 时函数结束，触发 defer，最后打印"结束"

</details>

### 练习 2：parseAge

写一个函数 `parseAge`，接收一个字符串，返回 `(int, error)`：

- 字符串不是数字（比如 `"abc"`），返回错误
- 数字小于 0 或大于 150，返回错误"年龄不合理"
- 其他情况，返回这个年龄

然后在 `main` 里分别用 `"25"`、`"abc"`、`"200"` 调用，出错就打印错误，成功就打印年龄。

<details>
<summary>答案</summary>

见 [parse-age/main.go](parse-age/main.go)：

```go
func parseAge(ageStr string) (int, error) {
    age, err := strconv.Atoi(ageStr)
    if err != nil {
        return 0, err   // 直接把错误往上传
    }
    if age < 0 || age > 150 {
        return 0, fmt.Errorf("年龄不合理：%d", age)
    }
    return age, nil
}

func check(s string) {
    age, err := parseAge(s)
    if err != nil {
        fmt.Println("出错：", err)
        return
    }
    fmt.Println("年龄：", age)
}

func main() {
    check("25")
    check("abc")
    check("200")
}
```

输出：

```
年龄： 25
出错： strconv.Atoi: parsing "abc": invalid syntax
出错： 年龄不合理：200
```

**常见错误：**

- `parseAge(25)`：25 是 int，函数要的是 string，Go 不会自动转换，要写 `"25"`
- `parseAge(abc)`：没有引号，Go 会把 abc 当成变量名，报 `undefined: abc`
- 只写 `parseAge("25")` 不接收返回值：能编译，但结果被丢掉，什么都不会打印

`check` 函数用了"出错就提前 return"的写法，不需要 `else`。**Go 代码通常这样写：先处理错误并返回，正常逻辑放在最后**，代码不会一层层往里缩进。

</details>

---

**本部分核心：类型写在名字后面；函数可以返回多个值；错误作为最后一个返回值，用 `if err != nil` 检查；defer 在函数结束时执行。**
