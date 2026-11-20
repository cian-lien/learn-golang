package main

import (
	"fmt"
	"sync"
)

func main() {
	names := []string{"茶", "咖啡", "水"}
	labels := make([]string, len(names))
	var wg sync.WaitGroup
	for i, name := range names {
		wg.Go(func() { labels[i] = "饮品：" + name })
	}
	wg.Wait()
	fmt.Println(labels)
}
