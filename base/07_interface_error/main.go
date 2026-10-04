package main

import (
	"errors"
	"fmt"
)

// 接口是隐式实现的：只要方法集满足就算实现
type Shape interface {
	Area() float64
}

type Rect struct{ W, H float64 }

func (r Rect) Area() float64 { return r.W * r.H }

type Circle struct{ R float64 }

func (c Circle) Area() float64 { return 3.14159 * c.R * c.R }

// 自定义错误类型
type MyErr struct{ Code int }

func (e *MyErr) Error() string { return fmt.Sprintf("自定义错误 code=%d", e.Code) }

// 哨兵错误：用于 errors.Is 比对
var ErrNotFound = errors.New("not found")

func find(id int) (string, error) {
	switch {
	case id < 0:
		return "", fmt.Errorf("非法 id %d: %w", id, ErrNotFound) // %w 包装，保留错误链
	case id == 0:
		return "", ErrNotFound
	default:
		return "item", nil
	}
}

func main() {
	shapes := []Shape{Rect{3, 4}, Circle{1}}
	for _, s := range shapes {
		fmt.Printf("%T 面积=%.2f\n", s, s.Area())
	}

	// 类型断言：value, ok := iface.(T)
	var any1 any = "hello"
	if str, ok := any1.(string); ok {
		fmt.Println("断言成功:", str)
	}
	if n, ok := any1.(int); !ok {
		fmt.Println("断言失败 → n =", n, ", ok =", ok) // 失败时得到零值
	}

	// type switch
	for _, v := range []any{1, "a", 1.5, true, nil, []int{1}, Rect{1, 2}} {
		switch t := v.(type) {
		case nil:
			fmt.Println("  type switch: nil")
		case int:
			fmt.Printf("  type switch: int %d\n", t)
		case string:
			fmt.Printf("  type switch: string %q\n", t)
		case float64:
			fmt.Printf("  type switch: float64 %v\n", t)
		case bool:
			fmt.Printf("  type switch: bool %v\n", t)
		default:
			fmt.Printf("  type switch: 其他 %T %v\n", t, t)
		}
	}

	// 错误链：errors.Is 判断 sentinel，errors.As 取出具体类型
	_, err := find(0)
	fmt.Println("哨兵错误:", err, "| errors.Is:", errors.Is(err, ErrNotFound))

	_, err = find(-5)
	fmt.Println("包装错误:", err)
	fmt.Println("  errors.Is 仍然命中:", errors.Is(err, ErrNotFound), "| Unwrap:", errors.Unwrap(err))

	_, err = find(1)
	fmt.Println("正常返回 err == nil:", err == nil)

	var my *MyErr
	err = &MyErr{Code: 404}
	if errors.As(err, &my) {
		fmt.Println("errors.As 取出 code:", my.Code)
	}

	// 易错点：nil 指针装进接口后，接口本身不是 nil
	var e error
	var nilPtr *MyErr
	fmt.Println("接口 nil 判断 → e == nil:", e == nil, "| error(nilPtr) == nil:", error(nilPtr) == nil)
}
