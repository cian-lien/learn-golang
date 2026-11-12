// 第 2 部分练习 2：FizzBuzz
// 运行：go run ./02-control-flow/fizzbuzz
package main

import "fmt"

func main() {
	for i := 1; i <= 15; i++ {
		switch { // 不带表达式的 switch：找第一个为 true 的 case
		case i%3 == 0 && i%5 == 0: // 必须放在第一个
			fmt.Println("FizzBuzz")
		case i%3 == 0:
			fmt.Println("Fizz")
		case i%5 == 0:
			fmt.Println("Buzz")
		default:
			fmt.Println(i)
		}
	}
}
