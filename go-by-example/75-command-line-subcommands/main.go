package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "用法：tasks add -title 名称 | tasks list")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "add":
		add := flag.NewFlagSet("add", flag.ExitOnError)
		title := add.String("title", "", "待办名称")
		add.Parse(os.Args[2:])
		if *title == "" {
			fmt.Fprintln(os.Stderr, "请提供 -title")
			os.Exit(2)
		}
		fmt.Println("收到新任务", *title)
	case "list":
		fmt.Println("当前任务清单为空")
	default:
		fmt.Fprintln(os.Stderr, "未知子命令", os.Args[1])
		os.Exit(2)
	}
}
