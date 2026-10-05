# 第 01 课：先知道我们要观察什么

目标：能解释一次 HTTP 请求的状态码和耗时。本课只运行 `cmd/01-http/main.go`，约 40 行标准库代码。

## 运行

在项目根目录执行：

```bash
go run ./cmd/01-http
```

另开终端，逐条执行：

```bash
curl -i http://127.0.0.1:18090/hello
curl -s -o /dev/null -w 'status=%{http_code} duration=%{time_total}s\n' http://127.0.0.1:18090/slow
curl -i http://127.0.0.1:18090/error
```

预期：`/hello` 返回 200；`/slow` 返回 200，耗时约 0.25 秒或稍长；`/error` 返回 500。
`curl` 的耗时包含客户端和传输开销，不完全等于服务端处理时间。

## 读四处代码

打开 `cmd/01-http/main.go`：

1. `http.NewServeMux()`：创建把请求分发给不同 handler 的路由器。
2. `HandleFunc("GET /hello", ...)`：为 `GET /hello` 注册处理函数。
3. `http.ResponseWriter` 用来写响应，`*http.Request` 表示收到的请求。
4. `ListenAndServe()`：开始监听，持续接收请求，所以终端会保持运行。

`/slow` 的延迟来自 `time.Sleep(250*time.Millisecond)`，`/error` 的 500 是故意返回的。
它们是实验条件，不代表真正的数据库或依赖服务。

## 思考与练习

- 把延迟改成 500ms，重启服务，再观察 curl 的结果。
- 请求 `/missing` 会发生什么？为什么？
- 如果一天收到 10 万个请求，如何知道其中有多少失败、95% 的请求在多久内完成？只看一次 curl 能回答吗？

完成后告诉我三个请求的状态码，以及 `/slow` 的耗时。按 Ctrl-C 停止应用，下一课再启动 OTel 版本。
