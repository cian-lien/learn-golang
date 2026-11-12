// 第 3 部分练习 1：defer 与错误处理
// 运行：go run ./03-functions-errors/defer-calc
// 输出：
// 4 <nil>
// 错误： 除数不能为0
// 结束
package main

import (
	"errors"
	"fmt"
)

func calc(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("除数不能为0")
	}
	return a / b, nil
}

func main() {
	defer fmt.Println("结束") // 函数结束时才执行

	r, err := calc(9, 2) // 整数除法：4
	fmt.Println(r, err)  // 没有错误时 err 是 nil

	r, err = calc(1, 0) // 已声明过，用 = 而不是 :=
	if err != nil {
		fmt.Println("错误：", err)
		return // 提前结束，触发 defer
	}
	fmt.Println("这行会打印吗？") // 不会执行
}
