package main

import (
	"fmt"
	"iter"
)

func countdown(from int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for n := from; n > 0; n-- {
			if !yield(n) {
				return
			}
		}
	}
}

func main() {
	for n := range countdown(5) {
		fmt.Println(n)
		if n == 3 {
			break
		}
	}
}
