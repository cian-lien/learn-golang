package main

import (
	"fmt"
)

func price(quantity, unitPrice int) int {
	return quantity * unitPrice
}

func main() {
	fmt.Println(price(3, 12))
}
