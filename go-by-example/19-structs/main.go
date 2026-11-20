package main

import (
	"fmt"
)

type Task struct {
	Title string
	Done  bool
}

func main() {
	task := Task{Title: "浇花"}
	fmt.Println(task.Title, task.Done)
	task.Done = true
	fmt.Println(task.Title, task.Done)
}
