package demo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// NewHandler 演示手动埋点：tracer / meter 是 API 接口，业务层不依赖 SDK。
func NewHandler(tracer trace.Tracer, meter metric.Meter, logger *slog.Logger) (http.Handler, error) {
	requests, err := meter.Int64Counter("demo.requests", metric.WithDescription("业务 HTTP 请求次数"), metric.WithUnit("{request}"))
	if err != nil {
		return nil, fmt.Errorf("create counter: %w", err)
	}
	duration, err := meter.Float64Histogram("demo.request.duration", metric.WithDescription("业务 HTTP 请求耗时"), metric.WithUnit("s"))
	if err != nil {
		return nil, fmt.Errorf("create histogram: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok") // 健康检查不进入业务指标和调用链。
	})
	for _, route := range []string{"/hello", "/slow", "/error"} {
		mux.HandleFunc("GET "+route, func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			// 从上游的 traceparent 请求头提取父节点；没有则创建新 trace。
			ctx := propagation.TraceContext{}.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
			ctx, span := tracer.Start(ctx, "GET "+route,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(attribute.String("http.request.method", r.Method), attribute.String("http.route", route)))
			defer span.End()
			w.Header().Set("X-Trace-ID", span.SpanContext().TraceID().String())

			body, status, err := work(ctx, tracer, route)
			span.SetAttributes(attribute.Int("http.response.status_code", status))
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(status)
			fmt.Fprintln(w, body)

			// 固定路由 + 状态码是有限的组合；不要加入 user_id 或 trace_id。
			labels := metric.WithAttributes(attribute.String("http.route", route), attribute.Int("http.response.status_code", status))
			requests.Add(ctx, 1, labels)
			duration.Record(ctx, time.Since(start).Seconds(), labels)
			logger.InfoContext(ctx, "request completed", "route", route, "status", status,
				"duration_ms", float64(time.Since(start).Microseconds())/1000,
				"trace_id", span.SpanContext().TraceID().String())
		})
	}
	return mux, nil
}

func work(ctx context.Context, tracer trace.Tracer, route string) (string, int, error) {
	ctx, span := tracer.Start(ctx, "demo.work") // 使用上层 ctx，成为 HTTP span 的子节点。
	defer span.End()
	if route == "/slow" {
		timer := time.NewTimer(250 * time.Millisecond)
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-ctx.Done():
			span.RecordError(ctx.Err())
			span.SetStatus(codes.Error, ctx.Err().Error())
			return "request canceled", http.StatusRequestTimeout, ctx.Err()
		}
	}
	if route == "/error" {
		err := errors.New("simulated failure")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "simulated failure", http.StatusInternalServerError, err
	}
	return "hello, Go", http.StatusOK, nil
}
