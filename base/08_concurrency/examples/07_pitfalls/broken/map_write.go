//go:build ignore

// 并发读写同一个 map 的后果。
//
// 运行（程序会崩溃，这是故意的）：
//
//	go run ./base/08_concurrency/examples/07_pitfalls/broken/map_write.go
//
// 预期输出：fatal error: concurrent map writes
// 这类错误由运行时直接终止进程，recover 也没用。
package main

import (
	"fmt"
	"sync"
)

func main() {
	m := map[int]int{}
	var wg sync.WaitGroup

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 10_000; j++ {
				m[i] = j // map 不是并发安全的
			}
		}(i)
	}
	wg.Wait()

	fmt.Println("map 大小:", len(m))
}
