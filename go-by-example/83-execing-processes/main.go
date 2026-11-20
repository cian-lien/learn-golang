package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	path, err := exec.LookPath("echo")
	if err != nil {
		panic(err)
	}
	defer fmt.Println("旧程序执行收尾")
	if err := syscall.Exec(path, []string{"echo", "新程序开始运行"}, os.Environ()); err != nil {
		panic(err)
	}
}
