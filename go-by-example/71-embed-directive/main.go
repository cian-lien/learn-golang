package main

import (
	_ "embed"
	"fmt"
)

//go:embed menu.txt
var menu string

func main() {
	fmt.Print(menu)
}
