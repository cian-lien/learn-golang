package main

import (
	"fmt"
	"time"
)

func main() {
	zone := time.FixedZone("UTC+8", 8*60*60)
	meeting, err := time.ParseInLocation("2006/01/02 15:04", "2026/04/03 09:30", zone)
	if err != nil {
		panic(err)
	}
	fmt.Println(meeting.Format(time.RFC3339))
	fmt.Println(meeting.UTC().Format(time.RFC3339))
}
