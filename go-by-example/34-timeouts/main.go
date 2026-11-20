package main

import (
	"fmt"
	"time"
)

func main() {
	reply := make(chan string)
	select {
	case value := <-reply:
		fmt.Println(value)
	case <-time.After(20 * time.Millisecond):
		fmt.Println("等待回复超时")
	}
}
