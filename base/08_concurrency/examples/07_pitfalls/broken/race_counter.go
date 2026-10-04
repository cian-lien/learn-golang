//go:build ignore

// 故意制造数据竞争的示例（注意文件开头的 build 标签，它不会参与 go build ./...）。
//
// 运行：
//
//	go run -race ./base/08_concurrency/examples/07_pitfalls/broken/race_counter.go
//
// 预期：打印 WARNING: DATA RACE，进程退出码为 66。
package main

import (
	"fmt"
	"sync"
)

func main() {
	var total int
	var wg sync.WaitGroup

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100_000; j++ {
				total++ // 读取、加一、写回，三步之间会被其它 goroutine 插进来
			}
		}()
	}
	wg.Wait()

	fmt.Println("结果:", total, "期望: 800000")
	fmt.Println("少掉的那些就是被覆盖掉的更新 —— 这就是数据竞争（data race）")
}
