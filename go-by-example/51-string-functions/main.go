package main

import (
	"fmt"
	"strings"
)

func main() {
	tags := strings.Split(" go, web ,cli ", ",")
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}
	fmt.Println(strings.Join(tags, " / "))
	fmt.Println(strings.HasPrefix(tags[0], "g"))
}
