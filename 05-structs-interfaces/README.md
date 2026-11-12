# 第 5 部分：结构体、方法、接口

**Go 没有 class，也没有继承。** 它用三样东西代替：

- **结构体（struct）**：存放数据，相当于对象的"形状"
- **方法（method）**：给结构体绑定函数，相当于类的方法
- **接口（interface）**：规定"能做什么"，用来实现多态

Go 的接口和 TypeScript 一样是按"结构"匹配的，学过 TS 会很好理解。

## 1. 结构体：定义数据的形状

```go
type User struct {
    Name string
    Age  int
}
```

很像 TS 的 `type User = { name: string; age: number }`，区别是 Go 的结构体是一个**实实在在的类型**，可以直接创建值来用。

### 创建和使用

```go
u := User{Name: "小明", Age: 20}
fmt.Println(u.Name)      // 小明
u.Age = 21               // 修改字段

var empty User           // 没赋值的字段是零值
fmt.Printf("%+v\n", empty)   // {Name: Age:0}
```

`%+v` 打印结构体时会带上字段名。也可以按顺序赋值 `User{"小明", 20}`，但**推荐总是写字段名**。

### ⚠️ 结构体赋值会复制一份

```go
a := User{Name: "小明", Age: 20}
b := a          // 复制了一整份
b.Name = "小红"
fmt.Println(a.Name)   // 小明  ← a 没变
```

JS 里 `const b = a` 之后改 `b.name`，`a.name` 也会变（JS 对象是引用）。**Go 的结构体是"值类型"**，赋值、传参时都会复制。

### 字段名大小写决定是否公开

```go
type User struct {
    Name     string   // 大写：其他包可以访问
    password string   // 小写：只有本包能访问
}
```

### 结构体标签（后端经常用）

把结构体转成 JSON 时，用**标签**指定 JSON 里的字段名：

```go
import "encoding/json"

type User struct {
    Name     string `json:"name"`
    Age      int    `json:"age"`
    password string
}

u := User{Name: "小明", Age: 20, password: "123"}
data, _ := json.Marshal(u)
fmt.Println(string(data))       // {"name":"小明","age":20}
```

`password` 是小写的，`json` 包访问不到，所以不会出现在 JSON 里。

---

## 2. 方法：给结构体绑定函数

```go
func (u User) Greet() string {
    return "你好，我是" + u.Name
}

u := User{Name: "小明", Age: 20}
fmt.Println(u.Greet())   // 你好，我是小明
```

`func` 后面的 **`(u User)`** 叫**接收者（receiver）**，表示"这个方法属于 User"。`u` 相当于 JS class 里的 `this`，只是要自己起名字（习惯用类型名首字母小写）。

### ⚠️ 想修改结构体，要用指针接收者

```go
// 值接收者：u 是一份副本
func (u User) BirthdayWrong() {
    u.Age++          // 改的是副本
}

// 指针接收者：u 指向原来的那个结构体
func (u *User) Birthday() {
    u.Age++          // 改的是原来的
}

u := User{Name: "小明", Age: 20}
u.BirthdayWrong()
fmt.Println(u.Age)   // 20  ← 没变
u.Birthday()
fmt.Println(u.Age)   // 21  ← 变了
```

结构体传递时会复制，值接收者拿到的是副本。`*User` 读作"指向 User 的指针"：不复制，直接拿原来那个值的地址（第 6 部分详细讲）。

> **方法需要修改字段 → 用 `*User`（指针接收者）**
> **只读取、不修改 → 用 `User`（值接收者）也可以**

实际项目里**大多数方法都用指针接收者**。调用时不用操心，`u.Birthday()` 会被 Go 自动处理。

### "构造函数"的写法

Go 没有 `constructor`，约定写一个 `New` 开头的普通函数：

```go
func NewUser(name string, age int) *User {
    return &User{Name: name, Age: age}   // & 表示取地址，返回指针
}

u := NewUser("小明", 20)
```

---

## 3. 嵌入：用组合代替继承

Go 没有 `extends`，但可以把一个结构体**嵌入**到另一个里面：

```go
type Admin struct {
    User          // 只写类型，不写字段名，就是"嵌入"
    Level int
}

a := Admin{
    User:  User{Name: "老王", Age: 40},
    Level: 9,
}

fmt.Println(a.Name)      // 老王          ← 直接访问 User 的字段
fmt.Println(a.Greet())   // 你好，我是老王  ← 直接调用 User 的方法
```

看起来像继承，本质上是 **Admin 里面"装着"一个 User**（完整写法是 `a.User.Name`）。Go 的设计思路是"组合优于继承"。

---

## 4. 接口：规定"能做什么"

接口只描述**有哪些方法**，不关心具体是什么类型：

```go
type Speaker interface {
    Speak() string
}
```

**任何拥有 `Speak() string` 方法的类型，都算是 Speaker：**

```go
type Dog struct{}
func (d Dog) Speak() string { return "汪汪" }

type Cat struct{}
func (c Cat) Speak() string { return "喵喵" }

func makeSound(s Speaker) {
    fmt.Println(s.Speak())
}

func main() {
    makeSound(Dog{})   // 汪汪
    makeSound(Cat{})   // 喵喵

    animals := []Speaker{Dog{}, Cat{}, Dog{}}
    for _, a := range animals {
        fmt.Println(a.Speak())
    }
}
```

**Go 不需要写 `implements`。** 只要有对应的方法，就自动满足接口。和 TypeScript 的思路一样，常被叫做"鸭子类型"。

反过来，**没有对应方法就不满足接口**，放进接口类型的变量会报错：

```
Rect does not implement Shape (missing method Area)
```

### 你已经用过接口了：`error`

```go
type error interface {
    Error() string
}
```

任何有 `Error() string` 方法的类型都可以当作 error 返回。`errors.New` 只是帮你创建了一个满足这个接口的值。

### 常用接口：`fmt.Stringer`

类型有 `String() string` 方法时，`fmt.Println` 会自动调用它：

```go
type Point struct{ X, Y int }

func (p Point) String() string {
    return fmt.Sprintf("(%d, %d)", p.X, p.Y)
}

fmt.Println(Point{1, 2})   // (1, 2)
```

类似 JS 里重写 `toString()`。`fmt.Sprintf` 和 `Printf` 用法一样，只是返回字符串而不打印。

### ⚠️ 指针接收者和接口

```go
type Robot struct{}
func (r *Robot) Speak() string { return "哔哔" }   // 注意是 *Robot

var s Speaker = Robot{}    // ❌ 编译报错
var s Speaker = &Robot{}   // ✅
```

报错：`Robot does not implement Speaker (method Speak has pointer receiver)`。方法定义在 `*Robot` 上，只有指针满足接口。遇到这个报错，在值前面加 `&`。

### 空接口 `any`：可以装任何类型

```go
var x any
x = 42
x = "hello"
```

`any` 相当于 TS 的 `unknown`，用之前要判断类型，方法是**类型 switch**：

```go
func describe(x any) {
    switch v := x.(type) {
    case int:
        fmt.Println("整数", v)
    case string:
        fmt.Println("字符串", v)
    default:
        fmt.Println("其他类型", v)
    }
}

describe(1)      // 整数 1
describe("hi")   // 字符串 hi
describe(3.5)    // 其他类型 3.5
```

`any` 要少用，用多了就失去了静态类型的好处。

---

## 📝 练习

### 练习 1：预测输出

代码在 [counter/main.go](counter/main.go)：

```go
type Counter struct{ n int }

func (c Counter) AddV()  { c.n++ }
func (c *Counter) AddP() { c.n++ }

func main() {
    a := Counter{}
    b := a
    a.AddV()
    a.AddP()
    a.AddP()
    b.AddP()
    fmt.Println(a.n, b.n)
}
```

<details>
<summary>答案</summary>

```
2 1
```

- `b := a` 复制了一份，之后 a 和 b 互不影响
- `a.AddV()` 是值接收者，改的是副本，a.n 还是 0
- `a.AddP()` 两次，a.n 变成 2
- `b.AddP()` 一次，b.n 变成 1

</details>

### 练习 2：Shape 接口

1. 定义接口 `Shape`，只有一个方法 `Area() float64`
2. 定义 `Rect`（宽、高，面积 = 宽 × 高）和 `Circle`（半径，面积 = π × r²，π 用 `math.Pi`）
3. 写函数 `totalArea(shapes []Shape) float64`，返回面积之和
4. 用一个 3×4 的矩形和半径为 1 的圆测试

进阶：给 `Rect` 加一个 `String()` 方法，让 `fmt.Println` 输出 `矩形 3x4`。

<details>
<summary>答案</summary>

见 [shapes/main.go](shapes/main.go)：

```go
type Shape interface {
    Area() float64
}

type Rect struct {
    Width  float64
    Height float64
}

type Circle struct {
    Radius float64
}

func (r Rect) Area() float64 {
    return r.Width * r.Height
}

func (c Circle) Area() float64 {
    return math.Pi * c.Radius * c.Radius   // Go 没有 ** 运算符
}

func (r Rect) String() string {
    return fmt.Sprintf("矩形 %gx%g", r.Width, r.Height)
}

func totalArea(shapes []Shape) float64 {
    var total float64
    for _, shape := range shapes {
        total += shape.Area()
    }
    return total
}
```

输出：

```
15.141592653589793
15.14
矩形 3x4
```

关键是两个 `Area()` 方法：加上之后，`Rect` 和 `Circle` 才满足 `Shape` 接口，才能放进 `[]Shape`。计算面积只读取字段，所以用值接收者，直接写 `Rect{...}` 就行，不需要 `&`。

`%g` 会去掉多余的小数位，所以 `3.0` 打印成 `3`。

</details>

---

**本部分核心：结构体赋值会复制；要修改字段就用指针接收者 `*T`；用嵌入代替继承；接口不需要写 implements，有对应的方法就自动满足。**
