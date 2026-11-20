package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	dir, err := os.MkdirTemp("", "go-note-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "note.txt")
	if err := os.WriteFile(path, []byte("明天读十页\n"), 0600); err != nil {
		panic(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(data))
}
