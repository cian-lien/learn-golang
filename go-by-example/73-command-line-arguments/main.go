package main

import (
	"fmt"
	"os"
)

func main() {
	for i, arg := range os.Args[1:] {
		fmt.Printf("第 %d 个参数 = %q\n", i+1, arg)
	}
}
