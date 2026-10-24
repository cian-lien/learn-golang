# Range over Channels

`for value := range ch` 会持续接收 channel 中的值，直到 channel 已关闭，而且所有数据都已接收完。

参考：[Go by Example](https://gobyexample.com/range-over-channels)、[Go 语言规范](https://go.dev/ref/spec#For_range)。

## 运行示例

在项目根目录执行：

```bash
go run ./examples/range-over-channels
```

输出：

```text
one
two
```

完整代码见 [main.go](./main.go)。核心顺序是：

```go
queue := make(chan string, 2)
queue <- "one"
queue <- "two"
close(queue)

for elem := range queue {
	fmt.Println(elem)
}
```

1. 缓冲区存入 `"one"`、`"two"`。
2. `close(queue)` 表示以后不再发送，已有数据仍然保留。
3. `range` 依次接收两个值，分别执行循环体。
4. channel 已关闭且排空，循环结束。

这里先发送、再接收，因此缓冲容量为 `2`。如果只改成 `1`，第二次发送就会阻塞，执行不到后面的 `close` 和 `range`。

## 和上一篇 Closing Channels 的联系

上一篇手动检查接收结果：

```go
for {
	value, ok := <-ch
	if !ok {
		break
	}
	fmt.Println(value)
}
```

这一篇可以把相同的接收逻辑写成：

```go
for value := range ch {
	fmt.Println(value)
}
```

`range` 帮你处理“接收 → 判断是否收完 → 继续或退出”。它不会自动关闭 channel。

`ok == false` 表示 channel 已关闭且排空。因此，关闭后缓冲区中尚未接收的数据仍然会进入循环体。

## 暂时为空和已经结束的区别

| channel 状态 | range 的行为 |
| --- | --- |
| 当前能够接收到数据 | 接收一个值，执行一次循环体 |
| 未关闭，暂时没有数据 | 等待后续数据 |
| 已关闭，缓冲区还有数据 | 继续接收剩余数据 |
| 已关闭且排空 | 自动结束循环 |

与遍历空切片不同，channel 暂时为空时，后续仍可能有人发送数据，所以不能直接结束。

## 无缓冲 channel 的并发写法

不需要先关闭 channel 才开始遍历。下面的代码放在 `main` 函数中，使用 `fmt` 包：

```go
queue := make(chan string)

go func() {
	queue <- "one"
	queue <- "two"
	close(queue)
}()

for elem := range queue {
	fmt.Println(elem)
}
```

发送和接收在不同 goroutine 中，可以配合完成无缓冲通信。发送方发完后关闭 channel，接收循环自然结束。

## 简化上一篇的 worker

在 [Closing Channels 示例](../closing-channels/main.go) 中，可以将 worker 的 goroutine 改成：

```go
go func() {
	for job := range jobs {
		fmt.Println("received job", job)
	}

	fmt.Println("received all jobs")
	done <- true
}()
```

- `close(jobs)`：发送方宣布任务发完。
- `range jobs`：worker 接收并处理任务，收完后结束循环。
- `done <- true`：通知主流程，任务已经处理完毕。

关闭通知和处理完成通知的职责仍然不同；关闭 `jobs` 本身不会等待 worker。

## 易错点

- `for value := range ch` 中的 `value` 是收到的值，不是下标；channel 遍历最多使用一个迭代变量。
- 真正发送的 `0` 或 `""` 都是正常数据，会使循环体执行一次。关闭且排空后则直接结束，不会额外执行一次循环体。
- `range` 不会自动关闭 channel，通常由发送方负责关闭。
- 删除关闭操作后，读完现有数据仍会等待。在原文这个只有 main goroutine 的程序中，会因无人继续发送而死锁。
- `break` 只结束当前循环，不会关闭 channel，也不会通知发送方停止。发送方之后继续发送时仍可能阻塞。

## 自测

1. 删除示例中的 `close(queue)`，打印完两个值后会怎样？
2. 把缓冲容量从 `2` 改成 `1`，为什么程序执行不到 `range`？
3. 向有缓冲的 `chan int` 发送 `0` 后关闭，`range` 会执行几次循环体？
