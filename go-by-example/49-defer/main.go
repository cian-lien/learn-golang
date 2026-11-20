package main

import (
	"fmt"
)

func main() {
	defer fmt.Println("关闭仓库")
	defer fmt.Println("收回钥匙")
	fmt.Println("完成盘点")
}
