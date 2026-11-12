// 第 4 部分练习 2：及格成绩
// 运行：go run ./04-slices-maps/passed
package main

import "fmt"

// passed 返回及格（>= 60）的成绩和及格人数
func passed(scores []int) ([]int, int) {
	var result []int // nil 切片，可以直接 append
	for _, s := range scores {
		if s >= 60 {
			result = append(result, s) // 记得赋值回去
		}
	}
	return result, len(result)
}

// printFailed 打印所有不及格的人（顺序随机，因为 map 遍历是随机的）
func printFailed(scores map[string]int) {
	for name, score := range scores {
		if score < 60 {
			fmt.Println(name, "不及格，分数：", score)
		}
	}
}

func main() {
	list, count := passed([]int{55, 78, 90, 42, 60, 88})
	fmt.Println("及格的成绩：", list)
	fmt.Println("及格人数：", count)

	printFailed(map[string]int{
		"小明": 90,
		"小红": 55,
		"小刚": 78,
		"小李": 42,
	})
}
