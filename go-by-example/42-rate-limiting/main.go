package main

import (
	"fmt"
	"time"
)

func main() {
	permits := time.NewTicker(10 * time.Millisecond)
	defer permits.Stop()
	for _, job := range []string{"甲", "乙", "丙"} {
		<-permits.C
		fmt.Println("开始处理", job)
	}
}
