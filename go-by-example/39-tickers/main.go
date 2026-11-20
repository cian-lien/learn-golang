package main

import (
	"fmt"
	"time"
)

func main() {
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for n := 1; n <= 3; n++ {
		<-ticker.C
		fmt.Println("检查", n)
	}
}
