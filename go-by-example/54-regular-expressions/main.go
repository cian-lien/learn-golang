package main

import (
	"fmt"
	"regexp"
)

func main() {
	ticket := regexp.MustCompile(`\bT-[0-9]{3}\b`)
	line := "open T-015, close T-207, ignore T-8"
	fmt.Println(ticket.FindAllString(line, -1))
	fmt.Println(ticket.ReplaceAllString(line, "[工单]"))
}
