package main

import (
	"fmt"
)

func main() {
	prices := make(chan int)
	go func() {
		defer close(prices)
		for _, price := range []int{6, 7, 8} {
			prices <- price
		}
	}()
	sum := 0
	for price := range prices {
		sum += price
	}
	fmt.Println(sum)
}
