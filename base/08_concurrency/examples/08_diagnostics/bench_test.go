package main

import (
	"sync"
	"sync/atomic"
	"testing"
)

// 这一组基准测试回答同一个问题：8 个 goroutine 抢着累加，哪种方式最快？
//
// 运行：
//
//	go test -bench . -benchmem ./base/08_concurrency/examples/08_diagnostics

func BenchmarkCounterMutex(b *testing.B) {
	var (
		mu sync.Mutex
		n  int64
	)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			n++
			mu.Unlock()
		}
	})
	_ = n
}

func BenchmarkCounterAtomic(b *testing.B) {
	var n atomic.Int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n.Add(1)
		}
	})
	_ = n.Load()
}

func BenchmarkCounterChannel(b *testing.B) {
	// 用容量 1 的 channel 当锁：语义正确，但开销通常是最大的。
	ch := make(chan struct{}, 1)
	var n int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ch <- struct{}{}
			n++
			<-ch
		}
	})
	_ = n
}

func BenchmarkCounterSharded(b *testing.B) {
	var shards [16]struct {
		mu sync.Mutex
		n  int64
	}
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			s := &shards[i%len(shards)]
			s.mu.Lock()
			s.n++
			s.mu.Unlock()
			i++
		}
	})
}
