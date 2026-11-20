package main

import (
	"fmt"
	"time"
)

func main() {
	moment := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	fmt.Println(moment.Unix())
	fmt.Println(moment.UnixMilli())
	fmt.Println(time.Unix(moment.Unix(), 0).UTC().Format(time.RFC3339))
}
