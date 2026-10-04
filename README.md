# Go 学习工作区

## 目录结构

```text
go/
├── base/                   Go 语言基础：11 个主题的可运行示例 + 复习正文
│   ├── Go语言基础复习.md     复习正文（命令都在 go/ 下执行）
│   ├── 01_vars … 10_stdlib  可运行示例（每个目录一个 main 包）
│   └── 11_testing           表驱动测试与基准测试（库包）
├── 12_observability/       OpenTelemetry 示例：net/http 两个服务 + Collector/Jaeger
├── gin-otel-demo/          Gin + OpenTelemetry（独立 module，需要 cd 进去跑）
├── go.mod / go.sum         模块 gobasics（base/ 与 12_observability 属于它）
└── README.md
```

## 基础部分（`base/`）

复习正文在 [base/Go语言基础复习.md](./base/Go语言基础复习.md)，示例代码按主题分目录。
所有命令都在本目录（模块根）执行：

```bash
go run ./base/01_vars           # 变量、常量与零值
go run ./base/02_types          # 类型转换、字符串与 rune
go run ./base/03_control        # if / for / switch / 标签 / goto
go run ./base/04_funcs          # 多返回值、变参、闭包、defer
go run ./base/05_slice_map      # 数组、切片、map 与共享底层数组
go run ./base/06_struct_method  # 结构体、值/指针接收者、嵌入
go run ./base/07_interface_error # 接口、类型断言、错误链
go run ./base/08_concurrency    # goroutine、channel、select、context
go run ./base/09_generics       # 类型参数与约束
go run ./base/10_stdlib         # strings / strconv / sort / time / json
go test ./base/11_testing -v    # 表驱动测试与基准测试
```

## 可观测性部分

| 目录 | 内容 | 怎么跑 |
| --- | --- | --- |
| [12_observability](./12_observability/README.md) | 不依赖 Web 框架：`net/http` 两个服务，trace / metric / log 三件套 | `go run ./12_observability -role=user`，另开终端跑 `-role=order` |
| [gin-otel-demo](./gin-otel-demo/README.md) | Gin 版：handler / service / store 三层埋点，含单元测试与 Collector 配置 | `cd gin-otel-demo && go run .`（独立 module） |

两者都能零依赖起步：加 `OTEL_EXPORTER=console` 就把 span/metric 打到标准输出；
需要看链路 UI 时，用各自目录里的 `docker compose up -d` 起 Jaeger（<http://localhost:16686>）+ Collector（:4317）。

## 常用命令

```bash
go vet ./...            # 静态检查
gofmt -l .              # 格式检查
go test ./... -race     # 全量测试 + 数据竞争检测
```

模块名 `gobasics`，`go.mod` 声明 `go 1.25`（`12_observability` 依赖的 OpenTelemetry SDK 要求 1.25+，
`base/` 下的基础示例本身在更低的 Go 版本上也能跑）。
`gin-otel-demo/` 是独立 module，不参与上面的 `./...`，需要 `cd` 进去单独 build / test。
