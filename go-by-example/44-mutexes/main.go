package main

import (
	"fmt"
	"sync"
)

type Tickets struct {
	mu   sync.Mutex
	left int
}

func (t *Tickets) Book() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.left == 0 {
		return false
	}
	t.left--
	return true
}

func main() {
	tickets := Tickets{left: 2}
	var wg sync.WaitGroup
	booked := make([]bool, 3)
	for i := range booked {
		wg.Go(func() { booked[i] = tickets.Book() })
	}
	wg.Wait()
	count := 0
	for _, ok := range booked {
		if ok {
			count++
		}
	}
	fmt.Println("成功订票", count)
}
