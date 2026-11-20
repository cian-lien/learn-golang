package main

import (
	"fmt"
)

func total(prices ...int) int {
	sum := 0
	for _, price := range prices {
		sum += price
	}
	return sum
}

func main() {
	fmt.Println(total(6, 8))
	cart := []int{3, 4, 5}
	fmt.Println(total(cart...))
}
