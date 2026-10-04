// Command gin-otel-demo 是一个接入 OpenTelemetry 的 Gin Web 服务示例。
//
// 各层职责：
//
//	main.go             配置、路由组装、优雅退出
//	internal/telemetry  OTel SDK 装配（整个项目唯一直接依赖 SDK 的地方）
//	internal/logging    slog + trace_id 关联
//	internal/book       handler / service / store 三层各自的埋点示例
//
// 最快跑起来：
//
//	OTEL_EXPORTER=console go run .        # 零依赖：span/metric 直接打标准输出
//	docker compose up -d && go run .      # 完整链路：Collector + Jaeger(:16686)
//
// 然后：
//
//	curl localhost:8080/api/v1/books/1
//	curl localhost:8080/api/v1/books/error    # 故意失败，span 会标红
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"gin-otel-demo/internal/book"
	"gin-otel-demo/internal/logging"
	"gin-otel-demo/internal/telemetry"
)

const version = "0.1.0"

func main() {
	addr := flag.String("addr", envOr("APP_ADDR", ":8089"), "HTTP 监听地址")
	flag.Parse()

	logger := logging.New(slog.LevelInfo)
	slog.SetDefault(logger)

	if err := run(*addr, logger); err != nil {
		logger.Error("服务异常退出", "error", err)
		os.Exit(1)
	}
}

func run(addr string, logger *slog.Logger) error {
	serviceName := envOr("OTEL_SERVICE_NAME", "gin-otel-demo")

	// Ctrl-C / SIGTERM 时取消 ctx，触发优雅退出。
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 第一步永远是初始化遥测，之后创建的 tracer / meter 才会绑定到真正的 provider。
	shutdownTelemetry, err := telemetry.Setup(ctx, telemetry.Config{
		ServiceName:    serviceName,
		ServiceVersion: version,
		Environment:    envOr("DEPLOY_ENV", "development"),
		Endpoint:       envOr("OTEL_EXPORTER_OTLP_ENDPOINT", "127.0.0.1:4317"),
		Console:        strings.EqualFold(envOr("OTEL_EXPORTER", "otlp"), "console"),
		SampleRatio:    envFloat("OTEL_SAMPLE_RATIO", 1.0),
		MetricInterval: envDuration("OTEL_METRIC_INTERVAL_MS", 10*time.Second),
	})
	if err != nil {
		return fmt.Errorf("初始化遥测: %w", err)
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           newRouter(serviceName, logger),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("HTTP 服务启动", "addr", addr, "service", serviceName)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		logger.Info("收到退出信号，开始优雅退出")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 先停流量，再关遥测——否则最后一批正在处理的请求会丢 span。
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("关闭 HTTP 服务出错", "error", err)
	}
	if err := shutdownTelemetry(shutdownCtx); err != nil {
		logger.Error("关闭遥测出错", "error", err)
	}
	logger.Info("已退出")
	return nil
}

// newRouter 组装 Gin 引擎。埋点的"第一层"就发生在这里。
func newRouter(serviceName string, logger *slog.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// ① 崩溃恢复放在最外层，保证 panic 时下面的 defer 都能正常收尾。
	r.Use(gin.Recovery())

	// ② otelgin：一行接入。
	//    它自动完成：提取 traceparent -> 创建 server span -> 记录 HTTP 语义属性
	//    （http.request.method / http.route / http.response.status_code ...）->
	//    在请求结束时结束 span 并按状态码设置 span 状态。
	//    注意它用的是 c.FullPath()（如 /api/v1/books/:id）命名，天然低基数。
	r.Use(otelgin.Middleware(serviceName,
		otelgin.WithFilter(func(req *http.Request) bool {
			return req.URL.Path != "/healthz" // 健康检查太吵，不采集
		}),
	))

	// ③ 访问日志：用 InfoContext，日志里就会带上 trace_id。
	r.Use(requestLogger(logger))

	book.RegisterRoutes(r, book.NewService(book.NewStore(), logger))
	return r
}

// requestLogger 是一个普通 Gin 中间件，演示"日志和链路如何关联"。
func requestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		logger.LogAttrs(c.Request.Context(), slog.LevelInfo, "http 访问",
			slog.String("http.request.method", c.Request.Method),
			slog.String("url.path", c.Request.URL.Path),
			slog.Int("http.response.status_code", c.Writer.Status()),
			slog.Float64("duration_ms", float64(time.Since(start).Microseconds())/1000),
		)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envFloat(key string, def float64) float64 {
	v, err := strconv.ParseFloat(os.Getenv(key), 64)
	if err != nil || v < 0 || v > 1 {
		return def
	}
	return v
}

func envDuration(key string, def time.Duration) time.Duration {
	ms, err := strconv.Atoi(os.Getenv(key))
	if err != nil || ms <= 0 {
		return def
	}
	return time.Duration(ms) * time.Millisecond
}
