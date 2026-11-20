package main

import (
	"fmt"
)

type Author struct{ Name string }

func (a Author) Byline() string { return "作者 " + a.Name }

type Article struct {
	Author
	Title string
}

func main() {
	article := Article{Author: Author{Name: "林"}, Title: "春天"}
	fmt.Println(article.Title, article.Byline())
	fmt.Println(article.Author.Name)
}
