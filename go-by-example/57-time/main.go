package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Date(2026, time.April, 3, 9, 30, 0, 0, time.UTC)
	end := start.Add(45 * time.Minute)
	fmt.Println(end.Format("15:04"))
	fmt.Println(end.Sub(start))
}
