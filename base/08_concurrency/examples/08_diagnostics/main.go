// 第 8 章：诊断与调优工具
//
// 运行（在 go/ 目录下执行）：
//
//	go run ./base/08_concurrency/examples/08_diagnostics
//	go test -bench . -benchmem ./base/08_concurrency/examples/08_diagnostics
package main

import (
	"bytes"
	"fmt"
	"net/http"
	_ "net/http/pprof" // 通过副作用注册 /debug/pprof/* 到 DefaultServeMux
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Printf("Go %s，GOMAXPROCS = %d，逻辑 CPU = %d\n",
		runtime.Version(), runtime.GOMAXPROCS(0), runtime.NumCPU())

	demoPprofServer()
	demoCPUProfile()
	demoGoroutineProfile()
	demoTrace()
	demoMemStats()
}

func demoPprofServer() {
	fmt.Println("\n[1] pprof HTTP 端点")

	go func() {
		_ = http.ListenAndServe("127.0.0.1:6060", nil) // 端口被占用时会静默失败，不影响其它演示
	}()

	fmt.Println("    已尝试启动 http://127.0.0.1:6060/debug/pprof/")
	fmt.Println("    go tool pprof http://127.0.0.1:6060/debug/pprof/profile?seconds=5   # CPU")
	fmt.Println("    go tool pprof http://127.0.0.1:6060/debug/pprof/heap              # 内存")
	fmt.Println("    curl 'http://127.0.0.1:6060/debug/pprof/goroutine?debug=2'        # 所有 goroutine 的栈")
}

// busyWork 制造 CPU 负载，让 profile 里有内容可看。
func busyWork(d time.Duration) {
	deadline := time.Now().Add(d)
	sum := 0
	for time.Now().Before(deadline) {
		for i := 0; i < 10_000; i++ {
			sum += i % 7
		}
	}
	_ = sum
}

func demoCPUProfile() {
	fmt.Println("\n[2] CPU profile：找热点函数")

	f, err := os.CreateTemp("", "cpu-*.pprof")
	if err != nil {
		fmt.Println("    创建临时文件失败:", err)
		return
	}
	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		fmt.Println("    启动失败:", err)
		return
	}
	busyWork(300 * time.Millisecond)
	pprof.StopCPUProfile()

	fmt.Printf("    profile 已写入 %s\n", f.Name())
	fmt.Printf("    go tool pprof -http=:8080 %s\n", f.Name())
}

func demoGoroutineProfile() {
	fmt.Println("\n[3] goroutine profile：排查泄漏最直接的手段")

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(200 * time.Millisecond)
		}()
	}
	time.Sleep(20 * time.Millisecond) // 让它们进入 sleep 状态

	var buf bytes.Buffer
	if err := pprof.Lookup("goroutine").WriteTo(&buf, 1); err != nil { // 1 = 带调用栈
		fmt.Println("    写入失败:", err)
		wg.Wait()
		return
	}

	firstLine := strings.SplitN(buf.String(), "\n", 2)[0]
	fmt.Printf("    当前 goroutine 数 = %d，profile 大小 = %d 字节\n", runtime.NumGoroutine(), buf.Len())
	fmt.Println("    profile 的第一行:", firstLine)

	wg.Wait()
}

// concurrentWork 制造可观测的阻塞与唤醒，方便在 trace 里看到调度。
func concurrentWork() {
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 3; j++ {
				time.Sleep(10 * time.Millisecond)
			}
		}()
	}
	wg.Wait()
}

func demoTrace() {
	fmt.Println("\n[4] runtime/trace：调度、GC、阻塞的时间线")

	f, err := os.CreateTemp("", "trace-*.out")
	if err != nil {
		fmt.Println("    创建临时文件失败:", err)
		return
	}
	defer f.Close()

	if err := trace.Start(f); err != nil {
		fmt.Println("    启动失败:", err)
		return
	}
	concurrentWork()
	trace.Stop()

	fmt.Printf("    trace 已写入 %s\n", f.Name())
	fmt.Printf("    go tool trace %s\n", f.Name())
}

func demoMemStats() {
	fmt.Println("\n[5] 运行期指标")

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Printf("    HeapAlloc = %d KiB，TotalAlloc = %d KiB，NumGC = %d，Goroutine = %d\n",
		m.HeapAlloc/1024, m.TotalAlloc/1024, m.NumGC, runtime.NumGoroutine())
	fmt.Println("    线上这些指标通常以 Prometheus 形式暴露：go_goroutines、go_memstats_*、go_gc_*")
	fmt.Println("    ↑ HeapAlloc 长期不降 + goroutine 数持续增长，就是需要看 profile 的信号")
}
