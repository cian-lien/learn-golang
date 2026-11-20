package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "TODO ") {
			fmt.Println(strings.TrimPrefix(line, "TODO "))
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
