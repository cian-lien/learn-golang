// 第 4 章：Mutex、RWMutex、atomic、Once、Pool
//
// 运行（在 go/ 目录下执行）：
//
//	go run ./base/08_concurrency/examples/04_sync_atomic
//	go run -race ./base/08_concurrency/examples/04_sync_atomic   # 本章代码应该是干净的
package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	workers   = 8
	perWorker = 50_000
)

func main() {
	demoMutex()
	demoRWMutex()
	demoAtomic()
	demoCAS()
	demoOnce()
	demoPool()
	demoSharded()
}

// counterMutex 用互斥锁保护共享变量。
func counterMutex() int64 {
	var (
		mu    sync.Mutex
		total int64
		wg    sync.WaitGroup
	)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				mu.Lock()
				total++ // 临界区只放必需的这一行
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	return total
}

func demoMutex() {
	fmt.Println("[1] Mutex：保护临界区")
	start := time.Now()
	got, want := counterMutex(), int64(workers*perWorker)
	fmt.Printf("    结果 = %d，期望 = %d，正确 = %v，耗时 %v\n",
		got, want, got == want, time.Since(start).Round(time.Millisecond))
	fmt.Println("    对照实验（会报 data race）：go run -race ./base/08_concurrency/examples/07_pitfalls/broken/race_counter.go")
	fmt.Println("    锁的粒度要点：临界区只放必需的几行，绝不在锁里做 IO 或调用可能阻塞的函数")
}

// cache 是读多写少的场景，用 RWMutex 允许多个读并发进行。
type cache struct {
	mu sync.RWMutex
	m  map[string]string
}

func newCache() *cache { return &cache{m: make(map[string]string)} }

func (c *cache) Get(k string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.m[k] // 多个 goroutine 可以同时进入读锁
	return v, ok
}

func (c *cache) Set(k, v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[k] = v
}

func demoRWMutex() {
	fmt.Println("\n[2] RWMutex：读多写少")

	c := newCache()
	c.Set("go", "goroutine")

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = c.Get("go") // 5 个读者可以同时在读锁里
		}()
	}
	wg.Wait()

	v, _ := c.Get("go")
	fmt.Println("    读到的值:", v)
	fmt.Println("    写锁要等所有读锁释放；读锁里做耗时操作会拖慢写，写非常频繁时 RWMutex 可能比 Mutex 更慢")
}

// counterAtomic 用原子变量做同样的计数。
func counterAtomic() int64 {
	var (
		total atomic.Int64
		wg    sync.WaitGroup
	)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				total.Add(1)
			}
		}()
	}

	wg.Wait()
	return total.Load()
}

func demoAtomic() {
	fmt.Println("\n[3] atomic：单个变量的无锁更新")
	got, want := counterAtomic(), int64(workers*perWorker)
	fmt.Printf("    结果 = %d，期望 = %d，正确 = %v\n", got, want, got == want)
	fmt.Println("    适合：计数器、开关标志、状态机、指针整体替换（atomic.Pointer）")
	fmt.Println("    不适合：需要多个变量保持一致（那还是得用锁）")
}

// takeQuota 用 CAS 循环实现名额扣减：无锁，且绝不会超发。
func takeQuota(quota *atomic.Int64) (left int64, ok bool) {
	for {
		cur := quota.Load()
		if cur <= 0 {
			return 0, false
		}
		if quota.CompareAndSwap(cur, cur-1) { // 期间没人改过，才算扣减成功
			return cur - 1, true
		}
		// 失败说明被别人抢先改了，重读当前值后重试
	}
}

func demoCAS() {
	fmt.Println("\n[4] CAS：读-改-写必须原子完成的场景")

	var quota atomic.Int64
	quota.Store(3) // 一共只有 3 个名额

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results []string
	)

	for i := 1; i <= 6; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			left, ok := takeQuota(&quota)

			mu.Lock()
			defer mu.Unlock()
			if ok {
				results = append(results, fmt.Sprintf("请求%d 拿到名额(剩余%d)", i, left))
			} else {
				results = append(results, fmt.Sprintf("请求%d 被拒", i))
			}
		}(i)
	}
	wg.Wait()

	sort.Strings(results)
	fmt.Println("    " + strings.Join(results, "，"))
	fmt.Println("    6 个并发请求，只有 3 个成功，没有超发")
}

type config struct{ name string }

var (
	configOnce sync.Once
	configVal  *config
)

// loadConfig 无论被多少个 goroutine 同时调用，初始化只执行一次。
func loadConfig() *config {
	configOnce.Do(func() {
		time.Sleep(20 * time.Millisecond) // 模拟昂贵初始化：读配置、建连接池
		configVal = &config{name: "prod"}
	})
	return configVal
}

func demoOnce() {
	fmt.Println("\n[5] sync.Once：并发调用也只执行一次")

	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = loadConfig()
		}()
	}
	wg.Wait()

	fmt.Printf("    5 个 goroutine 并发获取，总耗时 %v（≈ 单次初始化耗时）\n",
		time.Since(start).Round(5*time.Millisecond))
	fmt.Println("    其余 goroutine 会阻塞在 Once.Do 上，等初始化完成后拿到同一个指针")
}

func demoPool() {
	fmt.Println("\n[6] sync.Pool：复用临时对象，降低 GC 压力")

	pool := &sync.Pool{New: func() any { return make([]byte, 0, 4096) }}

	buf := pool.Get().([]byte)
	buf = append(buf, "hello"...)
	fmt.Printf("    取出的缓冲: %q（cap = %d）\n", buf, cap(buf))
	pool.Put(buf[:0]) // 归还前把长度清零，避免把内容带给下一个使用者

	buf2 := pool.Get().([]byte)
	fmt.Printf("    再次取出: len = %d，cap = %d（内容已清理）\n", len(buf2), cap(buf2))
	fmt.Println("    注意：池里的对象随时可能被 GC 回收，不能用来保存状态")
}

// shardedCounter 把一把大锁拆成 N 把小锁，用空间换竞争。
type shardedCounter struct {
	shards [16]struct {
		mu sync.Mutex
		n  int64
	}
}

func (c *shardedCounter) Inc(key int) {
	s := &c.shards[uint64(key)%uint64(len(c.shards))]
	s.mu.Lock()
	s.n++
	s.mu.Unlock()
}

func (c *shardedCounter) Total() int64 {
	var total int64
	for i := range c.shards {
		s := &c.shards[i]
		s.mu.Lock()
		total += s.n
		s.mu.Unlock()
	}
	return total
}

func demoSharded() {
	fmt.Println("\n[7] 分片锁：用空间换竞争")

	var c shardedCounter
	var wg sync.WaitGroup

	start := time.Now()
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				c.Inc(id) // 不同 goroutine 落在不同分片，冲突大幅减少
			}
		}(i)
	}
	wg.Wait()

	fmt.Printf("    结果 = %d，期望 = %d，耗时 %v\n",
		c.Total(), int64(workers*perWorker), time.Since(start).Round(time.Millisecond))
}
