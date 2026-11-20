package main

import (
	"fmt"
	"math/rand/v2"
)

func main() {
	dice := rand.New(rand.NewPCG(12, 34))
	for range 3 {
		fmt.Println(dice.IntN(6) + 1)
	}
}
