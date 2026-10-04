package main

import "fmt"

// 包级常量：iota 从 0 开始，每行 +1
const (
	StatusOK       = iota // 0
	StatusNotFound        // 1
	StatusError           // 2
)

const Pi = 3.14159

func main() {
	// 1) var：可显式写类型，也可让编译器推断
	var a int = 10
	var b = 20 // 推断为 int
	var c, d = 1, "x"

	// 2) 短变量声明 := 只能用在函数内部
	e := 30
	f, g := "hello", true

	// 3) 零值：Go 里没有"未初始化"的变量
	var (
		i  int
		s  string
		p  *int
		sl []int
		m  map[string]int
		fn func()
	)
	fmt.Println(a, b, c, d, e, f, g)
	fmt.Printf("零值: int=%d string=%q pointer=%v slice=%v map=%v\n", i, s, p, sl, m)
	fmt.Println("func 零值是不是 nil:", fn == nil)
	fmt.Printf("空 slice 是否 nil: %v, len=%d\n", sl == nil, len(sl))

	// 4) 常量与 iota
	fmt.Println("iota:", StatusOK, StatusNotFound, StatusError, "Pi =", Pi)

	// 5) 无类型常量：精度高，可按上下文隐式转换
	const big = 1 << 40
	var x float64 = big
	fmt.Println("无类型常量赋给 float64:", x)

	// 6) 交换与丢弃
	a, b = b, a
	_, _ = a, b
	fmt.Println("交换后:", a, b)
}
