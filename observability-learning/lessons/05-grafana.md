# 第 05 课：用 Grafana 观察变化

目标：把上一课的四条查询放在同一看板上，并能用图表解释一次故障实验。

## 打开看板

保持后端和 `go run . -exporter=otlp` 运行。打开 http://127.0.0.1:13000 ，
使用 `admin` / `learning-admin` 登录。在 Dashboards 中打开 `Go Learning / Go Observability Learning`。

四个图表分别展示：每条路由的请求速率、500 错误率、平均耗时、估算 P95。
数据源已经通过 `config/grafana/provisioning/datasources/prometheus.yaml` 配置，
Grafana 容器查询的是 `http://prometheus:9090`。

## 制造流量并解释

```bash
bash scripts/traffic.sh 120
```

确认时间范围为最近 15 分钟，刷新间隔为 5 秒。至少等两个 Prometheus 抓取周期。

- 正常与故障请求都会增加请求次数；错误率图应该接近 33%。
- `/slow` 平均耗时约 250ms；其他路由更快。
- P95 来源于桶内估算，与平均值不同，也未必等于实际固定延迟。

如果出现 No data，依次检查：应用是否启动且收到请求、Collector 的 `/metrics` 是否有 `demo_`、Prometheus 是否能查到同名指标、Grafana 数据源是否正确。

## 看图背后的查询

进入某个图表的编辑界面，找到其 PromQL。把它复制到 Prometheus 页面，结果应表达同一组数据。
Grafana 展示和查询数据，指标历史保存在 Prometheus 中。

文件导入的看板可以在 UI 中编辑，但下次重新导入可能覆盖 UI 的修改；要保留实验成果，导出到新的 JSON 或修改 `config/grafana/dashboards/learning.json`。

## 综合实验

1. 先运行一轮正常、慢、失败流量，说明每个图表为什么呈现当前形状。
2. 在 `internal/demo/handler.go` 把 250ms 改成 500ms，重启 Go 应用，再运行流量。
3. 判断哪些图表会变化、哪些比例应基本不变；注意 Counter 随进程重启归零，但 rate 会处理重置。
4. 切回 `go run . -exporter=console`，只请求 `/slow`，用 span 起止时间确认延迟发生在 `demo.work`。

本阶段练习的是“指标发现问题，调用链解释一次请求”。调用链暂时只在终端输出，没有配置持久化查询或 Grafana 的 Trace 数据源。

原理参考：[Grafana provisioning](https://grafana.com/docs/grafana/latest/administration/provisioning/)。
