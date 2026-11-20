package main

import (
	"fmt"
)

func main() {
	const secondsPerMinute = 60
	const minutes = 3
	fmt.Println(minutes * secondsPerMinute)
	fmt.Println(1.5 * secondsPerMinute)
}
