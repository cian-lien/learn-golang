package main

import (
	"fmt"
	"os"
)

func main() {
	data, err := os.ReadFile("books.txt")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(data))
}
