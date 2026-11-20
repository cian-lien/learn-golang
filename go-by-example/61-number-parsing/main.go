package main

import (
	"fmt"
	"strconv"
)

func main() {
	n, err := strconv.Atoi("42")
	if err != nil {
		panic(err)
	}
	hex, err := strconv.ParseInt("0x2a", 0, 64)
	if err != nil {
		panic(err)
	}
	_, invalid := strconv.Atoi("4x")
	fmt.Println(n, hex, invalid != nil)
}
