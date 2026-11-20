package main

import (
	"fmt"
)

func publish(out chan<- string) {
	out <- "新书到馆"
}

func receive(in <-chan string) string {
	return <-in
}

func main() {
	news := make(chan string, 1)
	publish(news)
	fmt.Println(receive(news))
}
