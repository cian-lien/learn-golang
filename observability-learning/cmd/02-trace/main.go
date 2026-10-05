package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	ctx := context.Background()

	// 1. 创建导出器：把 Span 输出到终端。
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatal(err)
	}

	// 2. 创建 Provider：负责提供具有记录能力的 Tracer。
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	defer func() {
		if err := provider.Shutdown(ctx); err != nil {
			log.Print(err)
		}
	}()

	// 3. 获取 Tracer：用它创建 Span。
	tracer := provider.Tracer("learning")

	// 4. 创建父 Span，返回的 parentCtx 携带这个 Span。
	parentCtx, parentSpan := tracer.Start(ctx, "say-hello")
	defer parentSpan.End()

	// 5. 把携带父 Span 的 Context 传给下一个函数。
	prepareMessage(parentCtx, tracer)
}

func prepareMessage(ctx context.Context, tracer trace.Tracer) {
	// 这里的 ctx 来自 parentCtx，所以新 Span 是 say-hello 的子 Span。
	_, span := tracer.Start(ctx, "prepare-message")
	defer span.End()

	fmt.Println("hello, OTel")
	time.Sleep(100 * time.Millisecond)
}
