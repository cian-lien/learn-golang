package main

import (
	"fmt"
	"os"
)

func main() {
	dir, err := os.MkdirTemp("", "go-draft-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	file, err := os.CreateTemp(dir, "draft-*.txt")
	if err != nil {
		panic(err)
	}
	if err := file.Close(); err != nil {
		panic(err)
	}
	info, err := os.Stat(file.Name())
	if err != nil {
		panic(err)
	}
	fmt.Println("临时文件已创建，字节数", info.Size())
}
