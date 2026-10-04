package main

import (
	"errors"
	"fmt"
)

// 多返回值 + 命名返回值（配合裸 return）
func divmod(a, b int) (q, r int) {
	q = a / b
	r = a % b
	return
}

// 变参：函数内拿到的是切片
func sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// 函数是一等公民
func apply(x int, f func(int) int) int { return f(x) }

// 闭包捕获外部局部变量，变量生命周期被延长
func counter() func() int {
	n := 0
	return func() int {
		n++
		return n
	}
}

// defer：LIFO 执行，参数在注册时求值
func deferDemo() {
	for i := 0; i < 3; i++ {
		defer fmt.Println("  defer", i)
	}
	fmt.Println("deferDemo 函数体结束")
}

// defer 可以修改命名返回值
func wrapErr() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("包装后: %w", err)
		}
	}()
	return errors.New("原始错误")
}

func main() {
	q, r := divmod(7, 3)
	fmt.Println("divmod:", q, r)

	fmt.Println("变参:", sum(1, 2, 3), sum(), sum([]int{4, 5}...))
	fmt.Println("函数作为参数:", apply(5, func(x int) int { return x * x }))

	next := counter()
	fmt.Println("闭包计数:", next(), next(), next())

	deferDemo()
	fmt.Println("命名返回值 + defer:", wrapErr())

	func(msg string) { fmt.Println("立即执行函数:", msg) }("来自匿名函数")
}
