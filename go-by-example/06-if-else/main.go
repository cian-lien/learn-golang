package main

import (
	"fmt"
)

func main() {
	tasks := []string{"洗衣", "读书"}
	if n := len(tasks); n == 0 {
		fmt.Println("今天没有待办")
	} else {
		fmt.Printf("还有 %d 项待办\n", n)
	}
}
