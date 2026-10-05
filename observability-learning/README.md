# 用 Go 学习 OTel、Prometheus、Grafana

这是一个独立 Go 模块。每课都围绕同一组接口，通过实验把数据产生、传输、存储和展示串起来。
先按顺序学；完整配置已备好，第一课只需要 Go。

## 学习路线

| 课次 | 目标 | 学习资料 |
| --- | --- | --- |
| 01 | 理解 HTTP 请求、状态码、耗时 | [普通 HTTP 服务](lessons/01-http.md) |
| 02 | 用 OTel 产生 span、Counter、Histogram | [控制台埋点](lessons/02-otel.md) |
| 03 | Collector 接收数据，Prometheus 拉取并存储指标 | [数据管道](lessons/03-prometheus.md) |
| 04 | 用 PromQL 算吞吐、错误率、平均耗时和 P95 | [查询练习](lessons/04-promql.md) |
| 05 | 在 Grafana 观察仪表盘，完成故障实验 | [Grafana](lessons/05-grafana.md) |

## 第一课从这里开始

要求：Go 1.25+。所有命令都在本目录执行。

```bash
cd /Users/cianlien/development/learning/go/observability-learning
go run ./cmd/01-http
```

另开终端：

```bash
curl -i http://127.0.0.1:18090/hello
curl -s -o /dev/null -w 'status=%{http_code} duration=%{time_total}s\n' http://127.0.0.1:18090/slow
curl -i http://127.0.0.1:18090/error
```

先解释这三个结果，再进入下一课。按 Ctrl-C 停止服务。

## 完整实验怎么运行

第三课起需要 Docker 和 Docker Compose。先停止第一课服务，释放 18090 端口。

```bash
docker compose up -d
go run . -exporter=otlp
```

另开终端持续制造流量：

```bash
bash scripts/traffic.sh 120
```

| 入口 | 地址 / 用途 |
| --- | --- |
| Go 应用 | http://127.0.0.1:18090 |
| Collector 接收 OTLP/gRPC | 127.0.0.1:14317 |
| Collector 指标端点 | http://127.0.0.1:18889/metrics |
| Prometheus | http://127.0.0.1:19090 |
| Grafana | http://127.0.0.1:13000 ，账号 `admin`，密码 `learning-admin` |

Grafana 自动导入 `Go Learning / Go Observability Learning` 看板。首次启动后，等至少两个抓取周期再查询 `rate`；停止流量后，速率也会逐渐下降。

```text
Go 应用中的 OTel API / SDK
  ├─ metrics ─OTLP 推送─> Collector ─被定时抓取─> Prometheus <─查询─ Grafana
  ├─ traces ─OTLP 推送─> Collector 的 debug 输出
  └─ logs ─> 应用标准输出（附 trace_id）
```

这里的 Grafana 看板只展示指标。调用链在第二课的终端中学习，本项目未配置 Trace 或 Log 的持久化后端。

## 文件导读

```text
cmd/01-http/main.go       第一课：仅标准库的起点
main.go                  第二课起：启动服务、选择 exporter、优雅退出
internal/demo/handler.go  span / counter / histogram 的手动埋点
internal/telemetry/      OTel SDK 装配与导出
config/collector.yaml    接收、批处理、导出管道
config/prometheus.yaml   抓取目标和间隔
config/grafana/          数据源和四个图表
lessons/                 每课的步骤、观察点、思考题
scripts/traffic.sh       可重复的正常、慢、故障流量
```

三个业务接口是固定路由。`/healthz`、未知路径和不支持的方法不计入示例的业务指标；这不是完整的生产 HTTP 监控中间件。

## 验证和停止

```bash
go test -race ./...
go vet ./...
docker compose config -q
```

应用按 Ctrl-C 退出；后端用 `docker compose down` 停止。命名卷保留指标历史和 Grafana 设置。

## 官方参考

- [OTel Go 手动埋点](https://opentelemetry.io/docs/languages/go/instrumentation/)
- [Collector Prometheus exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.161.0/exporter/prometheusexporter)
- [PromQL 函数](https://prometheus.io/docs/prometheus/latest/querying/functions/)
- [Grafana 配置导入](https://grafana.com/docs/grafana/latest/administration/provisioning/)

版本固定在 go.mod 和 compose.yaml，避免课程随着 latest 标签变化。容器版本对应官方发布：
[Collector 0.161.0](https://github.com/open-telemetry/opentelemetry-collector-releases/releases/tag/v0.161.0)、
[Prometheus 3.14.0](https://github.com/prometheus/prometheus/releases/tag/v3.14.0)、
[Grafana 13.1.6](https://github.com/grafana/grafana/releases/tag/v13.1.6)。
