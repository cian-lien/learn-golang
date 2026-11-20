package main

import (
	"fmt"
)

func openBoard(width int) {
	if width <= 0 {
		panic("棋盘宽度必须为正数")
	}
	fmt.Println("创建棋盘", width)
}

func main() {
	defer fmt.Println("清理棋盘资源")
	openBoard(0)
}
