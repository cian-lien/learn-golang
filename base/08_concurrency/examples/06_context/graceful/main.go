// HTTP 服务的优雅退出示例。
//
// 运行（在 go/ 目录下执行）：
//
//	go run ./base/08_concurrency/examples/06_context/graceful
//
// 程序自己给自己发 SIGTERM，然后观察在途请求是否被处理完。
// 注意：syscall.Kill 只在类 Unix 系统上可用。
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(300 * time.Millisecond):
			fmt.Fprintln(w, "处理完成")
		case <-r.Context().Done():
			// Shutdown 不会打断在途请求；如果客户端断开，这里会被触发
			fmt.Fprintln(w, "客户端提前断开")
		}
	})

	// 监听随机端口，避免和本机其它服务冲突
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	srv := &http.Server{Handler: mux}

	go func() {
		// Shutdown 会让 Serve 返回 http.ErrServerClosed
		_ = srv.Serve(ln)
	}()
	fmt.Println("→ 服务已启动:", ln.Addr())

	// NotifyContext：收到 SIGINT/SIGTERM 时自动取消 ctx
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 发起一个慢请求，让它停在处理中
	response := make(chan string, 1)
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String() + "/slow")
		if err != nil {
			response <- "请求出错: " + err.Error()
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		response <- strings.TrimSpace(string(body))
	}()

	time.Sleep(50 * time.Millisecond) // 让请求进入处理逻辑
	fmt.Println("→ 模拟运维发出 SIGTERM（真实场景由部署系统发送）")
	if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
		log.Fatal(err)
	}

	<-ctx.Done()
	fmt.Println("→ 收到停止信号，开始优雅退出")

	start := time.Now()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("优雅退出失败:", err)
	}

	fmt.Printf("→ 服务已退出，等待在途请求耗时 %v\n", time.Since(start).Round(10*time.Millisecond))
	fmt.Println("→ 在途请求的结果:", <-response)
	fmt.Println("→ 对比：如果调用 srv.Close()，在途请求会被直接切断")
}
