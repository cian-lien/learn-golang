package main

import (
	"context"
	"runtime"
	"testing"
	"time"
)

// TestNoGoroutineLeak 演示如何在测试里断言「没有泄漏」。
//
// 生产项目推荐直接用 go.uber.org/goleak，它会忽略运行时自身产生的 goroutine，
// 断言更严格也更稳定：
//
//	func TestMain(m *testing.M) {
//	    goleak.VerifyTestMain(m)
//	}
func TestNoGoroutineLeak(t *testing.T) {
	base := runtime.NumGoroutine()

	done := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())

	go func() { // 有明确退出路径的 worker
		defer close(done)
		<-ctx.Done()
	}()

	cancel()
	<-done

	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > base && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if got := runtime.NumGoroutine(); got > base {
		t.Fatalf("疑似 goroutine 泄漏: 基线 %d, 现在 %d", base, got)
	}
}
