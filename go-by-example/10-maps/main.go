package main

import (
	"fmt"
)

func main() {
	stock := map[string]int{"茶": 0}
	count, ok := stock["茶"]
	fmt.Println(count, ok)
	delete(stock, "茶")
	count, ok = stock["茶"]
	fmt.Println(count, ok)
}
