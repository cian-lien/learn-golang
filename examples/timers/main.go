// Timers：通过 channel 接收一次定时通知，也可以在通知送达前停止定时器。
// 参考：https://gobyexample.com/timers
//
// 在项目根目录运行：go run ./examples/timers
// 约 2 秒后依次输出 Timer 1 fired、Timer 2 stopped；不含编译时间。
// 完整学习说明见同目录 README.md。
package main

import (
	"fmt"
	"time"
)

func main() {
	// NewTimer 立即返回；计时从创建时开始，接收操作才会等待通知。
	timer1 := time.NewTimer(2 * time.Second)
	<-timer1.C
	fmt.Println("Timer 1 fired")

	timer2 := time.NewTimer(time.Second)
	// 自检：本例没有其他接收者，默认的 Go 1.23+ 语义下应能停止它。
	if !timer2.Stop() {
		panic("Timer 2 应在通知送达前成功停止")
	}
	fmt.Println("Timer 2 stopped")

	// 触发和 Stop 都不会关闭 C；本例不在它上面 range 或再次等待。
}
