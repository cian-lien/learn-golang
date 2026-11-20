package main

import (
	"fmt"
	"net/url"
)

func main() {
	u, err := url.Parse("https://example.org:8443/books?q=Go+Web")
	if err != nil {
		panic(err)
	}
	fmt.Println(u.Scheme, u.Hostname(), u.Port())
	fmt.Println(u.Path, u.Query().Get("q"))
}
