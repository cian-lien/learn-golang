package main

import (
	"fmt"
)

type Wallet struct{ Balance int }

func (w *Wallet) Deposit(amount int) {
	w.Balance += amount
}

func main() {
	w := Wallet{Balance: 10}
	w.Deposit(5)
	fmt.Println(w.Balance)
}
