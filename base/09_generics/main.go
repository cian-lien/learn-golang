package main

import (
	"cmp"
	"fmt"
	"slices"
)

// 类型约束：~ 表示底层类型是其中之一（自定义类型也算）
type Number interface {
	~int | ~int64 | ~float64
}

func Sum[T Number](nums []T) T {
	var total T
	for _, n := range nums {
		total += n
	}
	return total
}

// 泛型类型
type Stack[T any] struct{ items []T }

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }

func (s *Stack[T]) Pop() (T, bool) {
	var zero T // T 的零值
	if len(s.items) == 0 {
		return zero, false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

// 使用标准库约束 cmp.Ordered
func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// 类型推断：调用时一般不需要写 [T, U]
func Map[T, U any](in []T, f func(T) U) []U {
	out := make([]U, 0, len(in))
	for _, v := range in {
		out = append(out, f(v))
	}
	return out
}

func main() {
	fmt.Println("Sum[int]:", Sum([]int{1, 2, 3}), "| Sum[float64]:", Sum([]float64{1.5, 2.5}))

	s := &Stack[string]{}
	s.Push("a")
	s.Push("b")
	v1, _ := s.Pop()
	v2, _ := s.Pop()
	_, ok := s.Pop()
	fmt.Println("Stack 弹出:", v1, v2, "| 空栈 ok =", ok)

	fmt.Println("Max:", Max(3, 7), Max("a", "z"), Max(1.5, 1.2))

	names := []string{"go", "rust", "c"}
	fmt.Println("Map 求长度:", Map(names, func(s string) int { return len(s) }))

	nums := []int{5, 2, 9}
	slices.Sort(nums)
	fmt.Println("标准库 slices:", nums, "| Max:", slices.Max(nums), "| Contains:", slices.Contains(names, "go"))
}
