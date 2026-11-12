# 第 4 部分：数组、切片（slice）、map

这一部分讲 Go 里存放"一组数据"的方式。对照 JS：

- **切片（slice）** ≈ JS 的数组 `[]`
- **map** ≈ JS 的对象 `{}` 或 `Map`
- **数组（array）** 长度固定，JS 里没有直接对应的东西

实际开发中**切片和 map 用得最多**，数组很少直接用。

## 1. 数组：长度固定

```go
var nums [3]int              // 长度为 3 的 int 数组，零值是 [0 0 0]
nums[0] = 10
fmt.Println(nums)            // [10 0 0]

names := [3]string{"小明", "小红", "小刚"}
fmt.Println(len(names))      // 3
```

`[3]int` 中的 **3 是类型的一部分**：`[3]int` 和 `[4]int` 是不同的类型，长度不能变。

**数组赋值时会复制一份**：

```go
a := [3]int{1, 2, 3}
b := a          // 复制了一整份
b[0] = 100
fmt.Println(a)  // [1 2 3]    ← a 没变
fmt.Println(b)  // [100 2 3]
```

---

## 2. 切片：可变长度（最常用）

**方括号里不写长度，就是切片：**

```go
nums := []int{1, 2, 3}       // 切片
arr  := [3]int{1, 2, 3}      // 数组
```

### 基本操作

```go
nums := []int{10, 20, 30}

fmt.Println(nums[0])     // 10       读取
nums[1] = 99             //          修改
fmt.Println(len(nums))   // 3        长度
fmt.Println(nums)        // [10 99 30]
```

### 添加元素：append

相当于 JS 的 `push`：

```go
nums := []int{1, 2}
nums = append(nums, 3)          // 加一个
nums = append(nums, 4, 5, 6)    // 一次加多个
fmt.Println(nums)               // [1 2 3 4 5 6]
```

⚠️ **一定要写 `nums = append(...)`**。`append` 会**返回一个新切片**，只写 `append(nums, 3)` 不赋值会编译报错。

### 空切片和 make

```go
var s []int                   // 零值是 nil，长度为 0，可以直接 append
s = append(s, 1)              // ✅

s2 := make([]int, 3)          // 长度为 3：[0 0 0]
s3 := make([]int, 0, 100)     // 长度 0，预留 100 个位置的空间
```

`make` 的第三个参数是**容量（capacity）**，用 `cap(s)` 查看。提前知道大概要放多少个元素时，预留容量可以减少内存的重复分配。

### 越界会直接崩溃

```go
nums := []int{1, 2, 3}
fmt.Println(nums[5])
// panic: runtime error: index out of range [5] with length 3
```

JS 里访问不存在的下标得到 `undefined`，Go 会直接让程序崩溃（**panic**）。

### 截取：`s[开始:结束]`

包含开头，不包含结尾：

```go
nums := []int{0, 1, 2, 3, 4, 5}

fmt.Println(nums[1:4])   // [1 2 3]
fmt.Println(nums[:3])    // [0 1 2]
fmt.Println(nums[3:])    // [3 4 5]
```

### ⚠️ 截取出来的切片和原切片共享数据

```go
a := []int{1, 2, 3, 4, 5}
b := a[1:3]       // [2 3]
b[0] = 99
fmt.Println(a)    // [1 99 3 4 5]   ← a 也被改了！
```

JS 的 `slice()` 会复制一份新数组，但 Go 的截取**不会复制**。

切片之间直接赋值也一样（和数组相反）：

```go
s := []int{1, 2, 3}
t := s            // 不会复制，指向同一份数据
t[0] = 100
fmt.Println(s)    // [100 2 3]
```

需要独立的副本，可以用 `slices.Clone(s)`（需要 `import "slices"`）。

### 遍历切片

```go
fruits := []string{"苹果", "香蕉", "橙子"}

for i, f := range fruits {
    fmt.Println(i, f)
}
```

### 没有 map / filter 方法

JS 里的 `arr.filter(x => x > 2)`，在 Go 里要自己写循环：

```go
nums := []int{1, 2, 3, 4, 5}
var big []int
for _, n := range nums {
    if n > 2 {
        big = append(big, n)
    }
}
fmt.Println(big)   // [3 4 5]
```

标准库 `slices` 包提供了一些常用函数，比如 `slices.Contains(nums, 3)`、`slices.Sort(nums)`。

---

## 3. map：键值对

写法：`map[键的类型]值的类型`

```go
scores := map[string]int{
    "小明": 90,
    "小红": 85,    // ← 多行写法，最后一项后面也要加逗号
}

scores["小刚"] = 78           // 添加或修改
fmt.Println(scores["小明"])   // 90
fmt.Println(len(scores))     // 3
delete(scores, "小红")        // 删除
```

### 读取不存在的键，得到零值

```go
fmt.Println(scores["不存在的人"])   // 0（不是 undefined，也不报错）
```

### "逗号 ok"写法：判断键是否存在

拿到 0，是"考了 0 分"还是"没有这个人"？用第二个返回值判断：

```go
score, ok := scores["小明"]
if ok {
    fmt.Println("找到了，分数：", score)
} else {
    fmt.Println("没有这个人")
}

// 结合 if 初始化语句
if score, ok := scores["小李"]; ok {
    fmt.Println(score)
} else {
    fmt.Println("没有小李")
}
```

### 遍历 map

```go
for name, score := range scores {
    fmt.Println(name, score)
}
```

⚠️ **map 的遍历顺序是随机的**，每次运行都可能不一样。需要固定顺序时，先把键取出来放进切片，用 `slices.Sort` 排好序，再按顺序取值。

### ⚠️ nil map 不能写入

```go
var m map[string]int     // 零值是 nil
m["a"] = 1               // ❌ panic: assignment to entry in nil map

m2 := make(map[string]int)    // ✅ 用 make 创建
m3 := map[string]int{}        // ✅ 或者用空的字面量
```

这一点和切片不同：nil 切片可以直接 `append`，但 nil map 不能写入。

---

## 4. 综合例子：统计单词出现次数

```go
words := []string{"go", "js", "go", "py", "go", "js"}

count := make(map[string]int)
for _, w := range words {
    count[w]++     // 键不存在时读到零值 0，然后加 1
}

fmt.Println(count["go"], count["js"], count["py"])   // 3 2 1
```

`count[w]++` 利用了"不存在的键返回零值"。在 JS 里得先写 `if (!count[w]) count[w] = 0`。

---

## 📝 练习

### 练习 1：预测输出

代码在 [predict/main.go](predict/main.go)：

```go
s := []int{10, 20, 30}
s = append(s, 40)
t := s[1:3]
t[1] = 0
fmt.Println(s, len(t))

m := map[string]int{"a": 1}
m["b"] += 5
v, ok := m["c"]
fmt.Println(m["a"], m["b"], v, ok, len(m))
```

<details>
<summary>答案</summary>

```
[10 20 0 40] 2
1 5 0 false 2
```

**第一行：** `t := s[1:3]` 得到 `[20 30]`，和 `s` 共享数据。`t[1]` 对应 `s[2]`，改成 0 后 `s` 变成 `[10 20 0 40]`。`t` 有 2 个元素。

**第二行：**

- `m["a"]` → 1
- `m["b"]` → 5："b" 不存在，先读到零值 0，0 + 5 = 5，再写进去
- `v` → 0、`ok` → false："c" 不存在
- `len(m)` → 2：`m["c"]` 只是读取，**不会**让 map 多出一个键

</details>

### 练习 2：及格成绩

写一个函数 `passed`，接收一个成绩切片 `[]int`，返回：

1. 所有及格（≥ 60）的成绩组成的新切片
2. 及格的人数

用 `[]int{55, 78, 90, 42, 60, 88}` 测试。

进阶：再写一个函数，接收 `map[string]int`（姓名 → 成绩），打印所有不及格的人。

<details>
<summary>答案</summary>

见 [passed/main.go](passed/main.go)：

```go
func passed(scores []int) ([]int, int) {
    var result []int
    for _, s := range scores {
        if s >= 60 {
            result = append(result, s)
        }
    }
    return result, len(result)
}

func printFailed(scores map[string]int) {
    for name, score := range scores {
        if score < 60 {
            fmt.Println(name, "不及格，分数：", score)
        }
    }
}
```

输出：

```
及格的成绩： [78 90 60 88]
及格人数： 4
小红 不及格，分数： 55
小李 不及格，分数： 42
```

- 人数不用另外计数，直接 `len(result)`
- 60 也算及格，条件是 `>= 60`
- 不及格名单的顺序每次可能不同，因为 map 遍历是随机的

</details>

---

**本部分核心：切片用 `append` 添加，要赋值回去；截取的切片和原切片共享数据；map 读不存在的键得到零值，用 `v, ok :=` 判断；nil map 不能写入；map 遍历顺序随机。**
