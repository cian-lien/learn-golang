package main

import (
	"fmt"
	"time"
)

func main() {
	timer := time.NewTimer(10 * time.Millisecond)
	<-timer.C
	fmt.Println("休息结束")

	reminder := time.NewTimer(time.Hour)
	fmt.Println("取消下一次提醒", reminder.Stop())
}
