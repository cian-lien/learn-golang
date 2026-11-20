package main

import (
	"fmt"
	"os"
)

func main() {
	defer fmt.Println("执行延迟清理")
	fmt.Println("任务未完成")
	os.Exit(3)
}
