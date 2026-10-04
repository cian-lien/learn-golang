package book

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// instrumentationScope 会出现在每个 span/metric 上，用来区分"数据是哪个库产生的"。
const instrumentationScope = "gin-otel-demo/internal/book"

// Service 是业务层：既有业务语义的 span，也有业务指标。
type Service struct {
	store  *Store
	logger *slog.Logger

	tracer   trace.Tracer
	ops      metric.Int64Counter     // 业务操作次数
	duration metric.Float64Histogram // 业务操作耗时
}

func NewService(store *Store, logger *slog.Logger) *Service {
	meter := otel.Meter(instrumentationScope)
	return &Service{
		store:  store,
		logger: logger,
		tracer: otel.Tracer(instrumentationScope),
		ops: must(meter.Int64Counter("book.operations",
			metric.WithDescription("业务操作次数"),
			metric.WithUnit("{operation}"))),
		duration: must(meter.Float64Histogram("book.operation.duration",
			metric.WithDescription("业务操作耗时"),
			metric.WithUnit("ms"))),
	}
}

// Get 一次查询的完整埋点示例：业务 span + 调用 store 产生子 span + 业务指标 + 关联日志。
func (s *Service) Get(ctx context.Context, id string) (*Book, error) {
	start := time.Now()

	// 业务 span：加的是"业务能看懂"的属性，而不是 DB 细节。
	ctx, span := s.tracer.Start(ctx, "book.get",
		trace.WithAttributes(attribute.String("book.id", id)))
	defer span.End()

	b, err := s.store.Get(ctx, id)

	result := "ok"
	switch {
	case errors.Is(err, ErrNotFound):
		result = "not_found"
	case err != nil:
		result = "error"
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	s.record(ctx, "get", result, start)
	span.SetAttributes(attribute.String("book.result", result))

	// 带 ctx 的日志版本，日志里才会自动带上 trace_id。
	// 分级交给 logResult：失败打 ERROR，其余打 INFO。
	s.logResult(ctx, "查询图书", result, err, "book.id", id, "duration_ms", msSince(start))

	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *Service) List(ctx context.Context) ([]Book, error) {
	start := time.Now()
	ctx, span := s.tracer.Start(ctx, "book.list")
	defer span.End()

	books, err := s.store.List(ctx)

	result := "ok"
	if err != nil {
		result = "error"
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	s.record(ctx, "list", result, start)

	s.logResult(ctx, "查询图书列表", result, err, "count", len(books), "duration_ms", msSince(start))

	return books, err
}

func (s *Service) Create(ctx context.Context, title, author string) (*Book, error) {
	start := time.Now()
	ctx, span := s.tracer.Start(ctx, "book.create")
	defer span.End()

	// 业务规则：这里只记录结果，不把用户输入当属性（可能是敏感/高基数数据）。
	created, err := s.store.Insert(ctx, Book{Title: title, Author: author})

	result := "ok"
	if err != nil {
		result = "error"
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	s.record(ctx, "create", result, start)
	span.SetAttributes(attribute.String("book.id", created.ID))

	s.logResult(ctx, "创建图书", result, err, "book.id", created.ID, "duration_ms", msSince(start))

	if err != nil {
		return nil, err
	}
	return &created, nil
}

// logResult 统一日志分级：result=error 打 ERROR（附上 err），其余打 INFO。
// 这样在日志系统里直接筛 level=ERROR 就能捞到所有失败请求，
// 而 not_found 这种"正常的空结果"不会污染错误告警。
func (s *Service) logResult(ctx context.Context, msg, result string, err error, attrs ...any) {
	attrs = append(attrs, "book.result", result)
	if result == "error" {
		s.logger.ErrorContext(ctx, msg, append(attrs, "error", err)...)
		return
	}
	s.logger.InfoContext(ctx, msg, attrs...)
}

// record 统一打点。注意标签只能是低基数的枚举值：
// 把 book.id、用户 id 放进指标标签会让时间序列数量爆炸。
func (s *Service) record(ctx context.Context, op, result string, start time.Time) {
	s.ops.Add(ctx, 1, metric.WithAttributes(
		attribute.String("book.operation", op),
		attribute.String("book.result", result),
	))
	s.duration.Record(ctx, msSince(start),
		metric.WithAttributes(attribute.String("book.operation", op)))
}

func msSince(start time.Time) float64 {
	return float64(time.Since(start).Microseconds()) / 1000
}
