package main

import (
	"encoding/xml"
	"fmt"
)

type Order struct {
	XMLName xml.Name `xml:"order"`
	ID      string   `xml:"id,attr"`
	Count   int      `xml:"count"`
}

func main() {
	var order Order
	if err := xml.Unmarshal([]byte(`<order id="A7"><count>2</count></order>`), &order); err != nil {
		panic(err)
	}
	order.Count++
	out, err := xml.Marshal(order)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
}
