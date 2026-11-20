package main

import (
	"fmt"
	"sync"
)

func squareWorker(jobs <-chan int, results chan<- int) {
	for n := range jobs {
		results <- n * n
	}
}

func main() {
	jobs := make(chan int)
	results := make(chan int)
	var workers sync.WaitGroup
	for range 2 {
		workers.Go(func() { squareWorker(jobs, results) })
	}
	go func() {
		for n := 1; n <= 4; n++ {
			jobs <- n
		}
		close(jobs)
		workers.Wait()
		close(results)
	}()
	sum := 0
	for n := range results {
		sum += n
	}
	fmt.Println("平方和", sum)
}
