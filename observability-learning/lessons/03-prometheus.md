# 第 03 课：从 Go 到 Prometheus

目标：亲眼看到推送和拉取分别发生在哪一段。

## 启动

先停止上一课应用。项目根目录执行：

```bash
docker compose up -d
go run . -exporter=otlp
```

应用发送到宿主机 `127.0.0.1:14317`，Docker 把它转发给 Collector 的 4317 端口。
另开终端制造流量：

```bash
bash scripts/traffic.sh 60
```

## 第一站：Collector

应用每 3 秒用 OTLP/gRPC 推送指标，Collector 接收并批处理，然后通过 Prometheus exporter 暴露 `/metrics`。

```bash
curl -s http://127.0.0.1:18889/metrics | rg '^demo_'
```

在配置的命名转换策略下，会看到类似：

```text
demo_requests_total{http_route="/hello",http_response_status_code="200",...} 60
demo_request_duration_seconds_count{http_route="/slow",...} 60
demo_request_duration_seconds_sum{http_route="/slow",...} 15...
demo_request_duration_seconds_bucket{http_route="/slow",le="0.5",...} 60
```

省略号表示还有其他标签或数字，小数和标签顺序可能不同。应用未重启时次数会继续累计。
OTel 的点号会转成下划线；Counter 增加 `_total`；耗时使用秒，因此名称中带 `_seconds`。
指标没有出现时，先发请求，再等 5–10 秒。

## 第二站：Prometheus

打开 http://127.0.0.1:19090 ，查询：

```promql
up{job="learning-app"}
```

预期为 1，表示 Prometheus 最近一次成功抓取了 Collector。它不能证明 Go 应用当前健康；应用已停止而 Collector 仍可抓取时，也可能是 1。

再查询：

```promql
demo_requests_total{job="learning-app"}
```

现在指标进入了 Prometheus 的时间序列数据库，可以查看历史。
打开 `config/prometheus.yaml`：`scrape_interval: 5s` 控制抓取间隔，`collector:8889` 是容器网络内的地址。

数据流：Go 主动推送到 Collector；Prometheus 定时拉取 Collector；Grafana 查询 Prometheus。
暴露一个 `/metrics` 页面本身不等于已经把历史指标存进 Prometheus。

## 练习

- 改抓取间隔为 10s，执行 `docker compose restart prometheus`，观察采样点的间距。
- 停止 Go 应用但保留 Collector，观察 `up` 和业务 Counter。为什么两者表达的健康含义不同？

排查命令：`docker compose ps`、`docker compose logs --tail=30 collector`。
原理参考：[Collector Prometheus exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.161.0/exporter/prometheusexporter)。
