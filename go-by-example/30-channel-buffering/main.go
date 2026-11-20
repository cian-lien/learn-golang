package main

import (
	"fmt"
)

func main() {
	queue := make(chan string, 2)
	queue <- "早餐"
	queue <- "午餐"
	fmt.Println(len(queue), cap(queue))
	fmt.Println(<-queue)
	fmt.Println(<-queue)
}
