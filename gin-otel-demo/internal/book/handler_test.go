package book

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// newTestRouter 用 SpanRecorder 替换真实的 Exporter：
// 单元测试里不该真的往外发数据，但可以断言"span 到底有没有产生、长什么样"。
func newTestRouter(t *testing.T) (*gin.Engine, *tracetest.SpanRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	otel.SetTracerProvider(tp) // 全局注册，otel.Tracer(...) 才会返回这个 provider 的 tracer
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	r := gin.New()
	r.Use(otelgin.Middleware("gin-otel-demo-test"))
	RegisterRoutes(r, NewService(NewStore(), logger))
	return r, recorder
}

func do(t *testing.T, r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func spanNames(recorder *tracetest.SpanRecorder) map[string]sdktrace.ReadOnlySpan {
	out := make(map[string]sdktrace.ReadOnlySpan)
	for _, s := range recorder.Ended() {
		out[s.Name()] = s
	}
	return out
}

func TestGetBookProducesNestedSpans(t *testing.T) {
	r, recorder := newTestRouter(t)

	w := do(t, r, http.MethodGet, "/api/v1/books/1", "")
	if w.Code != http.StatusOK {
		t.Fatalf("状态码 = %d, 期望 200，body=%s", w.Code, w.Body.String())
	}

	spans := spanNames(recorder)
	for _, want := range []string{
		"GET /api/v1/books/:id", // otelgin 自动产生的 server span
		"book.get",              // service 层手动 span
		"db.books.select",       // store 层手动 span
	} {
		if _, ok := spans[want]; !ok {
			t.Errorf("缺少 span %q，实际产生的是 %v", want, keys(spans))
		}
	}

	// 子 span 的 parent 必须是 server span，链路才是连起来的。
	server := spans["GET /api/v1/books/:id"]
	svc := spans["book.get"]
	if svc.Parent().SpanID() != server.SpanContext().SpanID() {
		t.Errorf("book.get 的父 span 不是 server span：parent=%s server=%s",
			svc.Parent().SpanID(), server.SpanContext().SpanID())
	}
}

func TestGetBookNotFound(t *testing.T) {
	r, _ := newTestRouter(t)

	w := do(t, r, http.MethodGet, "/api/v1/books/9999", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("状态码 = %d, 期望 404", w.Code)
	}
}

func TestStorageFailureMarksSpanAsError(t *testing.T) {
	r, recorder := newTestRouter(t)

	w := do(t, r, http.MethodGet, "/api/v1/books/error", "")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("状态码 = %d, 期望 500", w.Code)
	}

	span, ok := spanNames(recorder)["book.get"]
	if !ok {
		t.Fatal("没有产生 book.get span")
	}
	if span.Status().Code != codes.Error {
		t.Errorf("span 状态 = %v, 期望 Error", span.Status().Code)
	}
	if len(span.Events()) == 0 {
		t.Error("期望 RecordError 产生一个 exception 事件")
	}
}

func TestCreateBook(t *testing.T) {
	r, _ := newTestRouter(t)

	w := do(t, r, http.MethodPost, "/api/v1/books", `{"title":"The Phoenix Project","author":"Kim"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("状态码 = %d, 期望 201，body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "The Phoenix Project") {
		t.Errorf("响应里没有新书：%s", w.Body.String())
	}
}

func keys(m map[string]sdktrace.ReadOnlySpan) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
