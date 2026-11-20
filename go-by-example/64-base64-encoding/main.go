package main

import (
	"encoding/base64"
	"fmt"
)

func main() {
	encoded := base64.StdEncoding.EncodeToString([]byte("一杯茶"))
	fmt.Println(encoded)
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(decoded))
}
