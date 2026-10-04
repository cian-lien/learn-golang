package main

import (
	"fmt"
	"sort"
)

func main() {
	// 数组：长度是类型的一部分，赋值会整体复制
	var arr [3]int
	arr2 := [3]int{1, 2, 3}
	arr3 := [...]int{4, 5} // 由元素个数推导长度
	arrCopy := arr2
	arrCopy[0] = 99
	fmt.Println("数组:", arr, arr2, arr3, "| 复制不影响原数组:", arr2[0])

	// 切片：make([]T, len, cap)
	s := make([]int, 0, 4)
	fmt.Println("初始 len/cap:", len(s), cap(s))
	for i := 0; i < 5; i++ {
		s = append(s, i) // append 返回新切片，必须接住
		fmt.Printf("  append %d → len=%d cap=%d\n", i, len(s), cap(s))
	}
	fmt.Println("切片:", s, "| s[1:3]:", s[1:3], "| s[:2]:", s[:2], "| s[3:]:", s[3:])

	// 易错点 1：切片共享底层数组
	base := []int{1, 2, 3, 4}
	view := base[1:3]
	view[0] = 99
	fmt.Println("共享底层数组 → base:", base, "view:", view)

	safe := make([]int, len(view))
	copy(safe, view) // 想独立就 copy
	safe[0] = 0
	fmt.Println("copy 之后互不影响 → base:", base, "safe:", safe)

	// 易错点 2：切片按值传递，改元素影响调用方，append 不一定
	grow := func(in []int) {
		in[0] = -1           // 同一底层数组 → 调用方可见
		in = append(in, 100) // 需要扩容时换新数组 → 调用方看不到
	}
	xs := []int{1, 2, 3} // len=cap=3，append 必然扩容
	grow(xs)
	fmt.Println("切片传参 → xs:", xs)

	// 二维切片
	grid := [][]int{{1, 2}, {3, 4}}
	fmt.Println("二维切片:", grid, "| grid[1][0]:", grid[1][0])

	// map：零值是 nil，读取安全，写入 panic
	var m map[string]int
	fmt.Println("nil map 读取:", m["x"], "| len:", len(m), "| 是否 nil:", m == nil)
	m = map[string]int{"a": 1}
	m["b"] = 2
	m["a"]++
	fmt.Println("map:", m, "| len:", len(m))

	// comma-ok：判断 key 是否存在
	v, ok := m["c"]
	fmt.Println("comma-ok →", v, ok)
	delete(m, "a")
	fmt.Println("delete 之后:", m)

	// 遍历顺序随机，需要顺序就自己排序
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Println("排序后的 key:", keys)
}
