package main

import (
	"fmt"
)

func main() {
	scores := [3]int{70, 80, 90}
	copied := scores
	copied[0] = 100
	fmt.Println(scores)
	fmt.Println(copied)
}
