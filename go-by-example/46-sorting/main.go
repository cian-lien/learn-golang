package main

import (
	"fmt"
	"slices"
)

func main() {
	scores := []int{28, 7, 14}
	fmt.Println(slices.IsSorted(scores))
	slices.Sort(scores)
	fmt.Println(scores)
}
