package main

import (
	"fmt"
)

func main() {
	mail := make(chan string, 1)
	chat := make(chan string, 1)
	mail <- "收到邮件"
	chat <- "收到消息"
	for range 2 {
		select {
		case msg := <-mail:
			fmt.Println(msg)
		case msg := <-chat:
			fmt.Println(msg)
		}
	}
}
