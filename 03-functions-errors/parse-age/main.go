// 第 3 部分练习 2：parseAge
// 运行：go run ./03-functions-errors/parse-age
// 输出：
// 年龄： 25
// 出错： strconv.Atoi: parsing "abc": invalid syntax
// 出错： 年龄不合理：200
package main

import (
	"fmt"
	"strconv"
)

func parseAge(ageStr string) (int, error) {
	age, err := strconv.Atoi(ageStr)
	if err != nil {
		return 0, err // 转换失败，直接把错误往上传
	}

	if age < 0 || age > 150 {
		return 0, fmt.Errorf("年龄不合理：%d", age) // 错误信息里带上出错的值
	}

	return age, nil
}

// check 封装重复的"调用 + 判断错误 + 打印"
func check(s string) {
	age, err := parseAge(s)
	if err != nil {
		fmt.Println("出错：", err)
		return // 先处理错误并返回，正常逻辑放在最后
	}
	fmt.Println("年龄：", age)
}

func main() {
	check("25")
	check("abc")
	check("200")
}
