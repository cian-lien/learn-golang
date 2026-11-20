package main

import (
	"fmt"
)

func main() {
	area := make(chan int)
	go func() {
		area <- 3 * 4
	}()
	result := <-area
	fmt.Println("面积", result)
}
