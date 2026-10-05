# 第 02 课：让 Go 程序产生观测数据

目标：区分一次请求的 Trace 和多次请求的 Metric，找到真正产生数据的 Go 调用。

## 运行

先停止第一课服务，再在项目根目录执行：

```bash
go run . -exporter=console
```

另开终端：

```bash
curl -i http://127.0.0.1:18090/hello
curl -i http://127.0.0.1:18090/slow
curl -i http://127.0.0.1:18090/error
```

回到应用终端：span 约 1 秒后批量输出，指标约每 3 秒输出一次。JSON 很长，先找 `Name`。

## Trace：这一请求经过了什么

每个业务请求有两个 span：

```text
GET /hello（server span）
└── demo.work（业务 span）
```

在 `internal/demo/handler.go` 找到 `tracer.Start`、`defer span.End()` 和 `work(ctx, tracer, route)`。
两个 span 共享 TraceID，各有不同 SpanID；业务 span 的 Parent.SpanID 指向 server span。
通过 ctx 传递当前 span，子调用才能建立父子关系。上游 traceparent 也通过请求头提取。

失败时需要 `RecordError(err)` 记录错误事件，并用 `SetStatus(codes.Error, ...)` 表示失败状态。
本课 `/error` 的两个 span 都会标记 Error。注意：未显式标记成功的正常 span 可显示 Unset，这不是故障。

响应头 `X-Trace-ID` 与 JSON 访问日志的 `trace_id` 可以帮助你找到同一次请求。
本项目日志由 slog 输出；我们没有接入 OTel Log SDK。

## Metric：多次请求如何聚合

继续在同一文件寻找：

```go
requests.Add(ctx, 1, labels)
duration.Record(ctx, time.Since(start).Seconds(), labels)
```

| 数据 | OTel 名称 | 类型 | 用途 |
| --- | --- | --- | --- |
| 请求次数 | demo.requests | Counter | 每处理一个请求加 1 |
| 请求耗时 | demo.request.duration | Histogram | 记录耗时分布，单位秒 |

标签包含固定路由和状态码，所以 `/hello` 的 200 与 `/error` 的 500 是不同的数据点。
Counter 输出的是累计次数；每 3 秒输出一次不会自动把它清零。Histogram 会聚合 Count、Sum、Buckets，默认不会保存每次请求的原始耗时。

`/healthz` 不计入业务数据。不要给 Metric 加 trace_id、user_id 这种不断变化的标签。

## SDK 和 API

业务文件使用 tracer / meter API。`internal/telemetry/setup.go` 则负责选择 exporter、创建 SDK Provider、配置批处理和直方图桶边界。
`main.go` 把 Provider 提供的 tracer / meter 传给业务层，退出时先停 HTTP，再 flush 遥测。

## 练习

- 再请求 `/hello` 两次：它的 Counter 应从 1 变成多少？
- 找到 `/slow` 的两个 span，比较起止时间。
- 为什么 trace_id 适合放进日志，却不适合成为指标标签？

原理参考：[OTel Go 埋点](https://opentelemetry.io/docs/languages/go/instrumentation/)。
