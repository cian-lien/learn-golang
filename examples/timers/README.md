# Timers

Timer 是一次性的定时通知：创建时开始计时，到期后，通过 channel 发送一个时间值。

参考：[Go by Example](https://gobyexample.com/timers)、[NewTimer](https://pkg.go.dev/time#NewTimer)、[Timer.Stop](https://pkg.go.dev/time#Timer.Stop)。

## 运行示例

在项目根目录执行：

```bash
go run ./examples/timers
```

通常约 2 秒后输出，编译时间另计：

```text
Timer 1 fired
Timer 2 stopped
```

完整代码见 [main.go](./main.go)，分别演示等待通知和停止定时器，并自检第二个定时器是否成功停止。

## 创建和等待是两个动作

下面的代码放在 `main` 函数中，使用 `fmt` 和 `time` 包：

```go
timer := time.NewTimer(2 * time.Second)
<-timer.C
fmt.Println("Timer fired")
```

- `time.NewTimer(d)` 立即返回 `*time.Timer`，从创建时开始计时。
- `timer.C` 的类型是 `<-chan time.Time`，即只接收 channel。
- `<-timer.C` 等待通知；这里丢弃收到的时间值，只关心“时间到了”。

如果创建定时器后先做了其他事情，执行接收操作时等待的是剩余时间。若通知已经就绪，就不必重新等待完整的时长。

定时器提供通知，打印等业务逻辑仍由接收通知的代码执行。

## 和 Timeouts 的联系

之前 [Timeouts 示例](../timeouts.go) 使用的 `time.After(d)`，等价于 `time.NewTimer(d).C`。

`time.After` 直接提供 channel；`NewTimer` 还提供定时器对象，可以调用 `Stop()`。如果只是单纯暂停当前 goroutine，也可以使用 `time.Sleep(d)`。

## 停止定时器

```go
timer := time.NewTimer(time.Second)
stopped := timer.Stop()
fmt.Println("stopped:", stopped)
```

- 返回 `true`：这次调用成功停止了定时器。
- 返回 `false`：没有停止一个活动的定时器，例如它此前已经停止或已经触发。

返回 `false` 不是一个 `error`，也不能据此判断等待它的 goroutine 是否已经退出。

本节讨论 `NewTimer` 创建的 channel 定时器。Go 1.23+ 的默认语义下，如果通知尚未被接收且定时器此前未停止，`Stop()` 保证返回 `true`；不能仅凭经过的时长判断返回值。旧版本定时器 channel 的细节不同，参见 [Go 1.23 Timer Channel Changes](https://go.dev/wiki/Go123Timer)。

## Stop 不会结束等待它的 goroutine

原文第二段启动了下面这个 goroutine。它是需要注意的等待写法，不是完整的取消方案：

```go
go func() {
	<-timer2.C
	fmt.Println("Timer 2 fired")
}()
```

如果随后成功执行 `timer2.Stop()`，这个 goroutine 收不到通知，就会继续等待。

原文最后通过 `main` 返回结束整个程序；在持续运行的服务中，需要另外提供退出信号，并让等待者监听它。`Stop()` 不会替等待者完成退出。

原文末尾的 `time.Sleep` 只用于观察第二个定时器是否触发，不是停止定时器所需的步骤。本目录的可运行示例直接演示停止操作。

## 和上一篇 Range over Channels 的联系

[上一篇](../range-over-channels/README.md) 的 `range` 依靠“channel 已关闭且排空”自然结束。

Timer 触发后，不会自动关闭 `timer.C`；停止 Timer 也不会关闭它。因此，下面是一个会继续等待的反例：

```go
timer := time.NewTimer(time.Second)
for firedAt := range timer.C {
	fmt.Println(firedAt)
}
```

1. 定时器到期，循环接收一个值，执行循环体。
2. 循环再次尝试接收。
3. 代码没有重新安排触发，channel 也没有关闭，于是继续等待。

| 操作 | 后续接收的行为 |
| --- | --- |
| `close(queue)`，并收完剩余数据 | 接收返回 `ok == false`，`range` 自然结束 |
| Timer 触发，并接收一次通知 | 后续接收继续等待 |
| 成功停止 Timer，且未重新启动 | 不再产生这次通知，接收仍会等待 |

等待一次定时通知，直接使用 `<-timer.C` 即可。

## 易错点

- 计时从 `NewTimer` 调用时开始，不是从执行 `<-timer.C` 时开始。
- Timer 触发一次，不等于 channel 关闭；不能据此期待 `range` 自动结束。
- `Stop()` 停止定时器，不会关闭 channel，也不会终止其他 goroutine。
- `timer.C` 是只接收 channel，不能像自己创建的发送端那样执行 `close(timer.C)`。

## 自测

1. 创建一个 3 秒定时器，先做了约 2 秒的工作，再执行 `<-timer.C`，通常还要等多久？
2. `timer.Stop()` 成功后，正在等待 `<-timer.C` 的 goroutine 会自动退出吗？
3. 为什么 `range queue` 可以在关闭且排空后自然结束，而上面的 `range timer.C` 不会？
