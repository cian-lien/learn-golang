package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	root, err := os.MkdirTemp("", "go-library-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(root)
	if err := os.MkdirAll(filepath.Join(root, "go", "examples"), 0700); err != nil {
		panic(err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		fmt.Println(entry.Name(), entry.IsDir())
	}
}
