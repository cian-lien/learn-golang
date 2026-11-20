package main

import (
	"fmt"
)

func contains[T comparable](items []T, target T) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func main() {
	fmt.Println(contains([]int{4, 8}, 8))
	fmt.Println(contains([]string{"茶", "水"}, "奶"))
}
