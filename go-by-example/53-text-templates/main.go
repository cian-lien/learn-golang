package main

import (
	"os"
	"text/template"
)

func main() {
	const source = "采购清单\n{{range .}}- {{.}}\n{{end}}"
	tmpl := template.Must(template.New("list").Parse(source))
	if err := tmpl.Execute(os.Stdout, []string{"米", "茶"}); err != nil {
		panic(err)
	}
}
