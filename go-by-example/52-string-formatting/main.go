package main

import (
	"fmt"
)

func main() {
	label := fmt.Sprintf("商品 %03d | %q | %.2f 元", 7, "茶杯", 12.5)
	fmt.Println(label)
}
