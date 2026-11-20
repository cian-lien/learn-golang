package main

import (
	"fmt"
)

func main() {
	score := 82
	switch {
	case score >= 90:
		fmt.Println("优秀")
	case score >= 60:
		fmt.Println("通过")
	default:
		fmt.Println("继续练习")
	}
}
