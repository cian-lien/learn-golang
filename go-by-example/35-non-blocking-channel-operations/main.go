package main

import (
	"fmt"
)

func main() {
	queue := make(chan string, 1)
	queue <- "已排队"
	select {
	case queue <- "新任务":
		fmt.Println("加入成功")
	default:
		fmt.Println("队列已满")
	}
	fmt.Println(<-queue)
}
