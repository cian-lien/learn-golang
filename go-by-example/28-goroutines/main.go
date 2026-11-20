package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("后台任务已完成")
	}()
	fmt.Println("主函数已启动")
	wg.Wait()
}
