package main

import (
	"fmt"
)

func main() {
	sum := 0
	for n := 1; n <= 5; n++ {
		if n%2 == 0 {
			continue
		}
		sum += n
	}
	fmt.Println(sum)
}
