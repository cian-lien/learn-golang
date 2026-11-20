package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	fmt.Println("正在等待中断信号，请按 Ctrl+C")
	<-ctx.Done()
	fmt.Println("收到中断，清理完成")
}
