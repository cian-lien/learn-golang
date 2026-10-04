package calc

import (
	"errors"
	"testing"
)

func TestAdd(t *testing.T) {
	cases := []struct {
		name string
		a, b int
		want int
	}{
		{"正数", 1, 2, 3},
		{"负数", -1, -2, -3},
		{"零", 0, 0, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Add(c.a, c.b); got != c.want {
				t.Errorf("Add(%d, %d) = %d, 想要 %d", c.a, c.b, got, c.want)
			}
		})
	}
}

func TestDivide(t *testing.T) {
	got, err := Divide(10, 4)
	if err != nil {
		t.Fatalf("意外错误: %v", err)
	}
	if got != 2.5 {
		t.Errorf("Divide(10, 4) = %v, 想要 2.5", got)
	}

	if _, err := Divide(1, 0); !errors.Is(err, ErrDivideByZero) {
		t.Errorf("想要 ErrDivideByZero, 实际得到 %v", err)
	}
}

func BenchmarkAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Add(1, 2)
	}
}
