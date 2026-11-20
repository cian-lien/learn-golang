package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	var visits atomic.Int64
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			for range 100 {
				visits.Add(1)
			}
		})
	}
	wg.Wait()
	fmt.Println(visits.Load())
}
