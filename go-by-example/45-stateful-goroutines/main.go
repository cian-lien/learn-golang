package main

import (
	"fmt"
)

type Deposit struct {
	Amount int
	Reply  chan int
}

func ledger(requests <-chan Deposit, done chan<- struct{}) {
	defer close(done)
	balance := 0
	for req := range requests {
		balance += req.Amount
		req.Reply <- balance
	}
}

func main() {
	requests := make(chan Deposit)
	done := make(chan struct{})
	go ledger(requests, done)
	reply := make(chan int)
	for _, amount := range []int{10, 7} {
		requests <- Deposit{Amount: amount, Reply: reply}
		fmt.Println(<-reply)
	}
	close(requests)
	<-done
}
