package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	word := "A茶B"
	fmt.Println(len(word), utf8.RuneCountInString(word))
	for offset, r := range word {
		fmt.Printf("%d %c\n", offset, r)
	}
}
