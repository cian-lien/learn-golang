// 第 5 部分练习 2：Shape 接口
// 运行：go run ./05-structs-interfaces/shapes
// 输出：
// 15.141592653589793
// 15.14
// 矩形 3x4
package main

import (
	"fmt"
	"math"
)

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

// Rect 有了 Area() 方法，就自动满足 Shape 接口
func (r Rect) Area() float64 {
	return r.Width * r.Height
}

// Circle 也一样
func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

// 进阶：满足 fmt.Stringer 接口，Println 时自动调用
func (r Rect) String() string {
	return fmt.Sprintf("矩形 %gx%g", r.Width, r.Height)
}

func totalArea(shapes []Shape) float64 {
	var total float64
	for _, shape := range shapes {
		total += shape.Area() // 不用关心具体是矩形还是圆
	}
	return total
}

func main() {
	shapes := []Shape{
		Rect{Width: 3, Height: 4},
		Circle{Radius: 1},
	}
	fmt.Println(totalArea(shapes))
	fmt.Printf("%.2f\n", totalArea(shapes))

	fmt.Println(Rect{Width: 3, Height: 4})
}
