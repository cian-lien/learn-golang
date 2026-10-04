package graceful

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

// ── 测试脚手架 ────────────────────────────────────────────────────────────

type testServer struct {
	lc   *Lifecycle
	ln   net.Listener
	addr string
}

// start 在随机端口上起一个 Lifecycle，并在测试结束时关干净。
func start(t *testing.T, cfg Config) *testServer {
	t.Helper()
	if cfg.DrainTimeout == 0 {
		cfg.DrainTimeout = 2 * time.Second
	}
	if cfg.Logger == nil {
		cfg.Logger = quietLogger()
	}

	lc := New(cfg)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("监听失败: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- lc.Serve(ctx, ln) }()

	t.Cleanup(func() {
		cancel()
		select {
		case err := <-errCh:
			// 用例里可能已经因为超时被强杀，Serve 会把超时错误带回来，这是预期行为。
			if err != nil && !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("Serve 返回非预期错误: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Error("Serve 没有在 5s 内返回")
		}
	})

	ts := &testServer{lc: lc, ln: ln, addr: "http://" + ln.Addr().String()}
	ts.waitReady(t)
	return ts
}

func (ts *testServer) waitReady(t *testing.T) {
	t.Helper()
	client := &http.Client{Timeout: time.Second}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := client.Get(ts.addr + "/readyz")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("服务没有变为 ready")
}

type httpResult struct {
	code   int
	body   string
	header http.Header
	// close 对应 resp.Close：服务端要求这条连接不再复用。
	// 注意 Go 客户端会把响应里的 Connection 头删掉（它是 hop-by-hop 头，
	// 见 net/http transfer.go 的 shouldClose(..., removeCloseHeader=true)），
	// 所以「服务端说了 close」这件事只能从 resp.Close 观察。
	close bool
	err   error
}

// tryGet 不碰 *testing.T，可以从任意 goroutine 调用。
func (ts *testServer) tryGet(path string) httpResult {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(ts.addr + path)
	if err != nil {
		return httpResult{err: err}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return httpResult{
		code:   resp.StatusCode,
		body:   strings.TrimSpace(string(body)),
		header: resp.Header,
		close:  resp.Close,
		err:    err,
	}
}

func (ts *testServer) get(t *testing.T, path string) httpResult {
	t.Helper()
	res := ts.tryGet(path)
	if res.err != nil {
		t.Fatalf("GET %s 失败: %v", path, res.err)
	}
	return res
}

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// ── 用例 1：在途请求必须跑完 ──────────────────────────────────────────────

func TestInFlightRequestCompletesDuringDrain(t *testing.T) {
	started := make(chan struct{})
	mux := http.NewServeMux()
	handlerDone := make(chan time.Time, 1)
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		close(started)
		select {
		case <-time.After(300 * time.Millisecond):
			fmt.Fprint(w, "done")
		case <-r.Context().Done():
			fmt.Fprint(w, "interrupted")
		}
		handlerDone <- time.Now()
	})

	ts := start(t, Config{Handler: mux})

	type res struct {
		code int
		body string
		err  error
	}
	got := make(chan res, 1)
	go func() {
		r := ts.tryGet("/slow")
		got <- res{r.code, r.body, r.err}
	}()
	<-started

	begin := time.Now()
	err := ts.lc.Shutdown(context.Background())
	waited := time.Since(begin)
	if err != nil {
		// 带上 handler 的实际结束时间，才能在失败时区分「handler 太慢」
		// 和「Shutdown 没在等 handler」。
		select {
		case end := <-handlerDone:
			t.Fatalf("Shutdown 返回错误: %v（等了 %v；handler 在 Shutdown 开始后 %v 才返回）",
				err, waited, end.Sub(begin).Round(time.Millisecond))
		default:
			t.Fatalf("Shutdown 返回错误: %v（等了 %v；handler 至今没有返回）", err, waited)
		}
	}

	select {
	case r := <-got:
		if r.err != nil {
			t.Fatalf("在途请求出错: %v", r.err)
		}
		if r.code != http.StatusOK || r.body != "done" {
			t.Fatalf("在途请求结果 = %d %q，想要 200 done", r.code, r.body)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("在途请求一直没有返回")
	}
	if waited < 250*time.Millisecond {
		t.Fatalf("Shutdown 只等了 %v 就返回了，说明它没有等在途请求", waited)
	}
}

// ── 用例 2：阶段一翻 readiness，但仍在服务 ────────────────────────────────

func TestReadinessFlipsButStillServesDuringPreDrain(t *testing.T) {
	started := make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		close(started)
		select {
		case <-time.After(500 * time.Millisecond):
			fmt.Fprint(w, "done")
		case <-r.Context().Done():
			fmt.Fprint(w, "interrupted")
		}
	})
	mux.HandleFunc("/fast", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "ok") })

	ts := start(t, Config{Handler: mux, PreDrainDelay: 300 * time.Millisecond})

	go func() { _ = ts.tryGet("/slow") }()
	<-started

	shutdownDone := make(chan error, 1)
	go func() { shutdownDone <- ts.lc.Shutdown(context.Background()) }()

	// 阶段一窗口内：readiness 已经 503，但普通请求仍然 200。
	time.Sleep(80 * time.Millisecond)
	if res := ts.get(t, "/readyz"); res.code != http.StatusServiceUnavailable {
		t.Errorf("排空期间 /readyz = %d，想要 503", res.code)
	}
	if res := ts.get(t, "/healthz"); res.code != http.StatusOK {
		t.Errorf("排空期间 /healthz = %d，想要 200（存活探针不能看关机状态）", res.code)
	}
	if res := ts.get(t, "/fast"); res.code != http.StatusOK || res.body != "ok" {
		t.Errorf("阶段一应该继续服务，得到 %d %q", res.code, res.body)
	}
	if state := ts.lc.State(); state != StateDraining {
		t.Errorf("State() = %v，想要 draining", state)
	}

	if err := <-shutdownDone; err != nil {
		t.Fatalf("Shutdown 返回错误: %v", err)
	}
	if state := ts.lc.State(); state != StateStopped {
		t.Errorf("关机后 State() = %v，想要 stopped", state)
	}
}

// ── 用例 3：RejectNewRequests 开关 ────────────────────────────────────────

func TestRejectNewRequestsDuringDrain(t *testing.T) {
	started := make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		close(started)
		select {
		case <-time.After(400 * time.Millisecond):
		case <-r.Context().Done():
		}
	})
	mux.HandleFunc("/fast", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "ok") })

	ts := start(t, Config{
		Handler:           mux,
		PreDrainDelay:     300 * time.Millisecond,
		RejectNewRequests: true,
	})

	go func() { _ = ts.tryGet("/slow") }()
	<-started

	done := make(chan error, 1)
	go func() { done <- ts.lc.Shutdown(context.Background()) }()

	time.Sleep(80 * time.Millisecond)
	res := ts.get(t, "/fast")
	if res.code != http.StatusServiceUnavailable {
		t.Fatalf("/fast = %d，想要 503", res.code)
	}
	if !strings.Contains(res.body, "draining") {
		t.Errorf("响应体 = %q，想要包含 draining", res.body)
	}
	// 服务端确实要求关连接（Go 客户端把 Connection 头删了，只能看 resp.Close）。
	if !res.close {
		t.Errorf("resp.Close = false，想要服务端声明 Connection: close")
	}
	// 健康检查不受闸门影响，否则 Pod 会被直接判定为不健康。
	if res := ts.get(t, "/healthz"); res.code != http.StatusOK {
		t.Errorf("排空期间 /healthz = %d，想要 200", res.code)
	}
	if err := <-done; err != nil {
		t.Fatalf("Shutdown 返回错误: %v", err)
	}
}

// ── 用例 4：卡住的请求会被超时兜底强杀 ────────────────────────────────────

func TestShutdownTimeoutForcesClose(t *testing.T) {
	started := make(chan struct{})
	ctxDone := make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("/stuck", func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done() // 只有连接被强制关闭才会走到这里
		close(ctxDone)
	})

	ts := start(t, Config{Handler: mux, DrainTimeout: 200 * time.Millisecond})
	go func() { _ = ts.tryGet("/stuck") }()
	<-started

	begin := time.Now()
	err := ts.lc.Shutdown(context.Background())
	elapsed := time.Since(begin)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Shutdown 错误 = %v，想要 context.DeadlineExceeded", err)
	}
	if elapsed > time.Second {
		t.Fatalf("Shutdown 用了 %v，超时兜底没有生效", elapsed)
	}
	select {
	case <-ctxDone:
		// 强杀确实传导到了 handler 的 r.Context()
	case <-time.After(time.Second):
		t.Fatal("srv.Close() 之后 handler 的 context 没有被取消")
	}
}

// ── 用例 5：钩子逆序执行，且在 HTTP 排空之后 ──────────────────────────────

func TestHooksRunInReverseOrderAfterDrain(t *testing.T) {
	var mu sync.Mutex
	var order []string
	inFlight := make(chan struct{})
	handlerDone := false

	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(200 * time.Millisecond):
		case <-r.Context().Done():
		}
		mu.Lock()
		handlerDone = true
		mu.Unlock()
		close(inFlight)
		fmt.Fprint(w, "done")
	})

	ts := start(t, Config{Handler: mux})
	// 先注册数据库，再注册遥测 → 期望执行顺序反过来。
	ts.lc.OnShutdown("db", func(context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		order = append(order, "db")
		return nil
	})
	ts.lc.OnShutdown("telemetry", func(ctx context.Context) error {
		select {
		case <-inFlight:
		default:
			return errors.New("遥测钩子跑在在途请求结束之前")
		}
		mu.Lock()
		defer mu.Unlock()
		order = append(order, "telemetry")
		return nil
	})

	go func() { _ = ts.tryGet("/slow") }()
	time.Sleep(50 * time.Millisecond)

	if err := ts.lc.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown 返回错误: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if !handlerDone {
		t.Fatal("handler 没有跑完")
	}
	if want := []string{"telemetry", "db"}; fmt.Sprint(order) != fmt.Sprint(want) {
		t.Fatalf("钩子顺序 = %v，想要 %v（注册顺序的逆序）", order, want)
	}
}

func TestHookErrorIsJoined(t *testing.T) {
	sentinel := errors.New("db 已经关了")
	ts := start(t, Config{})
	ts.lc.OnShutdown("db", func(context.Context) error { return sentinel })

	err := ts.lc.Shutdown(context.Background())
	if !errors.Is(err, sentinel) {
		t.Fatalf("Shutdown 错误 = %v，想要包含 %v", err, sentinel)
	}
}

// ── 用例 6：Hijack 出去的长连接要自己关 ───────────────────────────────────

func TestHijackedConnectionIsClosedOnShutdown(t *testing.T) {
	// 用 channel 把 handler 的结果传回来，顺便建立 happens-before 关系。
	hijackErr := make(chan error, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			hijackErr <- errors.New("trackedWriter 没有透出 Hijack")
			return
		}
		conn, rw, err := hj.Hijack()
		if err != nil {
			hijackErr <- err
			return
		}
		// 升级成「WebSocket」：写完握手就不管了，net/http 也不再管这条连接。
		fmt.Fprint(rw, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n\r\n")
		_ = rw.Flush()
		_ = conn // 连接由 Lifecycle 在阶段三统一关闭
		hijackErr <- nil
	})

	ts := start(t, Config{Handler: mux, DrainTimeout: time.Second})

	conn, err := net.Dial("tcp", ts.ln.Addr().String())
	if err != nil {
		t.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()
	fmt.Fprint(conn, "GET /ws HTTP/1.1\r\nHost: test\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\n")

	// 读完整段响应头（直到空行）。只读一行的话，客户端缓冲区里剩下的
	// "\r\n" 会被后面的断言误当成「连接还活着」。
	br := bufio.NewReader(conn)
	status, err := br.ReadString('\n')
	if err != nil || !strings.Contains(status, "101") {
		t.Fatalf("握手响应 = %q err=%v，想要 101", status, err)
	}
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			t.Fatalf("响应头读不完: %v", err)
		}
		if line == "\r\n" {
			break
		}
	}
	if err := <-hijackErr; err != nil {
		t.Fatalf("handler 里的 Hijack 出错: %v", err)
	}

	if err := ts.lc.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown 返回错误: %v", err)
	}

	// Shutdown 返回之后，这条被 Hijack 的连接必须已经被关掉。
	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	if b, err := br.ReadByte(); err == nil {
		t.Fatalf("关机后 Hijack 连接仍然可读（读到 %q），说明它没有被关闭", b)
	}
}

// ── 用例 7：长连接靠 Draining() 主动收尾 ──────────────────────────────────

func TestStreamingHandlerStopsOnDraining(t *testing.T) {
	events := 0
	var mu sync.Mutex

	var lc *Lifecycle
	mux := http.NewServeMux()
	mux.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		rc := http.NewResponseController(w)
		for i := 1; ; i++ {
			select {
			case <-r.Context().Done():
				return
			case <-lc.Draining():
				fmt.Fprint(w, "event: closing\n\n")
				_ = rc.Flush()
				return
			case <-time.After(20 * time.Millisecond):
				mu.Lock()
				events++
				mu.Unlock()
				fmt.Fprint(w, "data: ping\n\n")
				if err := rc.Flush(); err != nil {
					return
				}
			}
		}
	})
	lc = New(Config{
		Handler:       mux,
		Logger:        quietLogger(),
		PreDrainDelay: 100 * time.Millisecond,
		DrainTimeout:  time.Second,
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("监听失败: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- lc.Serve(ctx, ln) }()
	defer func() {
		cancel()
		select {
		case <-errCh:
		case <-time.After(5 * time.Second):
			t.Error("Serve 没有返回")
		}
	}()

	body := make(chan string, 1)
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String() + "/sse")
		if err != nil {
			body <- "error: " + err.Error()
			return
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		body <- string(b)
	}()

	time.Sleep(60 * time.Millisecond)

	// 如果 SSE handler 不监听 Draining()，这里会一直等到超时并返回错误。
	if err := lc.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown 返回错误: %v", err)
	}

	select {
	case got := <-body:
		if !strings.Contains(got, "event: closing") {
			t.Fatalf("SSE 响应 = %q，想要包含 closing 事件", got)
		}
	case <-time.After(time.Second):
		t.Fatal("SSE 连接没有结束")
	}

	mu.Lock()
	defer mu.Unlock()
	if events == 0 {
		t.Fatal("关得太早，一条推送都没发出去")
	}
}

// ── 用例 8：关机之后新连接被拒；Shutdown 幂等 ─────────────────────────────

func TestNewConnectionRefusedAndShutdownIdempotent(t *testing.T) {
	ts := start(t, Config{})

	if err := ts.lc.Shutdown(context.Background()); err != nil {
		t.Fatalf("第一次 Shutdown: %v", err)
	}
	if err := ts.lc.Shutdown(context.Background()); err != nil {
		t.Fatalf("第二次 Shutdown 应该复用第一次的结果，得到: %v", err)
	}

	if conn, err := net.DialTimeout("tcp", ts.ln.Addr().String(), 300*time.Millisecond); err == nil {
		conn.Close()
		t.Fatal("关机后仍能建立新连接")
	}
	if ts.lc.State() != StateStopped {
		t.Errorf("State() = %v，想要 stopped", ts.lc.State())
	}
}
