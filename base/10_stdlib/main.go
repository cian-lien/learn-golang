package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type User struct {
	Name      string    `json:"name"`
	Age       int       `json:"age"`
	Email     string    `json:"email,omitempty"` // 空值时省略
	CreatedAt time.Time `json:"created_at"`
}

var ErrEOF = errors.New("EOF")

func main() {
	// strings
	fmt.Println(strings.ToUpper("go"), strings.Contains("golang", "lang"), strings.Split("a,b,c", ","))
	fmt.Println(strings.TrimSpace("  x  "), strings.Repeat("=", 3), strings.ReplaceAll("a-b-a", "a", "A"))
	fmt.Println(strings.Join([]string{"go", "is", "fast"}, " "), strings.HasPrefix("golang", "go"))

	// strconv：字符串与数字互转
	n, err := strconv.Atoi("42")
	fmt.Println("Atoi:", n, err)
	_, err = strconv.Atoi("x")
	fmt.Println("Atoi 失败时 err != nil:", err != nil)
	fv, _ := strconv.ParseFloat("3.14", 64)
	fmt.Println("Itoa:", strconv.Itoa(7), "| ParseFloat:", fv, "| FormatFloat:", strconv.FormatFloat(1.0/3, 'f', 3, 64))
	fmt.Println("Quote:", strconv.Quote("a\nb"))

	// sort
	nums := []int{5, 1, 9, 3}
	sort.Ints(nums)
	fmt.Println("排序:", nums)
	people := []User{{Name: "b", Age: 30}, {Name: "a", Age: 20}}
	sort.Slice(people, func(i, j int) bool { return people[i].Age < people[j].Age })
	fmt.Println("按年龄排序后第一个:", people[0].Name, people[0].Age)

	// time：Go 的参考时间是 2006-01-02 15:04:05
	now := time.Date(2026, 9, 30, 10, 0, 0, 0, time.UTC)
	fmt.Println("格式化:", now.Format("2006-01-02 15:04:05"))
	fmt.Println("加 48 小时:", now.Add(48*time.Hour).Format(time.DateOnly), "| 时长字面量:", 2*time.Hour+30*time.Minute)
	parsed, _ := time.Parse(time.DateOnly, "2026-01-02")
	fmt.Println("解析:", parsed.Format("2026年01月02日"), "| Before:", parsed.Before(now))

	// json
	u := User{Name: "小明", Age: 18, CreatedAt: now}
	b, _ := json.Marshal(u)
	fmt.Println("Marshal:", string(b))
	b, _ = json.MarshalIndent(u, "", "  ")
	fmt.Println("MarshalIndent:\n" + string(b))

	var u2 User
	err = json.Unmarshal([]byte(`{"name":"小红","age":20,"extra":1}`), &u2)
	fmt.Println("Unmarshal:", u2.Name, u2.Age, "| 未知字段默认忽略, err =", err)

	// 错误包装：errors.Is 能穿透 %w 错误链
	err = fmt.Errorf("读取用户失败: %w", ErrEOF)
	fmt.Println("包装错误:", err, "| errors.Is(err, ErrEOF):", errors.Is(err, ErrEOF))
}
