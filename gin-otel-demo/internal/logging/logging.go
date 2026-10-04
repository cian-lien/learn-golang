// Package logging 提供一个会"自动补 trace_id / span_id"的 slog Logger。
//
// 日志和链路是两套系统，靠 trace_id 才能互相跳转：
// Jaeger 里看到一条慢链路 -> 用 trace_id 去日志系统里搜 -> 找到当时打出的上下文。
package logging

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

// traceContextHandler 把当前 span 的 id 追加到每条日志上。
type traceContextHandler struct {
	slog.Handler
}

func (h traceContextHandler) Handle(ctx context.Context, r slog.Record) error {
	// 只有 ctx 里带着"有效且已采样"的 span 时才补，否则补出来的 id 在后端查不到。
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		r.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}
	return h.Handler.Handle(ctx, r)
}

// WithAttrs / WithGroup 必须重新包一层，否则 logger.With(...) 之后字段会丢。
func (h traceContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return traceContextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h traceContextHandler) WithGroup(name string) slog.Handler {
	return traceContextHandler{Handler: h.Handler.WithGroup(name)}
}

// New 返回 JSON 格式的结构化日志（生产环境方便被采集器解析）。
func New(level slog.Level) *slog.Logger {
	base := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(traceContextHandler{Handler: base})
}
