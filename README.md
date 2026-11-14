# Go 语言语法入门

一套面向前端开发者的 Go 基础语法讲义，共 6 个部分。每部分都有讲解、代码例子（大量和 JavaScript / TypeScript 对比）、练习和答案。

## 学习路线

| 部分 | 内容 | 核心要点 |
|---|---|---|
| [01 程序结构、变量、常量、基本类型](01-basics/README.md) | `package`、`:=`、零值、`iota`、类型转换、`fmt` | `:=` 最常用；变量有零值；类型不会自动转换 |
| [02 流程控制](02-control-flow/README.md) | `if`、`for`、`switch`、`range` | 不加括号；只有 `for` 一种循环；`switch` 不用 `break` |
| [03 函数、多返回值、错误处理](03-functions-errors/README.md) | 函数、多返回值、`error`、`defer`、闭包 | 错误作为最后一个返回值，用 `if err != nil` 检查 |
| [04 数组、切片、map](04-slices-maps/README.md) | 数组、切片、`append`、截取、map | `append` 要赋值回去；截取共享数据；`v, ok :=` |
| [05 结构体、方法、接口](05-structs-interfaces/README.md) | 结构体、方法、嵌入、接口、`any` | 结构体赋值会复制；改字段用指针接收者；不用写 implements |
| [06 指针和并发](06-pointers-concurrency/README.md) | 指针、goroutine、`WaitGroup`、channel、`select`、`Mutex` | 传地址才能改原值；main 结束程序就退出；共享变量要加锁 |

建议按顺序学习。每部分先读讲义，做完练习再看答案。

## 如何运行代码

需要安装 Go 1.22 或更高版本。没有安装的话，可以把代码粘贴到官方在线环境 [Go Playground](https://go.dev/play/) 运行。

每个练习的答案都放在单独的文件夹里，可以直接运行：

```bash
go run ./01-basics/exercise
go run ./02-control-flow/fizzbuzz
go run ./06-pointers-concurrency/squares
```

## 目录结构

```
01-basics/                  第 1 部分
  README.md                 讲义 + 练习 + 答案
  exercise/main.go          练习代码
02-control-flow/
  range-switch/main.go      练习 1：预测输出
  fizzbuzz/main.go          练习 2：FizzBuzz
03-functions-errors/
  defer-calc/main.go        练习 1：defer 与错误
  parse-age/main.go         练习 2：parseAge
04-slices-maps/
  predict/main.go           练习 1：预测输出
  passed/main.go            练习 2：及格成绩
05-structs-interfaces/
  counter/main.go           练习 1：值接收者与指针接收者
  shapes/main.go            练习 2：Shape 接口
06-pointers-concurrency/
  pointer/main.go           练习 1：指针传参
  squares/main.go           练习 2：并发计算平方和
07-todo-api/                综合练习：待办事项 HTTP API
```

## 07 待办事项 API

[07-todo-api](07-todo-api/) 使用标准库 `net/http` 实现待办事项的增删查改。服务启动时预置 3 条待办，数据保存在内存中，重启后恢复为预置内容。

### 启动服务

使用 Go 1.22 或更高版本，在仓库根目录运行以下命令。服务监听 `8080` 端口，按 `Ctrl+C` 停止。

```bash
go run ./07-todo-api
```

### 接口示例

另开一个终端，按顺序执行下方命令。服务刚启动时，第一条新建待办的 `id` 为 `4`。如果服务已经处理过创建请求，把后续命令中的 `4` 替换为创建响应中的 `id`。

```bash
# 创建待办（201 Created）
curl -i http://localhost:8080/todos \
  -H 'Content-Type: application/json' \
  -d '{"title":"练习 Go HTTP 接口"}'
```

创建成功时返回待办对象。

```json
{"id":4,"title":"练习 Go HTTP 接口","done":false}
```

```bash
# 查看全部待办（200 OK）
curl -i http://localhost:8080/todos

# 查询指定待办（200 OK）
curl -i http://localhost:8080/todos/4

# 标记为已完成（200 OK）
curl -i -X PATCH http://localhost:8080/todos/4 \
  -H 'Content-Type: application/json' \
  -d '{"done":true}'

# 筛选已完成的待办（200 OK）；改用 done=false 可查看未完成的待办
curl -i 'http://localhost:8080/todos?done=true'

# 删除待办（204 No Content）
curl -i -X DELETE http://localhost:8080/todos/4

# 再次查询已删除的待办（404 Not Found）
curl -i http://localhost:8080/todos/4
```
