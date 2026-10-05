# 第 04 课：把数字变成判断

目标：能够解释请求速率、错误率、平均耗时与 P95 各自回答什么问题。
继续运行上一课的后端和 Go 应用，执行 `bash scripts/traffic.sh 120`。
在 Prometheus 页面逐条查询；`rate` 至少需要同一序列的两个采样点，刚启动时可先等 15 秒。

## 从累计值到每秒次数

```promql
sum by (http_route) (
  rate(demo_requests_total{job="learning-app"}[1m])
)
```

`[1m]` 取最近一分钟的数据；`rate` 计算 Counter 的每秒平均增长率，并处理计数器重置。
`sum by (http_route)` 合并其他标签，保留每条路由。先 rate 再 sum，避免聚合掩盖单个实例的重置。
累计请求数不是 QPS；请求停止后累计值不下降，但 rate 会逐渐接近 0。

## 错误率

```promql
sum(rate(demo_requests_total{job="learning-app",http_response_status_code="500"}[1m]))
/
clamp_min(sum(rate(demo_requests_total{job="learning-app"}[1m])), 0.000001)
```

分子是每秒失败次数，分母是每秒总次数，结果是 0 到 1 的比例。
本实验流量每轮 3 个请求，其中 1 个失败；稳定流量下，错误率应接近 1/3。
clamp_min 避免分母为 0；如果错误序列从未产生，结果仍可能为空，空结果不等于 0。

## 平均耗时

```promql
sum by (http_route) (rate(demo_request_duration_seconds_sum{job="learning-app"}[1m]))
/
clamp_min(sum by (http_route) (rate(demo_request_duration_seconds_count{job="learning-app"}[1m])), 0.000001)
```

`_sum` 是累计耗时（秒），`_count` 是累计样本数。两者增长率相除，得到窗口内平均耗时。
`/slow` 的平均值应约 0.25 秒或稍长。

## P95

```promql
histogram_quantile(
  0.95,
  sum by (le, http_route) (
    rate(demo_request_duration_seconds_bucket{job="learning-app"}[1m])
  )
)
```

P95 表示约 95% 的请求耗时不超过这个值。`le` 是桶的上界，聚合时必须保留它。
这里使用经典 Histogram：桶内做插值，所以结果是估算值；慢请求刚超过 0.25s 时落在 0.25–0.5s 桶，P95 可能明显大于实测的约 0.25s。这是桶边界造成的精度差异。

## 练习

- 去掉 `by (http_route)`，看看多条路由混合后的平均值是否还能反映 `/slow` 的耗时。
- 把 P95 改成 P50，比较结果；解释二者和平均值的差异。
- 从流量脚本中暂时移除 `/error` 请求，等待一分钟，观察错误率变化。

公式参考：[Prometheus rate 和 histogram_quantile](https://prometheus.io/docs/prometheus/latest/querying/functions/)。
