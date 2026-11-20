package main

import (
	"fmt"
)

func runJob() {
	defer func() {
		if reason := recover(); reason != nil {
			fmt.Println("任务失败", reason)
		}
	}()
	panic("缺少配置")
}

func main() {
	runJob()
	fmt.Println("开始下一项任务")
}
