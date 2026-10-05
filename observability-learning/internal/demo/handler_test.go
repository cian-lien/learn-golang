package demo

import (
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel/codes"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// 删除埋点、把不同状态混在一起，或把上下文传错，都应使这个测试失败。
func TestRequestsProduceMetricsAndRelatedSpans(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() {
		_ = tp.Shutdown(context.Background())
		_ = mp.Shutdown(context.Background())
	})
	handler, err := NewHandler(tp.Tracer("test"), mp.Meter("test"), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		path   string
		status int
	}{
		{"/hello", 200}, {"/hello", 200}, {"/error", 500}, {"/healthz", 200},
	} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("GET", tc.path, nil))
		if w.Code != tc.status {
			t.Fatalf("%s: status=%d, want %d", tc.path, w.Code, tc.status)
		}
	}
	var data metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &data); err != nil {
		t.Fatal(err)
	}
	counts := map[string]int64{}
	var durationCount uint64
	for _, scope := range data.ScopeMetrics {
		for _, m := range scope.Metrics {
			switch m.Name {
			case "demo.requests":
				for _, point := range m.Data.(metricdata.Sum[int64]).DataPoints {
					route, _ := point.Attributes.Value("http.route")
					status, _ := point.Attributes.Value("http.response.status_code")
					counts[route.AsString()+":"+status.Emit()] = point.Value
				}
			case "demo.request.duration":
				for _, point := range m.Data.(metricdata.Histogram[float64]).DataPoints {
					durationCount += point.Count
				}
			}
		}
	}
	if len(counts) != 2 || counts["/hello:200"] != 2 || counts["/error:500"] != 1 {
		t.Fatalf("request counts=%v, want /hello:200=2 and /error:500=1, no health checks", counts)
	}
	if durationCount != 3 {
		t.Fatalf("duration samples=%d, want 3", durationCount)
	}
	spans := recorder.Ended()
	if len(spans) != 6 {
		t.Fatalf("spans=%d, want 2 for each business request", len(spans))
	}
	byID := map[trace.SpanID]sdktrace.ReadOnlySpan{}
	for _, span := range spans {
		byID[span.SpanContext().SpanID()] = span
	}
	var children, failures int
	for _, span := range spans {
		if span.Status().Code == codes.Error {
			failures++
		}
		if span.Name() != "demo.work" {
			continue
		}
		children++
		parent, ok := byID[span.Parent().SpanID()]
		if !ok || parent.SpanKind() != trace.SpanKindServer || parent.SpanContext().TraceID() != span.SpanContext().TraceID() {
			t.Fatalf("work span is not a child of the HTTP span: %v", span.Parent())
		}
	}
	if children != 3 || failures != 2 {
		t.Fatalf("children=%d error spans=%d, want 3 and 2", children, failures)
	}
}

// 丢掉上游 traceparent 时，跨服务的调用链会断开。
func TestIncomingTraceContextIsPreserved(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	mp := sdkmetric.NewMeterProvider()
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()); _ = mp.Shutdown(context.Background()) })
	handler, err := NewHandler(tp.Tracer("test"), mp.Meter("test"), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET", "/hello", nil)
	req.Header.Set("traceparent", "00-0123456789abcdef0123456789abcdef-0123456789abcdef-01")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Header().Get("X-Trace-ID") != "0123456789abcdef0123456789abcdef" {
		t.Fatalf("trace header=%s", w.Header().Get("X-Trace-ID"))
	}
	for _, span := range recorder.Ended() {
		if span.SpanKind() == trace.SpanKindServer && (span.Parent().SpanID().String() != "0123456789abcdef" || !span.Parent().IsRemote()) {
			t.Fatalf("lost remote parent: %v", span.Parent())
		}
	}
}
