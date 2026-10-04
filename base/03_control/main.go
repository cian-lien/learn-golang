package main

import "fmt"

func main() {
	// if 支持初始化语句，作用域只在这个 if/else 内
	if v := 10; v > 5 {
		fmt.Println("if 初始化语句:", v)
	}

	// Go 只有 for 一个循环关键字
	sum := 0
	for i := 0; i < 5; i++ { // 经典三段式
		sum += i
	}
	fmt.Println("经典 for 求和:", sum)

	n := 0
	for n < 3 { // 相当于 while
		n++
	}
	fmt.Println("while 风格:", n)

	c := 0
	for { // 无限循环
		c++
		if c == 3 {
			break
		}
	}
	fmt.Println("无限循环 + break:", c)

	for i := range 3 { // Go 1.22+：range over int
		fmt.Print(i, " ")
	}
	fmt.Println()

	for idx, v := range []string{"a", "b"} { // range 遍历切片/map/字符串/channel
		fmt.Printf("range %d=%s ", idx, v)
	}
	fmt.Println()

	// switch：默认不穿透；可以带初始化语句；也可以没有表达式
	switch x := 7; {
	case x < 5:
		fmt.Println("switch 无表达式: 小")
	case x < 10:
		fmt.Println("switch 无表达式: 中")
	default:
		fmt.Println("switch 无表达式: 大")
	}

	switch 3 {
	case 1, 2: // 一个 case 可以匹配多个值
		fmt.Println("命中 1 或 2")
	case 3:
		fmt.Println("命中 3，接着 fallthrough")
		fallthrough
	case 4:
		fmt.Println("fallthrough 落到这里")
	}

	// 标签 + continue/break 跳出多层循环
outer:
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if j == 1 {
				continue outer
			}
			fmt.Print(i, j, " ")
		}
	}
	fmt.Println()

	// goto：少用，但用在内层循环统一清理时偶尔可见
	i := 0
loop:
	if i < 2 {
		i++
		goto loop
	}
	fmt.Println("goto 结束:", i)
}
