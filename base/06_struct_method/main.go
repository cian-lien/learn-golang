package main

import "fmt"

type Point struct {
	X, Y int
}

// 值接收者：操作的是副本
func (p Point) MoveByValue(dx int) Point {
	p.X += dx
	return p
}

// 指针接收者：能修改原对象，也避免复制大结构体
func (p *Point) MoveByPointer(dx int) {
	p.X += dx
}

type Animal struct{ Name string }

func (a Animal) Speak() string { return a.Name + " 发出声音" }

// 嵌入（匿名字段）≈ 组合 + 字段/方法提升
type Dog struct {
	Animal
	Breed string
}

func (d Dog) Speak() string { return d.Name + " 汪汪" }

func (d Dog) Info() string { return fmt.Sprintf("%s(%s)", d.Name, d.Breed) }

func main() {
	p := Point{1, 2}
	p2 := p.MoveByValue(10)
	fmt.Println("值接收者:", p, p2)

	p.MoveByPointer(10)
	fmt.Println("指针接收者:", p)

	// 字段都可比较时，结构体可以 ==
	fmt.Println("结构体可比较:", Point{1, 2} == Point{1, 2}, Point{1, 2} == Point{2, 2})

	d := Dog{Animal: Animal{Name: "旺财"}, Breed: "柴犬"}
	fmt.Println("嵌入字段提升:", d.Name)
	fmt.Println("方法提升:", d.Animal.Speak(), "| 外层同名方法覆盖:", d.Speak())
	fmt.Println("组合:", d.Info())

	// 指针对结构体自动解引用
	pp := &p
	pp.Y = 100
	(*pp).X = 50
	fmt.Println("指针访问:", p, "| pp.X =", pp.X)

	// 复合字面量
	fmt.Println("字面量:", Point{X: 7}, Point{})
}
