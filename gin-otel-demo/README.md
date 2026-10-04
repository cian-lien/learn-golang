# Gin + OpenTelemetry 分步示例

一个**独立的** Go Web 服务：Gin 提供 HTTP API，OpenTelemetry 负责把链路（trace）、指标（metric）、
日志（log）送到 Collector / Jaeger。README 按"从零到能看到链路"的顺序，把每一步单独讲清楚。

## 项目结构

```text
gin-otel-demo/
├── go.mod / go.sum            独立模块，不依赖任何外部仓库
├── main.go                    第 1 步：配置、路由组装、优雅退出
├── internal/
│   ├── telemetry/telemetry.go 第 3 步：OTel SDK 装配（唯一直接依赖 SDK 的地方）
│   ├── logging/logging.go     第 7 步：slog 自动补 trace_id / span_id
│   └── book/                  第 5、6 步：三层各自的埋点示例
│       ├── handler.go         Gin 路由与参数/错误翻译（本身不埋点）
│       ├── service.go         业务 span + 业务指标
│       ├── store.go           模拟 DB 的 IO span
│       ├── helpers.go         must / sleep
│       └── handler_test.go    第 11 步：用 SpanRecorder 断言 span
├── docker-compose.yml         第 9 步：Jaeger + Collector
├── otel-collector.yaml        Collector 的 pipeline 配置
└── README.md
```

## 数据流

```text
curl
 └─> gin-otel-demo（本机 go run）
      ├─ trace/metric ──OTLP/gRPC:4317──> otel-collector ──OTLP──> jaeger（UI :16686）
      │                                                 └─Prometheus 格式──> :8889/metrics
      └─ 结构化日志（JSON，带 trace_id）──> 标准输出 / 未来的 Loki、ELK
```

---

## 第 0 步：准备与概念

- Go 1.25+（本项目 go.mod 声明 `go 1.25.0`）
- Docker（可选，只有第 9 步起后端时会用到；不做也能用 console 模式看数据）

三个必须先分清的词：

| 概念 | 是什么 | 在本项目里 |
| --- | --- | --- |
| **Trace** | 一次请求的完整调用链，由一串 span 组成 | `GET /api/v1/books/:id` → `book.get` → `db.books.select` |
| **Metric** | 可聚合的数值，用于告警和看板 | `book.operations`（计数器）、`book.operation.duration`（直方图） |
| **Log** | 离散事件文本 | `slog` 打的 JSON 日志，带 `trace_id` |

还要记住 OTel 的 **API / SDK 分离**：业务代码只依赖 API（`otel.Tracer` / `otel.Meter`），
SDK（采样、导出、批量）只在启动时装配一次。这样换后端、调采样率都不用改业务代码。

---

## 第 1 步：创建项目并跑起最小的 Gin 服务

```bash
mkdir gin-otel-demo && cd gin-otel-demo
go mod init gin-otel-demo
go get github.com/gin-gonic/gin@latest
```

最小可运行的 `main.go`：

```go
package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.Run(":8080")
}
```

```bash
go run . & curl -i localhost:8080/healthz
```

到这一步，出问题时你只有一行行日志：**哪个请求慢、慢在哪一层、有没有报错，全靠猜**。
后面的步骤就是把"猜"变成"看"。本项目的 `main.go` 在这个骨架上加了：
配置（`-addr` / 环境变量）、`http.Server`（为了优雅退出，而不是 `r.Run`）、`internal/book` 路由。

---

## 第 2 步：安装 OpenTelemetry 依赖

```bash
go get go.opentelemetry.io/otel@latest \
       go.opentelemetry.io/otel/trace@latest \
       go.opentelemetry.io/otel/metric@latest \
       go.opentelemetry.io/otel/sdk@latest \
       go.opentelemetry.io/otel/sdk/metric@latest \
       go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc@latest \
       go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc@latest \
       go.opentelemetry.io/otel/exporters/stdout/stdouttrace@latest \
       go.opentelemetry.io/otel/exporters/stdout/stdoutmetric@latest \
       go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin@latest
```

（懒人做法：直接写代码，然后 `go mod tidy` 让 Go 根据 import 自动补全。）

每个包做什么：

| 包 | 角色 |
| --- | --- |
| `otel` | API 入口：`otel.Tracer()` / `otel.Meter()` / 全局 provider 注册 |
| `otel/trace`、`otel/metric` | span、tracer、meter、instrument 的类型定义 |
| `otel/sdk` | TracerProvider / SpanProcessor / Sampler / Resource |
| `otel/sdk/metric` | MeterProvider / Reader |
| `.../exporters/otlp/otlptrace/otlptracegrpc` | 用 OTLP/gRPC 发 **trace** 给 Collector |
| `.../exporters/otlp/otlpmetric/otlpmetricgrpc` | 用 OTLP/gRPC 发 **metric** |
| `.../exporters/stdout/stdouttrace`、`stdoutmetric` | 打到标准输出——本地调试神器，零依赖 |
| `.../contrib/instrumentation/.../otelgin` | Gin 的自动埋点中间件 |

本项目实际解析到的版本：`gin v1.12.0`、`otel v1.46.0`、`otelgin v0.71.0`。
注意 contrib 的版本号和 SDK 不是一回事（otelgin v0.71 配 otel v1.46）。

---

## 第 3 步：初始化 SDK（`internal/telemetry/telemetry.go`）

这一步是整个接入的地基，也是最值得抄进真实项目的一段。按顺序做六件事：

**① Resource：说明"数据是谁产生的"**

```go
res, err := resource.New(ctx,
	resource.WithProcess(), resource.WithOS(), resource.WithHost(), resource.WithContainer(),
	resource.WithAttributes(
		semconv.ServiceNameKey.String(cfg.ServiceName),          // 最关键，缺了会变成 unknown_service
		semconv.ServiceVersionKey.String(cfg.ServiceVersion),
		semconv.DeploymentEnvironmentNameKey.String(cfg.Environment),
	),
)
```

**② Exporter：数据发到哪**

```go
if cfg.Console { // OTEL_EXPORTER=console：零依赖调试
	spanExporter, _ = stdouttrace.New(stdouttrace.WithPrettyPrint())
} else {
	spanExporter, _ = otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint), // 127.0.0.1:4317
		otlptracegrpc.WithInsecure())         // 本地明文；生产用 TLS
}
```

**③ Sampler：采多少**

```go
sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRatio))
```

`ParentBased` 是必须的：父 span 被采样时子 span 一定被采样，否则一条链路只会剩下半截。

**④ 注册全局 Provider**

```go
otel.SetTracerProvider(sdktrace.NewTracerProvider(
	sdktrace.WithResource(res),
	sdktrace.WithSampler(sampler),
	sdktrace.WithBatcher(spanExporter), // 攒批发送
))
otel.SetMeterProvider(metric.NewMeterProvider(
	metric.WithResource(res),
	metric.WithReader(metric.NewPeriodicReader(metricExporter, metric.WithInterval(10*time.Second))),
))
```

**⑤ 传播器：链路怎么跨服务串起来**

```go
otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
	propagation.TraceContext{}, // W3C traceparent / tracestate
	propagation.Baggage{},      // 业务上下文透传
))
```

**⑥ 打开错误日志**

```go
otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
	slog.Error("opentelemetry sdk error", "error", err)
}))
```

SDK 默认静默吞掉导出错误。没这行，"后端为什么没数据"能查到你怀疑人生。

**别忘了 Shutdown**：`Setup` 返回一个 `shutdown func(context.Context) error`，
里面 `errors.Join(tracerProvider.Shutdown(ctx), meterProvider.Shutdown(ctx))`。
span 和 metric 都是异步批量发送的，进程退出前不调用就等于丢掉最后一批数据。

---

## 第 4 步：接上 Gin —— 一行中间件

`main.go` 里路由组装的部分：

```go
func newRouter(serviceName string, logger *slog.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(gin.Recovery())            // 最外层兜 panic，保证 span 能正常收尾
	r.Use(otelgin.Middleware(serviceName,
		otelgin.WithFilter(func(req *http.Request) bool {
			return req.URL.Path != "/healthz" // 健康检查太吵，不采集
		}),
	))
	r.Use(requestLogger(logger))     // 普通访问日志，用 InfoContext 才能带 trace_id

	book.RegisterRoutes(r, book.NewService(book.NewStore(), logger))
	return r
}
```

`otelgin.Middleware` 自动完成了这些事，你一行都不用写：

1. 从请求头提取 `traceparent`，把上游 span 接成父节点（没有上游就新建根 span）；
2. 创建 **server span**，span 名用 `c.FullPath()`，即 `GET /api/v1/books/:id` —— 天然低基数；
3. 记录 HTTP 语义约定属性：`http.request.method`、`url.path`、`http.response.status_code`…；
4. 自动产出 HTTP 相关指标（`http.server.request.duration` 等）；
5. 请求结束时结束 span，并按状态码设置 span 状态（5xx → Error）。

---

## 第 5 步：业务层手动埋点（span 的父子关系）

自动埋点只覆盖"边界"（HTTP 出入口、DB 驱动），**业务内部发生了什么必须自己埋**。
本项目分两层：

**store（IO 边界，`internal/book/store.go`）**

```go
ctx, span := s.tracer.Start(ctx, "db.books.select",
	trace.WithSpanKind(trace.SpanKindClient), // 出站调用用 Client
	trace.WithAttributes(
		attribute.String("db.system.name", "postgresql"),
		attribute.String("db.operation.name", "SELECT"),
		attribute.String("db.collection.name", "books"),
	),
)
defer span.End()
```

**service（业务语义，`internal/book/service.go`）**

```go
ctx, span := s.tracer.Start(ctx, "book.get",
	trace.WithAttributes(attribute.String("book.id", id)))
defer span.End()

b, err := s.store.Get(ctx, id) // ← ctx 一定要往下传！
...
span.RecordError(err)                        // 记下错误（会生成 exception 事件）
span.SetStatus(codes.Error, err.Error())     // 把 span 标红
```

四条必须记住的规则：

1. **ctx 一路传**。handler 里用 `c.Request.Context()`，以后每一层都用上层传来的 ctx；
   断在这里，后面所有 span 都会变成没有父节点的孤儿。
2. **`return err` 不会让 span 变红**。必须 `RecordError` + `SetStatus(codes.Error, ...)`，
   否则后端看到的还是"成功"，这三点是本示例 `TestStorageFailureMarksSpanAsError` 专门验证的。
3. **span 名要低基数**：业务动作用 `book.get` 这种固定名，不要拼 id 进去。
4. **属性用语义约定**：`db.*`、`http.*`、`service.*` 这些名字，Grafana/Jaeger 能自动识别并展示。

不需要"错误"语义但想留痕的中间过程，用事件：

```go
span.AddEvent("db.books.not_found")
```

---

## 第 6 步：自定义指标

指标和 span 的区别：span 记录**这一次**发生了什么，指标记录**很多次之后**的聚合规律。

```go
meter := otel.Meter("gin-otel-demo/internal/book")

ops := must(meter.Int64Counter("book.operations",
	metric.WithDescription("业务操作次数"), metric.WithUnit("{operation}")))
dur := must(meter.Float64Histogram("book.operation.duration",
	metric.WithDescription("业务操作耗时"), metric.WithUnit("ms")))

ops.Add(ctx, 1, metric.WithAttributes(
	attribute.String("book.operation", op),   // get / list / create
	attribute.String("book.result", result),  // ok / not_found / error
))
dur.Record(ctx, elapsedMS, metric.WithAttributes(attribute.String("book.operation", op)))
```

**标签基数是最容易翻车的地方**：`book.operation` 只有 3 个取值，`book.result` 只有 3 个取值，
组合数可控。如果把 `book.id`、用户 ID 也当标签，时间序列数量会随数据量线性爆炸，
Prometheus 内存会被吃光——所以示例里 id 只放进 **span 属性**，不放进**指标标签**。

---

## 第 7 步：让日志带上 trace_id

日志和链路是两套系统，靠 `trace_id` 才能互相跳转。`internal/logging/logging.go` 用
一个 slog handler 实现：

```go
func (h traceContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		r.AddAttrs(slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()))
	}
	return h.Handler.Handle(ctx, r)
}
```

配套的代码习惯只有一个：**打日志时用带 ctx 的版本**。

```go
s.logger.InfoContext(ctx, "查询图书", "book.id", id, "book.result", result) // ✅ 带 trace_id
s.logger.Info("查询图书", "book.id", id)                                    // ❌ 不带
```

效果（实际输出）：

```json
{"msg":"查询图书","book.id":"1","book.result":"ok","duration_ms":3.221,"trace_id":"6bc36212d0d9f99500091f27b5c3279c"}
```

---

## 第 8 步：优雅退出与 Shutdown 顺序

`main.go` 里的顺序不能颠倒：

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
...
select {
case err := <-serverErr: return err
case <-ctx.Done():       // 收到 Ctrl-C / SIGTERM
}

srv.Shutdown(shutdownCtx)        // 1. 先停止接收新请求、等在途请求处理完
shutdownTelemetry(shutdownCtx)   // 2. 再 flush 并关闭遥测
```

先关遥测的话，正在处理的请求结束时 span 已经无处可发。注意 `Shutdown(ctx)` 也要给超时
（本项目 10s），避免卡死退出流程。

---

## 第 9 步：启动后端（Collector + Jaeger）

```bash
docker compose up -d          # jaeger(:16686) + otel-collector(:4317/:4318/:8889)
```

`otel-collector.yaml` 的三个概念——**receiver → processor → exporter**：

```yaml
receivers:  { otlp: { protocols: { grpc: {endpoint: 0.0.0.0:4317}, http: {endpoint: 0.0.0.0:4318} } } }
processors: { memory_limiter: {...}, batch: {...} }
exporters:
  otlp/jaeger: { endpoint: jaeger:4317, tls: { insecure: true } }  # trace 存 Jaeger
  prometheus:  { endpoint: 0.0.0.0:8889 }                          # metric 暴露成 Prometheus 格式
  debug:       { verbosity: basic }                                # 打到 collector 自己的日志
service:
  pipelines:
    traces:  { receivers: [otlp], processors: [memory_limiter, batch], exporters: [otlp/jaeger, debug] }
    metrics: { receivers: [otlp], processors: [memory_limiter, batch], exporters: [prometheus, debug] }
```

为什么不直连 Jaeger？因为 Collector 是"可观测性数据网关"：换后端、加尾采样、脱敏、
限流都改配置即可。`memory_limiter` 必须排在 `batch` 前面，否则内存先爆。

应用侧的环境变量：

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `OTEL_EXPORTER` | `otlp` | 设为 `console` 时改打标准输出，不需要任何后端 |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `127.0.0.1:4317` | Collector 的 OTLP/gRPC 地址 |
| `OTEL_SERVICE_NAME` | `gin-otel-demo` | 覆盖 `service.name` |
| `DEPLOY_ENV` | `development` | 写进 `deployment.environment.name` |
| `OTEL_SAMPLE_RATIO` | `1.0` | 根 span 采样比例 |
| `OTEL_METRIC_INTERVAL_MS` | `10000` | 指标导出间隔 |
| `OTEL_SDK_DISABLED` | `false` | `true` 时全局降级为 noop（线上紧急关停） |
| `APP_ADDR` / `-addr` | `:8080` | 监听地址 |

---

## 第 10 步：跑起来并验证

```bash
# 终端 1：后端
docker compose up -d

# 终端 2：应用
go run .                     # 默认 OTLP 发给 Collector；加 OTEL_EXPORTER=console 可零依赖

# 终端 3：造点流量
curl -s localhost:8080/api/v1/books                    # 列表
curl -s localhost:8080/api/v1/books/1                  # 单查
curl -s -X POST localhost:8080/api/v1/books \
     -H 'Content-Type: application/json' \
     -d '{"title":"Accelerate","author":"Forsgren"}'   # 新增
curl -s -o /dev/null -w '%{http_code}\n' localhost:8080/api/v1/books/9999   # 404
curl -s -o /dev/null -w '%{http_code}\n' localhost:8080/api/v1/books/error  # 500（故意失败）
```

看结果的三条命令：

```bash
# ① Jaeger：Service 选 gin-otel-demo（UI 在 http://localhost:16686）
curl -s http://localhost:16686/api/v3/services
# {"services":["gin-otel-demo"]}

# 注意：Jaeger v2 的查询 API 是 /api/v3/*，老教程里的 /api/services 会 404

# ② Collector 暴露的指标
curl -s http://localhost:8889/metrics | grep book_operations_total

# ③ 日志里的 trace_id（应用标准输出）
# {"msg":"查询图书","book.result":"error","trace_id":"0d26718dcab5..."}
```

在本机实测到的链路（每个 HTTP 请求一棵树，同一个 trace_id）：

```text
GET  /api/v1/books          → book.list   → db.books.select_all
GET  /api/v1/books/:id      → book.get    → db.books.select
POST /api/v1/books          → book.create → db.books.insert
GET  /api/v1/books/error    → book.get    → db.books.select    （status=Error，带 exception 事件）
```

---

## 第 11 步：给埋点写测试

埋点也是代码，也要测。用 `tracetest.SpanRecorder` 把 span 抓下来断言，
不需要真的启动 Collector（见 `internal/book/handler_test.go`）：

```go
recorder := tracetest.NewSpanRecorder()
tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
otel.SetTracerProvider(tp)

r := gin.New()
r.Use(otelgin.Middleware("gin-otel-demo-test"))
RegisterRoutes(r, NewService(NewStore(), logger))

// ... 发请求 ...
spans := recorder.Ended()
// 断言：是否产生了 GET /api/v1/books/:id、book.get、db.books.select，
// 以及 book.get 的父 span 是不是 server span
```

```bash
go test ./... -v
```

本项目 4 个用例全部通过：正常查询的 span 树、404、存储故障时 span 状态为 Error、创建图书。

---

## 第 12 步：怎么看错误

错误在三个地方各有一份，排查顺序是 **指标发现 → 链路定位 → 日志看细节**，
三者靠 `trace_id`（日志）、`traceId`（链路）串起来。

### ① 指标：先知道"有没有错、错多少"（适合告警）

```bash
curl -s http://localhost:8889/metrics | grep -E '^book_operations_total|^http_server_request_duration_seconds_count'
```

```text
book_operations_total{book_operation="get",book_result="error",...} 1
http_server_request_duration_seconds_count{http_response_status_code="500",http_route="/api/v1/books/:id",...} 1
```

告警规则就写在这上面，例如 `sum(rate(book_operations_total{book_result="error"}[5m])) > 0`。
注意这类 Counter 是**累计值**，进程重启才会归零；不同实例靠 `service.name` / `process.pid` 等标签区分。

### ② 链路：定位"错在哪一层"（适合定位）

Jaeger UI（<http://localhost:16686>）里：Service 选 `gin-otel-demo`，**Tags 填 `error=true`**，点 Find Traces。
打开链路后：出错的 span 是红色的，点开右侧面板看 `Status` 和 `Events` 里的 `exception`。

命令行等价写法（`error=true` 是 Jaeger 根据 span 状态合成出来的，不是代码里写的属性）：

```bash
MAX=$(date -u +%Y-%m-%dT%H:%M:%SZ); MIN=$(date -u -v-10M +%Y-%m-%dT%H:%M:%SZ)
curl -s "http://localhost:16686/api/v3/traces?query.serviceName=gin-otel-demo\
&query.startTimeMin=$MIN&query.startTimeMax=$MAX&query.numTraces=10\
&query.attributes=%7B%22error%22%3A%22true%22%7D" \
| jq -r '.result.resourceSpans[].scopeSpans[].spans[] | "\(.traceId[0:12])… \(.name) \(.status.message // "")"'
```

也可以按**自己写的属性**筛，比如把所有"没找到"的请求捞出来：

```bash
# query.attributes={"book.result":"not_found"}
&query.attributes=%7B%22book.result%22%3A%22not_found%22%7D
```

已知 `trace_id` 时直接查单条链路：`curl -s http://localhost:16686/api/v3/traces/<trace_id>`。

### ③ 日志：看"到底报了什么错"（适合看细节）

```bash
go run . 2>&1 | grep '"level":"ERROR"'
```

```json
{"level":"ERROR","msg":"查询图书","book.id":"error","duration_ms":3.48,
 "book.result":"error","error":"select book \"error\": simulated storage failure",
 "trace_id":"2770da8f072bb47e0fafb1ff574c9e14","span_id":"63c52984b14c4d0b"}
```

三个关键点：

- 末尾的 `trace_id` 可以直接拿去 Jaeger 查那一条链路（这就是第 7 步那个 slog handler 的价值）；
- `error` 字段是原始错误链，`errors.Is/As` 判断用的也是它；
- 只有 `book.result=error` 才打 ERROR，`not_found` 是 INFO —— 避免"正常的空结果"淹没真正的错误。

### 三个必须知道的坑

1. **4xx 不会被标红**。按 HTTP 语义约定，服务端只有 5xx 算 Error，404/400 的 span 状态是 `unset`
   （实测：`GET /api/v1/books/9999` → 404，链路全绿）。业务上的"没找到"要用自己的属性/指标
   （`book.result`）表达，别指望红色 span。
2. **`return err` 不会让 span 变红**，必须 `span.RecordError(err)` + `span.SetStatus(codes.Error, ...)`，
   参考 `store.fail`。只 return 不改状态，后端看到的仍是"成功"。
3. **什么都看不到时**，按这个顺序查：
   - 应用日志里有没有 `opentelemetry sdk error`（导出失败，来自 `otel.SetErrorHandler`）；
   - `OTEL_SDK_DISABLED=true` 没被误设、`OTEL_SAMPLE_RATIO` 没被调成 0；
   - Collector 到底收到没有：`docker compose logs otel-collector | grep -i "Traces"`；
   - Jaeger 里 Service 列表有没有你的服务：`curl -s http://localhost:16686/api/v3/services`。

---

## 生产环境 checklist

- [ ] **采样**：`ParentBased` 之外通常是"默认按比例 + 错误/慢请求全采"，尾采样放在 Collector 做。
- [ ] **TLS**：`otlptracegrpc.WithTLSCredentials(...)` 替代 `WithInsecure()`。
- [ ] **Collector 部署形态**：每台机器一个 agent（贴近应用、做批处理和贴标签），
      或集中 gateway（做尾采样、脱敏、路由）。
- [ ] **资源属性**：补上 `service.version`、`deployment.environment.name`、k8s 的
      `k8s.pod.name` / `k8s.namespace.name`，否则多环境数据混在一起没法区分。
- [ ] **告警**：用指标（`book.operations{book.result="error"}`）而不是从 trace 里捞。
- [ ] **日志**：如果日志也要走 OTLP，可以把 `slog` 接上 OTel 的 log bridge，
      统一由 Collector 分发到 Loki/ES。
- [ ] **不要**把 PII、token、完整 SQL 当属性写进 span——它们会被长期存储。

## 常见坑速查

| 现象 | 原因 |
| --- | --- |
| 后端里是 `unknown_service` | 没设 `service.name` |
| 只有一条 span，没有子 span | 手动埋点时没把 `ctx` 传下去 |
| 链路跨服务断掉 | 出站请求没用 `otelhttp`/带传播器的 client，或没传请求头 |
| 报错了但 span 是绿的 | 只 `return err`，没 `RecordError` + `SetStatus` |
| 最后一批数据没了 | 退出前没调用 provider 的 `Shutdown` |
| 导出失败却毫无提示 | 没设置 `otel.SetErrorHandler` |
| 后端里 span 名成千上万 | 用真实路径/ID 命名，没用路由模式 |
| Prometheus 内存暴涨 | 指标标签用了高基数（user id、订单号） |
| `go run` 起来后 Jaeger 里没数据 | Collector 没起，或 `OTEL_EXPORTER_OTLP_ENDPOINT` 写错 |
