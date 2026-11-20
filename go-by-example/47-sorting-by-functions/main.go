package main

import (
	"cmp"
	"fmt"
	"slices"
)

type Task struct {
	Name    string
	Minutes int
}

func main() {
	tasks := []Task{{"read", 20}, {"walk", 10}, {"cook", 20}}
	slices.SortFunc(tasks, func(a, b Task) int {
		if order := cmp.Compare(a.Minutes, b.Minutes); order != 0 {
			return order
		}
		return cmp.Compare(a.Name, b.Name)
	})
	fmt.Println(tasks)
}
