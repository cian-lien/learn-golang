package main

import (
	"fmt"
)

func newCounter() func() int {
	n := 0
	return func() int {
		n++
		return n
	}
}

func main() {
	a, b := newCounter(), newCounter()
	fmt.Println(a(), a(), b())
}
