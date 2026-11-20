package main

import (
	"fmt"
)

func restock(stock *int, quantity int) {
	*stock += quantity
}

func main() {
	stock := 2
	restock(&stock, 3)
	fmt.Println(stock)
}
