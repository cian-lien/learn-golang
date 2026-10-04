// Package graceful 把 Go Web 服务的「平滑关机」拆成四个阶段来实现。
//
//	StateReady     正常服务
//	   │  收到 SIGTERM / SIGINT
//	StateDraining  阶段一：readiness 立刻转 503，给负载均衡摘流量留出时间
//	   │           阶段二：srv.Shutdown —— 关监听、关空闲连接、等在途请求结束
//	   │           阶段三：关闭被 Hijack 走的长连接（WebSocket / SSE）
//	   │           阶段四：按注册的逆序执行收尾钩子（flush 遥测、关数据库……）
//	StateStopped   进程可以安全退出
//
// 三个容易被忽略的事实，也是这个包存在的原因：
//
//  1. http.Server.Shutdown 不会取消在途请求的 context。它只是「不再接新连接」+
//     「等现有连接处理完」。想让 SSE / 长轮询主动收尾，得靠 Lifecycle.Draining()。
//  2. Shutdown 不管被 Hijack 的连接（WebSocket 升级后就离开了 net/http 的管辖）。
//     这些连接要自己登记、自己关，否则 Shutdown 会直接返回，进程带着连接退出。
//  3. Shutdown 自己也需要超时，否则一个卡住的请求就能把发布流程挂到天荒地老。
package graceful

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

// State 描述服务在关机流程中的位置。
type State int32

const (
	StateStarting State = iota // 还没开始监听
	StateReady                 // 正常服务
	StateDraining              // 已收到停止信号，正在摘流量 / 排空
	StateStopped               // 排空与收尾都已完成
)

func (s State) String() string {
	switch s {
	case StateStarting:
		return "starting"
	case StateReady:
		return "ready"
	case StateDraining:
		return "draining"
	case StateStopped:
		return "stopped"
	default:
		return fmt.Sprintf("state(%d)", int32(s))
	}
}

// Config 是 Lifecycle 的配置，零值可用。
type Config struct {
	// Addr 只在 ListenAndServe 时用到；自己 net.Listen 后再 Serve 可以忽略它。
	Addr    string
	Handler http.Handler
	Logger  *slog.Logger

	// HTTP 超时。注意 WriteTimeout 会掐断超过它的 SSE / 大文件下载，
	// 有流式接口时应把它设为 0，改由 handler 自己控制单次写入的 deadline。
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration

	// PreDrainDelay 是阶段一的时长：readiness 转 503 之后，等负载均衡真的
	// 把本实例摘掉。K8s 上通常设 2~5s，等于「摘流量延迟」+ 一点余量。
	PreDrainDelay time.Duration

	// DrainTimeout 是阶段一 + 二 + 三的总预算，也就是「最后一次请求允许跑多久」。
	DrainTimeout time.Duration

	// HookTimeout 是阶段四所有收尾钩子共享的预算。
	// 它和 DrainTimeout 分开，是为了避免一个慢请求吃光预算后，数据库连接
	// 和遥测 exporter 被直接跳过。
	HookTimeout time.Duration

	// RejectNewRequests 为 true 时，排空期间非健康检查请求直接 503 并关闭连接。
	// 默认 false：靠 readiness 让 LB 停止导流，已经连上来的客户端仍能拿到正常结果。
	RejectNewRequests bool

	// 健康检查路径，默认 /healthz（存活）、/readyz（就绪）。
	HealthPath string
	ReadyPath  string
}

func (c Config) withDefaults() Config {
	if c.Handler == nil {
		c.Handler = http.NewServeMux()
	}
	if c.Logger == nil {
		c.Logger = slog.Default()
	}
	if c.ReadHeaderTimeout == 0 {
		c.ReadHeaderTimeout = 5 * time.Second
	}
	if c.ReadTimeout == 0 {
		c.ReadTimeout = 15 * time.Second
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = 15 * time.Second
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = 60 * time.Second
	}
	if c.DrainTimeout == 0 {
		c.DrainTimeout = 15 * time.Second
	}
	if c.HookTimeout == 0 {
		c.HookTimeout = 5 * time.Second
	}
	if c.HealthPath == "" {
		c.HealthPath = "/healthz"
	}
	if c.ReadyPath == "" {
		c.ReadyPath = "/readyz"
	}
	return c
}

type hook struct {
	name string
	fn   func(context.Context) error
}

// Lifecycle 包住一个 http.Server，负责状态流转、探针、在途统计和四阶段排空。
// 所有导出方法都可以并发调用。
type Lifecycle struct {
	cfg Config
	log *slog.Logger
	srv *http.Server

	state    atomic.Int32
	inflight atomic.Int64
	requests atomic.Int64
	rejected atomic.Int64

	mu       sync.Mutex
	hooks    []hook
	hijacked map[net.Conn]struct{}

	draining     chan struct{}
	drainingOnce sync.Once

	shutdownOnce sync.Once
	shutdownErr  error
}

// New 组装 Lifecycle。此时还没有开始监听。
func New(cfg Config) *Lifecycle {
	cfg = cfg.withDefaults()
	l := &Lifecycle{
		cfg:      cfg,
		log:      cfg.Logger,
		hijacked: make(map[net.Conn]struct{}),
		draining: make(chan struct{}),
	}
	l.srv = &http.Server{
		Addr:              cfg.Addr,
		Handler:           l.wrap(cfg.Handler),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		ErrorLog:          slog.NewLogLogger(cfg.Logger.Handler(), slog.LevelWarn),
	}
	return l
}

// State 返回当前状态。
func (l *Lifecycle) State() State { return State(l.state.Load()) }

// InFlight 返回「已经进入 handler 但还没返回」的请求数。
// 它用于观测和断言；排空的最终依据仍然是 http.Server.Shutdown 自己维护的连接状态。
func (l *Lifecycle) InFlight() int64 { return l.inflight.Load() }

// Draining 返回一个在进入排空阶段时关闭的 channel。
//
// 长连接 handler（SSE、长轮询、消息推送）必须监听它：
//
//	select {
//	case <-r.Context().Done():   // 客户端断开
//	case <-lc.Draining():        // 进程要退出了，主动收尾
//	}
//
// 不监听也可以，代价是 Shutdown 只能干等到 DrainTimeout 再强杀。
func (l *Lifecycle) Draining() <-chan struct{} { return l.draining }

// OnShutdown 注册收尾钩子。执行顺序是注册顺序的逆序（后注册的先关），
// 也就是初始化的反方向：先起遥测后起数据库，退出时就先 flush 遥测再关数据库。
// 钩子在 HTTP 排空之后执行，共享 Config.HookTimeout 预算；错误会合并返回。
func (l *Lifecycle) OnShutdown(name string, fn func(context.Context) error) {
	if fn == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.hooks = append(l.hooks, hook{name: name, fn: fn})
}

// ListenAndServe 在 cfg.Addr 上服务，直到 ctx 取消。
func (l *Lifecycle) ListenAndServe(ctx context.Context) error {
	ln, err := net.Listen("tcp", l.cfg.Addr)
	if err != nil {
		return fmt.Errorf("监听 %s: %w", l.cfg.Addr, err)
	}
	return l.Serve(ctx, ln)
}

// Serve 在 ln 上服务，直到 ctx 取消；ctx 取消后同步完成排空再返回。
// 返回 nil 表示「干净地退了」，非 nil 表示启动失败或排空不干净。
func (l *Lifecycle) Serve(ctx context.Context, ln net.Listener) error {
	l.state.Store(int32(StateReady))
	l.log.Info("HTTP 服务就绪", "addr", ln.Addr().String())

	serveErr := make(chan error, 1)
	go func() {
		err := l.srv.Serve(ln)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil // Shutdown / Close 之后的正常返回
		}
		serveErr <- err
	}()

	select {
	case err := <-serveErr:
		// 还没收到停止信号，Serve 自己先挂了（端口冲突、Accept 错误……）。
		l.state.Store(int32(StateStopped))
		if err == nil {
			return nil
		}
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
		l.log.Info("收到退出信号，开始平滑关机", "in_flight", l.InFlight())
	}

	if err := l.Shutdown(ctx); err != nil {
		return err
	}
	// Shutdown 返回时所有连接都已关闭，这里只是收走 Serve 的返回值。
	return <-serveErr
}

// Shutdown 执行完整的四阶段排空。重复调用返回第一次的结果（幂等）。
func (l *Lifecycle) Shutdown(parent context.Context) error {
	l.shutdownOnce.Do(func() {
		l.shutdownErr = l.drain(parent)
	})
	return l.shutdownErr
}

func (l *Lifecycle) drain(parent context.Context) error {
	// parent 此刻多半已经被信号取消（如果直接拿它做超时，会立刻超时）。
	// context.WithoutCancel 保留 ctx 上的值（日志字段、trace），只切断取消信号。
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), l.cfg.DrainTimeout)
	defer cancel()

	// ── 阶段一：翻 readiness，等 LB 摘流量 ──────────────────────────────
	l.state.Store(int32(StateDraining))
	l.drainingOnce.Do(func() { close(l.draining) })
	l.log.Info("阶段一 readiness → 503，等待负载均衡摘流量",
		"pre_drain_delay", l.cfg.PreDrainDelay, "in_flight", l.InFlight())
	if l.cfg.PreDrainDelay > 0 {
		if err := sleepCtx(ctx, l.cfg.PreDrainDelay); err != nil {
			l.log.Warn("等待摘流量被中断，直接进入排空", "error", err)
		}
	}

	// ── 阶段二：关监听 + 等在途请求 ────────────────────────────────────
	// Shutdown 会：关闭所有 listener（新连接被拒）、关闭空闲连接、
	// 然后轮询等待活跃连接变成空闲。它不会断开活跃连接，也不会取消它们的 context。
	l.log.Info("阶段二 关闭监听并等待在途请求", "in_flight", l.InFlight())
	shutdownErr := l.srv.Shutdown(ctx)
	if shutdownErr != nil {
		l.log.Warn("在途请求未在超时内结束，强制关闭连接",
			"error", shutdownErr, "in_flight", l.InFlight())
		_ = l.srv.Close()
	}

	// ── 阶段三：关掉 Hijack 走的长连接 ─────────────────────────────────
	n := l.closeHijacked()
	if n > 0 {
		l.log.Info("阶段三 关闭被 Hijack 的长连接", "conns", n)
	}

	// ── 阶段四：收尾钩子（逆序） ───────────────────────────────────────
	hookCtx, cancelHooks := context.WithTimeout(context.WithoutCancel(parent), l.cfg.HookTimeout)
	defer cancelHooks()
	hookErr := l.runHooks(hookCtx)

	l.state.Store(int32(StateStopped))
	l.log.Info("已停止",
		"requests", l.requests.Load(), "rejected", l.rejected.Load(),
		"in_flight", l.InFlight(), "shutdown_error", errString(shutdownErr))
	return errors.Join(shutdownErr, hookErr)
}

func (l *Lifecycle) runHooks(ctx context.Context) error {
	l.mu.Lock()
	hooks := slices.Clone(l.hooks)
	l.mu.Unlock()

	var errs []error
	for i := len(hooks) - 1; i >= 0; i-- {
		h := hooks[i]
		start := time.Now()
		err := h.fn(ctx)
		l.log.Info("收尾钩子", "hook", h.name, "elapsed", time.Since(start).Round(time.Millisecond), "error", errString(err))
		if err != nil {
			errs = append(errs, fmt.Errorf("钩子 %s: %w", h.name, err))
		}
	}
	return errors.Join(errs...)
}

// wrap 串起探针、排空闸门、在途计数。
func (l *Lifecycle) wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case l.cfg.HealthPath:
			// 存活探针只回答「进程还活着」，绝不能看关机状态：
			// 否则 K8s 会在排空途中把 Pod 杀掉，前面的等待白做。
			writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "state": l.State().String()})
			return
		case l.cfg.ReadyPath:
			// 就绪探针回答「能不能接新流量」。
			if l.State() == StateReady {
				writeJSON(w, http.StatusOK, map[string]any{"status": "ready", "state": l.State().String()})
				return
			}
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "draining", "state": l.State().String()})
			return
		}

		if l.cfg.RejectNewRequests && l.State() >= StateDraining {
			l.rejected.Add(1)
			// Connection: close 让客户端别再往这条连接上排队。
			w.Header().Set("Connection", "close")
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "draining"})
			return
		}

		l.requests.Add(1)
		l.inflight.Add(1)
		defer l.inflight.Add(-1)
		next.ServeHTTP(&trackedWriter{ResponseWriter: w, l: l}, r)
	})
}

func (l *Lifecycle) trackConn(c net.Conn) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.hijacked[c] = struct{}{}
}

func (l *Lifecycle) closeHijacked() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	n := 0
	for c := range l.hijacked {
		if err := c.Close(); err == nil {
			n++
		}
		delete(l.hijacked, c)
	}
	return n
}

// trackedWriter 让 handler 里的 Hijack 能被记录，同时把 Flush /
// SetWriteDeadline 等能力通过 Unwrap 透传给 http.ResponseController。
type trackedWriter struct {
	http.ResponseWriter
	l *Lifecycle
}

func (w *trackedWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (w *trackedWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("底层 ResponseWriter 不支持 Hijack: %w", http.ErrNotSupported)
	}
	conn, rw, err := h.Hijack()
	if err != nil {
		return nil, nil, err
	}
	w.l.trackConn(conn)
	return conn, rw, nil
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
