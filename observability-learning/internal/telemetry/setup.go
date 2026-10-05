// Package telemetry 只在启动时装配 SDK：资源、导出器、采集器和 Provider。
package telemetry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Providers struct {
	Traces  *sdktrace.TracerProvider
	Metrics *sdkmetric.MeterProvider
}

// mode=console 时输出到终端；mode=otlp 时发送到本地 Collector。
func Setup(ctx context.Context, mode, endpoint string) (*Providers, error) {
	var traces sdktrace.SpanExporter
	var metrics sdkmetric.Exporter
	var err error
	switch mode {
	case "console":
		traces, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err == nil {
			metrics, err = stdoutmetric.New(stdoutmetric.WithPrettyPrint())
		}
	case "otlp":
		traces, err = otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(endpoint), otlptracegrpc.WithInsecure())
		if err == nil {
			metrics, err = otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpoint(endpoint), otlpmetricgrpc.WithInsecure())
		}
	default:
		return nil, fmt.Errorf("unknown exporter %q: use console or otlp", mode)
	}
	if err != nil {
		if traces != nil {
			_ = traces.Shutdown(ctx)
		}
		return nil, fmt.Errorf("create exporters: %w", err)
	}
	res := resource.NewSchemaless(attribute.String("service.name", "observability-learning"))
	return &Providers{
		Traces: sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample())),
			sdktrace.WithBatcher(traces, sdktrace.WithBatchTimeout(time.Second)),
		),
		Metrics: sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metrics, sdkmetric.WithInterval(3*time.Second))),
			sdkmetric.WithView(sdkmetric.NewView(
				sdkmetric.Instrument{Name: "demo.request.duration"},
				sdkmetric.Stream{Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
					Boundaries: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2},
				}},
			)),
		),
	}, nil
}

func (p *Providers) Shutdown(ctx context.Context) error {
	return errors.Join(p.Traces.Shutdown(ctx), p.Metrics.Shutdown(ctx))
}
