package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
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
		// 已结束的 Span 进入队列，由后台批量导出。
		sdktrace.WithBatcher(exporter),
		// Resource 描述产生这些 Span 的应用。
		sdktrace.WithResource(resource.NewSchemaless(
			attribute.String("service.name", "otel-learning"),
		)),
	)
	defer func() {
		// 程序退出前，等待队列中的 Span 导出并关闭 Provider。
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
	if err := prepareMessage(parentCtx, tracer); err != nil {
		// 这个练习中，消息处理失败意味着整个问候操作失败。
		parentSpan.RecordError(err)
		parentSpan.SetStatus(codes.Error, "greeting failed")
		log.Print(err)
	}
}

func prepareMessage(ctx context.Context, tracer trace.Tracer) error {
	// 这里的 ctx 来自 parentCtx，所以新 Span 是 say-hello 的子 Span。
	_, span := tracer.Start(ctx, "prepare-message")
	defer span.End()

	message := "hello, OTel"
	// 给当前 Span 添加业务信息。
	span.SetAttributes(attribute.String("app.message", message))

	fmt.Println(message)
	// 在当前 Span 中记录一个带时间戳的事件。
	span.AddEvent("message.printed")
	time.Sleep(100 * time.Millisecond)

	// 模拟消息打印后的处理失败，学习如何记录错误。
	err := errors.New("message processing failed")
	span.RecordError(err)
	span.SetStatus(codes.Error, "message processing failed")
	return err
}
