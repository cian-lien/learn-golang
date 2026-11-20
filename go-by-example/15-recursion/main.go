package main

import (
	"fmt"
)

func sumTo(n int) int {
	if n <= 0 {
		return 0
	}
	return n + sumTo(n-1)
}

func main() {
	fmt.Println(sumTo(4))
}
