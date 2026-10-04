package calc

import "errors"

// ErrDivideByZero 是导出的哨兵错误，便于调用方用 errors.Is 判断
var ErrDivideByZero = errors.New("除以零")

// Add 求和
func Add(a, b int) int { return a + b }

// Divide 除法，除数为 0 时返回错误
func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivideByZero
	}
	return a / b, nil
}
