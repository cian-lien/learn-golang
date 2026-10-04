// Command graceful-demo 演示一个 Go Web 服务从收到 SIGTERM 到进程退出的完整时间线。
//
// 运行（在 go/ 目录下）：
//
//	go run ./13_graceful_shutdown
//
// 流程：起服务 → 发一个慢请求和一个 SSE 流 → 给自己发 SIGTERM →
// 观察探针翻转、在途请求收尾、长连接主动断开、钩子逆序执行、新连接被拒。
//
// 注意 syscall.Kill 只在类 Unix 系统上可用。
package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"gobasics/13_graceful_shutdown/graceful"
)

func main() {
	start := time.Now()
	log := newLogger(start)

	// 服务实例要在路由之前建好（handler 里要用 lc.Draining()），
	// 所以先声明、再在闭包里引用，最后赋值。
	var lc *graceful.Lifecycle

	mux := http.NewServeMux()
	mux.HandleFunc("/fast", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "ok")
	})
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		d := 1200 * time.Millisecond
		if v := r.URL.Query().Get("ms"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				d = time.Duration(n) * time.Millisecond
			}
		}
		select {
		case <-time.After(d):
			fmt.Fprintf(w, "处理完成，耗时 %v", d)
		case <-r.Context().Done():
			// 客户端断开，或者排空超时后走了 srv.Close()：这条是强制中断的痕迹。
			fmt.Fprint(w, "请求被强制中断")
		}
	})
	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		rc := http.NewResponseController(w)
		for i := 1; ; i++ {
			select {
			case <-r.Context().Done(): // 客户端主动断开
				return
			case <-lc.Draining(): // 进程要退出了：主动收尾，别让 Shutdown 干等
				fmt.Fprint(w, "event: closing\ndata: 服务正在平滑关机\n\n")
				_ = rc.Flush()
				return
			case <-time.After(150 * time.Millisecond):
				fmt.Fprintf(w, "data: 第 %d 条推送\n\n", i)
				if err := rc.Flush(); err != nil { // 依赖 trackedWriter.Unwrap
					return
				}
			}
		}
	})

	lc = graceful.New(graceful.Config{
		Handler:       mux,
		Logger:        log,
		PreDrainDelay: 400 * time.Millisecond, // 阶段一：等 LB 摘流量
		DrainTimeout:  3 * time.Second,        // 阶段二上限：等完在途请求
		HookTimeout:   2 * time.Second,        // 阶段四预算：收尾钩子
		// RejectNewRequests 默认 false：阶段一还在正常服务，
		// 因为在 LB 真正摘掉本实例之前，客户端并不知道要换台机器。
	})

	// 钩子按注册顺序的逆序执行：数据库先注册，遥测后注册 → 退出时先 flush 遥测。
	lc.OnShutdown("关闭数据库连接池", func(ctx context.Context) error {
		return work(ctx, 80*time.Millisecond)
	})
	lc.OnShutdown("flush 遥测并关闭 exporter", func(ctx context.Context) error {
		return work(ctx, 60*time.Millisecond)
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Error("监听失败", "error", err)
		os.Exit(1)
	}
	base := "http://" + ln.Addr().String()
	log.Info("① 服务启动", "addr", base)

	// Ctrl-C / SIGTERM 取消 ctx，Serve 内部随即开始四阶段排空。
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() { serveErr <- lc.Serve(ctx, ln) }()

	client := &http.Client{Timeout: 10 * time.Second}
	waitReady(client, base, log)

	slow := make(chan result, 1)
	stream := make(chan result, 1)
	go func() { slow <- get(client, base+"/slow?ms=1200") }()
	go func() { stream <- get(client, base+"/stream") }()
	time.Sleep(150 * time.Millisecond)
	log.Info("② 一个耗时 1.2s 的请求 + 一条 SSE 流正在处理中",
		"in_flight", lc.InFlight())

	log.Info("③ 模拟运维发出 SIGTERM（真实场景由部署系统发送）")
	if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
		log.Error("发信号失败", "error", err)
		os.Exit(1)
	}

	// 阶段一窗口内观察：探针翻转，但普通请求照常成功。
	time.Sleep(60 * time.Millisecond)
	log.Info("④ 探针表现",
		"readyz", statusLine(client, base+"/readyz"),
		"healthz", statusLine(client, base+"/healthz"),
		"state", lc.State().String())
	log.Info("⑤ 阶段一仍在服务新请求（LB 摘流量需要时间）",
		"fast", statusLine(client, base+"/fast"))

	streamRes := <-stream
	log.Info("⑥ SSE 长连接在阶段一收到 Draining() 后主动收尾",
		"events", strings.Count(streamRes.body, "data:"),
		"elapsed", streamRes.elapsed)

	slowRes := <-slow
	log.Info("⑦ 在途的慢请求被完整处理完（没有 502/连接重置）",
		"status", slowRes.code, "body", slowRes.body, "elapsed", slowRes.elapsed)

	if err := <-serveErr; err != nil {
		log.Error("服务异常退出", "error", err)
		os.Exit(1)
	}

	// 监听套接字已经关闭：新连接一定失败。
	conn, dialErr := net.DialTimeout("tcp", ln.Addr().String(), 300*time.Millisecond)
	if dialErr == nil {
		conn.Close()
	}
	log.Info("⑧ 关机完成",
		"state", lc.State().String(),
		"dial_error", fmt.Sprint(dialErr),
		"total", time.Since(start).Round(time.Millisecond))
}

type result struct {
	code    int
	body    string
	err     error
	elapsed time.Duration
}

func get(client *http.Client, url string) result {
	begin := time.Now()
	resp, err := client.Get(url)
	if err != nil {
		return result{err: err, elapsed: time.Since(begin).Round(time.Millisecond)}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	res := result{code: resp.StatusCode, body: strings.TrimSpace(string(body)),
		elapsed: time.Since(begin).Round(time.Millisecond)}
	if err != nil {
		res.err = err
	}
	return res
}

// statusLine 把一次 GET 压成 "200 {...}" 这种便于打日志的一行。
func statusLine(client *http.Client, url string) string {
	res := get(client, url)
	if res.err != nil {
		return "error: " + res.err.Error()
	}
	return fmt.Sprintf("%d %s", res.code, res.body)
}

func waitReady(client *http.Client, base string, log *slog.Logger) {
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := client.Get(base + "/readyz")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	log.Warn("等待就绪超时")
}

func work(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

// newLogger 把 slog 的时间字段换成「相对启动的毫秒数」，让时间线一目了然。
func newLogger(start time.Time) *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey && len(groups) == 0 {
				return slog.String("t", fmt.Sprintf("+%04dms", time.Since(start).Milliseconds()))
			}
			return a
		},
	}))
}
