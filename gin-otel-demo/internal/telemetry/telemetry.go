// Package telemetry 负责装配 OpenTelemetry SDK。
//
// 它是整个应用里唯一直接依赖 SDK 的包：业务代码只 import otel 的 API
// （otel.Tracer / otel.Meter），这样"换后端、改采样"都不需要动业务逻辑，
// 这也是 OpenTelemetry 把 API 和 SDK 分开的意义。
package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

// Config 是初始化遥测需要的全部参数，由 main.go 从环境变量读出后传进来。
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	Endpoint       string        // Collector 的 OTLP/gRPC 地址，形如 127.0.0.1:4317
	Insecure       bool          // 本地开发用明文传输，生产应走 TLS
	Console        bool          // true 时改为打到标准输出，零依赖调试用
	SampleRatio    float64       // 根跨度的采样比例，0~1
	MetricInterval time.Duration // 指标导出间隔
}

// Setup 初始化全局 TracerProvider / MeterProvider / 传播器。
//
// 返回的 shutdown 必须在进程退出前调用：span 和 metric 都是批量、异步发送的，
// 不 Shutdown 就等于丢掉最后一批数据。
func Setup(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	// 社区约定的"总开关"，线上出问题时可以一键退化成 noop。
	if strings.EqualFold(os.Getenv("OTEL_SDK_DISABLED"), "true") {
		slog.Warn("OTEL_SDK_DISABLED=true，遥测已关闭")
		return func(context.Context) error { return nil }, nil
	}

	// ① Resource：描述"数据是谁产生的"。service.name 最关键，
	//    缺了它后端里所有数据都会显示成 unknown_service。
	res, err := resource.New(ctx,
		resource.WithFromEnv(),   // OTEL_RESOURCE_ATTRIBUTES 里的额外属性
		resource.WithProcess(),   // 进程信息（pid、命令行、运行时版本）
		resource.WithOS(),        // 操作系统
		resource.WithHost(),      // 主机名
		resource.WithContainer(), // 容器信息（不在容器里会自动跳过）
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(cfg.ServiceVersion),
			semconv.DeploymentEnvironmentNameKey.String(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("构造 resource: %w", err)
	}

	// ② Exporter：数据往哪发。默认 OTLP/gRPC，本地调试可以换成 stdout。
	var (
		spanExporter   sdktrace.SpanExporter
		metricExporter metric.Exporter
	)
	if cfg.Console {
		if spanExporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint()); err != nil {
			return nil, fmt.Errorf("构造 stdout trace exporter: %w", err)
		}
		if metricExporter, err = stdoutmetric.New(stdoutmetric.WithPrettyPrint()); err != nil {
			return nil, fmt.Errorf("构造 stdout metric exporter: %w", err)
		}
	} else {
		endpoint := strings.TrimPrefix(strings.TrimPrefix(cfg.Endpoint, "http://"), "https://")
		if spanExporter, err = otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(endpoint),
			otlptracegrpc.WithInsecure(),
		); err != nil {
			return nil, fmt.Errorf("构造 OTLP trace exporter: %w", err)
		}
		if metricExporter, err = otlpmetricgrpc.New(ctx,
			otlpmetricgrpc.WithEndpoint(endpoint),
			otlpmetricgrpc.WithInsecure(),
		); err != nil {
			return nil, fmt.Errorf("构造 OTLP metric exporter: %w", err)
		}
	}

	// ③ Sampler：ParentBased 保证"父 span 采了，子 span 一定采"，
	//    否则一条链路会缺一半，排查问题时非常难用。
	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRatio))

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(spanExporter), // 攒批再发，减少网络往返
	)

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			metric.WithInterval(cfg.MetricInterval))),
	)

	// ④ 注册为全局，之后 otel.Tracer(...) / otel.Meter(...) 拿到的就是它们。
	otel.SetTracerProvider(tracerProvider)
	otel.SetMeterProvider(meterProvider)

	// ⑤ 传播器：traceparent 决定链路能否跨服务串起来，Baggage 用来透传业务上下文。
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// ⑥ 默认错误处理器会静默丢弃 SDK 内部错误（比如导出失败），
	//    排查"后端为什么没数据"时第一件事就是把它打开。
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		slog.Error("opentelemetry sdk error", "error", err)
	}))

	slog.Info("遥测初始化完成",
		"service", cfg.ServiceName,
		"env", cfg.Environment,
		"exporter", map[bool]string{true: "console", false: "otlp/grpc"}[cfg.Console],
		"endpoint", cfg.Endpoint,
		"sample_ratio", cfg.SampleRatio,
	)

	// 两个 provider 都要关，错误用 errors.Join 合并上报。
	return func(ctx context.Context) error {
		return errors.Join(tracerProvider.Shutdown(ctx), meterProvider.Shutdown(ctx))
	}, nil
}
