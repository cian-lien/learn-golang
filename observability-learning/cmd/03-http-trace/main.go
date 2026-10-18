package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// 沿用上一节的导出器、Provider 和 Resource 配置。
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return err
	}
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewSchemaless(
			attribute.String("service.name", "otel-http-learning"),
		)),
	)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := provider.Shutdown(ctx); err != nil {
			log.Print(err)
		}
	}()
	tracer := provider.Tracer("learning-http")

	// 本节重点：每次请求进入处理函数时，创建一个服务端 Span。
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(
			r.Context(),
			"GET /hello",
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		// ctx 携带 HTTP Span，让业务函数创建它的子 Span。
		message := prepareMessage(ctx, tracer)
		fmt.Fprintln(w, message)
	})

	server := &http.Server{
		Addr:              "127.0.0.1:18091",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Ctrl+C 时先停止 HTTP 服务、等待请求结束，再关闭 Provider。
	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.ListenAndServe()
	}()
	log.Print("访问 http://127.0.0.1:18091/hello；按 Ctrl+C 退出")

	select {
	case err := <-serveErr:
		return err
	case <-stopCtx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(ctx)
	}
}

func prepareMessage(ctx context.Context, tracer trace.Tracer) string {
	_, span := tracer.Start(ctx, "prepare-message")
	defer span.End()

	// 模拟准备消息的耗时，方便观察父子 Span 的时间范围。
	time.Sleep(50 * time.Millisecond)
	return "hello, OTel"
}
