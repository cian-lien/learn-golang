package main

import (
	"fmt"
)

type Titled interface{ Title() string }

type Book struct{ Name string }

func (b Book) Title() string { return b.Name }

type Film struct{ Name string }

func (f Film) Title() string { return f.Name }

func printTitle(item Titled) { fmt.Println(item.Title()) }

func main() {
	printTitle(Book{Name: "城南旧事"})
	printTitle(Film{Name: "千与千寻"})
}
