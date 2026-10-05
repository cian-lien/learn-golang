package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"

	"observability-learning/internal/demo"
	"observability-learning/internal/telemetry"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:18090", "HTTP 监听地址")
	exporter := flag.String("exporter", "console", "console 或 otlp")
	endpoint := flag.String("endpoint", "127.0.0.1:14317", "Collector OTLP/gRPC 地址")
	flag.Parse()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(*addr, *exporter, *endpoint, logger); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func run(addr, exporter, endpoint string, logger *slog.Logger) (runErr error) {
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) { logger.Error("OTel export failed", "error", err) }))
	providers, err := telemetry.Setup(context.Background(), exporter, endpoint)
	if err != nil {
		return err
	}
	// 所有退出路径都 flush 遥测；先关闭 HTTP，再执行这个 defer。
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		runErr = errors.Join(runErr, providers.Shutdown(ctx))
	}()
	handler, err := demo.NewHandler(providers.Traces.Tracer("learning/http"), providers.Metrics.Meter("learning/http"), logger)
	if err != nil {
		return err
	}
	server := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	serverErr := make(chan error, 1)
	go func() { serverErr <- server.ListenAndServe() }()
	logger.Info("server starting", "addr", addr, "exporter", exporter)
	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		_ = server.Close()
		return err
	}
	logger.Info("HTTP stopped; flushing telemetry")
	return nil
}
