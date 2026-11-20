package main

import (
	"errors"
	"fmt"
)

type LimitError struct{ Limit int }

func (e *LimitError) Error() string {
	return fmt.Sprintf("每次最多借 %d 本", e.Limit)
}

func borrow(n int) error {
	if n > 2 {
		return &LimitError{Limit: 2}
	}
	return nil
}

func main() {
	err := borrow(3)
	var limit *LimitError
	if errors.As(err, &limit) {
		fmt.Println(limit.Error())
		fmt.Println("上限", limit.Limit)
	}
}
