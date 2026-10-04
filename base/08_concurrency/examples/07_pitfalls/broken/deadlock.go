//go:build ignore

// 死锁示例：两个 goroutine 以相反的顺序加两把锁。
//
// 运行（程序会崩溃，这是故意的）：
//
//	go run ./base/08_concurrency/examples/07_pitfalls/broken/deadlock.go
//
// 预期输出：fatal error: all goroutines are asleep - deadlock!
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var (
		muA, muB sync.Mutex
		wg       sync.WaitGroup
	)

	wg.Add(2)

	go func() { // 顺序：A → B
		defer wg.Done()
		muA.Lock()
		defer muA.Unlock()

		time.Sleep(10 * time.Millisecond)

		muB.Lock()
		defer muB.Unlock()
		fmt.Println("goroutine 1 完成")
	}()

	go func() { // 顺序：B → A，正好相反
		defer wg.Done()
		muB.Lock()
		defer muB.Unlock()

		time.Sleep(10 * time.Millisecond)

		muA.Lock()
		defer muA.Unlock()
		fmt.Println("goroutine 2 完成")
	}()

	wg.Wait()
	fmt.Println("没崩溃说明这次恰好没死锁（时序问题，多跑几次就会出现）")
}
