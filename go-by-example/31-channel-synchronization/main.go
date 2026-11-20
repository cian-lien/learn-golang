package main

import (
	"fmt"
)

func main() {
	done := make(chan struct{})
	go func() {
		fmt.Println("保存完成")
		close(done)
	}()
	<-done
	fmt.Println("可以退出")
}
