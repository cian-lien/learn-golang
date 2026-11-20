package main

import (
	"flag"
	"fmt"
)

func main() {
	count := flag.Int("n", 1, "打印次数")
	flag.Parse()
	for range *count {
		fmt.Println(flag.Args())
	}
}
