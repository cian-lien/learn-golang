package main

import (
	"fmt"
)

func main() {
	tasks := []string{"吃饭", "散步", "写代码"}
	firstTwo := tasks[:2]
	firstTwo[0] = "做饭"
	fmt.Println(tasks)
	tasks = append(tasks, "读书")
	fmt.Println(len(tasks), tasks)
}
