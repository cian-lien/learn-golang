package main

import (
	"fmt"
)

type Status int

const (
	Pending Status = iota
	Running
	Done
)

func main() {
	s := Running
	fmt.Println(Pending, Running, Done)
	fmt.Println(s == Running)
}
