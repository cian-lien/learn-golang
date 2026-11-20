package main

import (
	"errors"
	"fmt"
)

var errOutOfStock = errors.New("库存不足")

func reserve(n int) error {
	if n > 3 {
		return fmt.Errorf("预订 %d 件商品: %w", n, errOutOfStock)
	}
	return nil
}

func main() {
	if err := reserve(5); err != nil {
		fmt.Println(err)
		fmt.Println(errors.Is(err, errOutOfStock))
	}
}
