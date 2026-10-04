package book

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// ErrNotFound 是业务错误，handler 会把它翻译成 HTTP 404。
var ErrNotFound = errors.New("book not found")

// errStoreDown 用来模拟一次真实的存储故障（访问 id=error 时触发）。
var errStoreDown = errors.New("simulated storage failure")

// Book 是领域对象，同时也是 HTTP 响应体。
type Book struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

// Store 用内存 map 冒充数据库。它是应用的 IO 边界，
// 所以"数据库 span"打在这一层最合适。
type Store struct {
	mu     sync.RWMutex
	books  map[string]Book
	seq    int
	tracer trace.Tracer
}

func NewStore() *Store {
	s := &Store{
		books:  make(map[string]Book),
		tracer: otel.Tracer("gin-otel-demo/internal/book/store"),
	}
	for _, b := range []Book{
		{Title: "The Go Programming Language", Author: "Donovan & Kernighan"},
		{Title: "Designing Data-Intensive Applications", Author: "Martin Kleppmann"},
		{Title: "Site Reliability Engineering", Author: "Google SRE Team"},
	} {
		s.seq++
		b.ID = fmt.Sprint(s.seq)
		s.books[b.ID] = b
	}
	return s
}

// Get 模拟一次 SELECT。注意 ctx 一路传下来：
// span 的父子关系、超时取消的传播，全靠它。
func (s *Store) Get(ctx context.Context, id string) (Book, error) {
	ctx, span := s.tracer.Start(ctx, "db.books.select",
		trace.WithSpanKind(trace.SpanKindClient), // 出站调用（这里是 DB）用 Client
		trace.WithAttributes(
			attribute.String("db.system.name", "postgresql"),
			attribute.String("db.operation.name", "SELECT"),
			attribute.String("db.collection.name", "books"),
			attribute.String("book.id", id),
		),
	)
	defer span.End()

	if err := sleep(ctx, 3*time.Millisecond); err != nil {
		return Book{}, s.fail(span, err)
	}
	if id == "error" {
		return Book{}, s.fail(span, fmt.Errorf("select book %q: %w", id, errStoreDown))
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	b, ok := s.books[id]
	if !ok {
		// "查不到"不是故障，但记一个事件，排查时能看出是空结果还是压根没查。
		span.AddEvent("db.books.not_found")
		return Book{}, ErrNotFound
	}
	return b, nil
}

// List 模拟一次全表扫描。
func (s *Store) List(ctx context.Context) ([]Book, error) {
	ctx, span := s.tracer.Start(ctx, "db.books.select_all",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system.name", "postgresql"),
			attribute.String("db.operation.name", "SELECT"),
			attribute.String("db.collection.name", "books"),
		),
	)
	defer span.End()

	if err := sleep(ctx, 5*time.Millisecond); err != nil {
		return nil, s.fail(span, err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Book, 0, len(s.books))
	for _, b := range s.books {
		out = append(out, b)
	}
	span.SetAttributes(attribute.Int("db.response.returned_rows", len(out)))
	return out, nil
}

// Insert 模拟一次 INSERT，并返回带自增 ID 的记录。
func (s *Store) Insert(ctx context.Context, b Book) (Book, error) {
	ctx, span := s.tracer.Start(ctx, "db.books.insert",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system.name", "postgresql"),
			attribute.String("db.operation.name", "INSERT"),
			attribute.String("db.collection.name", "books"),
		),
	)
	defer span.End()

	if err := sleep(ctx, 8*time.Millisecond); err != nil {
		return Book{}, s.fail(span, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	b.ID = fmt.Sprint(s.seq)
	s.books[b.ID] = b

	span.SetAttributes(attribute.String("book.id", b.ID))
	return b, nil
}

// fail 统一处理"记录错误"这件事：
// 只 return err 是不会让 span 变红的，必须 RecordError + SetStatus。
func (s *Store) fail(span trace.Span, err error) error {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	return err
}
