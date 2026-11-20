package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func main() {
	out, err := exec.Command("go", "version").Output()
	if err != nil {
		panic(err)
	}
	fmt.Println(strings.TrimSpace(string(out)))
	fmt.Println("子进程执行完毕")
}
