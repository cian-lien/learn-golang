package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	// 整数类型长度明确；溢出在运行期回绕（不 panic）
	var i8 int8 = 127
	var u8 uint8 = 255
	i8++
	fmt.Println("int8 溢出回绕:", i8, "| uint8:", u8)

	// 类型转换必须显式写出来，没有隐式提升
	var i int = 42
	var f float64 = float64(i)
	var u uint = uint(f)
	fmt.Printf("int=%d float64=%v uint=%d\n", i, f, u)

	// string 是只读的字节序列，UTF-8 编码
	s := "Go语言"
	fmt.Println("字节长度:", len(s), "| 字符个数:", utf8.RuneCountInString(s))
	fmt.Printf("按字节的十六进制: % x\n", []byte(s))
	for idx, r := range s { // range 按 UTF-8 解码，idx 是字节下标
		fmt.Printf("  index=%d rune=%c code=%U\n", idx, r, r)
	}

	// 反引号是原始字符串，不做转义
	raw := `第一行\n第二行`
	fmt.Println("原始字符串:", raw)

	// []rune 按字符切分，string(...) 还原
	rs := []rune(s)
	fmt.Println("rune 切片:", rs, "| 还原:", string(rs))
	fmt.Println("拼接:", s[:2]+"-"+string(rs[2:]))
}
