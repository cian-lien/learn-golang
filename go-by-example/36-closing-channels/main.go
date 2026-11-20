package main

import (
	"fmt"
)

func main() {
	queue := make(chan int, 1)
	queue <- 7
	close(queue)
	n, ok := <-queue
	fmt.Println(n, ok)
	n, ok = <-queue
	fmt.Println(n, ok)
}
