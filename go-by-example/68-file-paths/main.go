package main

import (
	"fmt"
	"path/filepath"
)

func main() {
	path := filepath.Join("notes", "go", "day1.md")
	fmt.Println(path)
	fmt.Println(filepath.Base(path))
	fmt.Println(filepath.Ext(path))
}
