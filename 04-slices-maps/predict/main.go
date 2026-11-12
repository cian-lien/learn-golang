// 第 4 部分练习 1：预测输出
// 运行：go run ./04-slices-maps/predict
// 输出：
// [10 20 0 40] 2
// 1 5 0 false 2
package main

import "fmt"

func main() {
	s := []int{10, 20, 30}
	s = append(s, 40) // [10 20 30 40]
	t := s[1:3]       // [20 30]，和 s 共享数据
	t[1] = 0          // t[1] 就是 s[2]
	fmt.Println(s, len(t))

	m := map[string]int{"a": 1}
	m["b"] += 5     // "b" 不存在，读到零值 0，0+5=5 写回
	v, ok := m["c"] // "c" 不存在：v=0，ok=false；只读不会新增键
	fmt.Println(m["a"], m["b"], v, ok, len(m))
}
