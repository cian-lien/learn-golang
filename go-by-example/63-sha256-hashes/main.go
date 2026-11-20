package main

import (
	"crypto/sha256"
	"fmt"
)

func main() {
	data := []byte("go study notes")
	digest := sha256.Sum256(data)
	fmt.Println(len(digest))
	fmt.Printf("%x\n", digest)
}
