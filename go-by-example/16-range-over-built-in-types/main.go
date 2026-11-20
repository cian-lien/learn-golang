package main

import (
	"fmt"
)

func main() {
	for i, task := range []string{"读书", "跑步"} {
		fmt.Println(i, task)
	}
	for i := range 3 {
		fmt.Println(i)
	}
}
