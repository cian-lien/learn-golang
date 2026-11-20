package main

import (
	"fmt"
	"os"
)

func main() {
	topic, ok := os.LookupEnv("STUDY_TOPIC")
	if !ok {
		topic = "自由练习"
	}
	fmt.Println("今日主题", topic)
}
