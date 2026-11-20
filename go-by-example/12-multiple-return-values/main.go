package main

import (
	"fmt"
)

func indexOf(names []string, wanted string) (int, bool) {
	for i, name := range names {
		if name == wanted {
			return i, true
		}
	}
	return -1, false
}

func main() {
	names := []string{"阿青", "小林"}
	fmt.Println(indexOf(names, "阿青"))
	fmt.Println(indexOf(names, "小周"))
}
