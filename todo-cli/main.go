package main

import (
	"flag"
	"fmt"
	"github/cian-lien/todo-cli/internal/todo"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "用法: todo <add|list> [参数]")
		os.Exit(1)
	}

	homeDir, err := os.UserHomeDir()

	if err != nil {
		fmt.Fprintln(os.Stderr, "获取用户目录失败：", err)
		os.Exit(1)
	}

	tasksPatch := filepath.Join(homeDir, ".todo.json")

	switch os.Args[1] {
	case "add":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "用法: todo add <任务描述>")
			os.Exit(1)
		}
		// 将任务描述连接成一个字符串
		task := strings.TrimSpace(strings.Join(os.Args[2:], " "))

		if task == "" {
			fmt.Fprintln(os.Stderr, "任务描述不能为空")
			os.Exit(1)
		}

		tasks, err := todo.LoadTasks(tasksPatch)
		if err != nil {
			fmt.Fprintln(os.Stderr, "加载任务失败:", err)
			os.Exit(1)
		}

		updatedTasks := todo.AddTask(tasks, task)

		if err := todo.SaveTasks(tasksPatch, updatedTasks); err != nil {
			fmt.Fprintln(os.Stderr, "保存任务失败", err)
			os.Exit(1)
		}

		fmt.Println("已保存任务：", task)

	case "list":
		listFlags := flag.NewFlagSet("list", flag.ContinueOnError)
		all := listFlags.Bool("all", false, "显示所有任务")

		if err := listFlags.Parse(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "解析参数失败:", err)
			os.Exit(1)
		}

		tasks, err := todo.LoadTasks(tasksPatch)
		if err != nil {
			fmt.Fprintln(os.Stderr, "加载任务失败:", err)
			os.Exit(1)
		}

		for _, task := range tasks {
			if task.Done && !*all {
				continue
			}
			status := "未完成"
			if task.Done {
				status = "已完成"
			}
			fmt.Printf("%d. %s [%s]\n", task.ID, task.Title, status)
		}

	case "done":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "用法: todo done <任务ID>")
			os.Exit(1)
		}

		taskID, err := strconv.Atoi(os.Args[2])
		if err != nil || taskID <= 0 {
			fmt.Fprintln(os.Stderr, "任务ID必须是正整数")
			os.Exit(1)
		}

		tasks, err := todo.LoadTasks(tasksPatch)
		if err != nil {
			fmt.Fprintln(os.Stderr, "加载任务失败:", err)
			os.Exit(1)
		}

		if !todo.MarkDone(tasks, taskID) {
			fmt.Fprintf(os.Stderr, "找不到任务: %d\n", taskID)
			os.Exit(1)
		}

		if err := todo.SaveTasks(tasksPatch, tasks); err != nil {
			fmt.Fprintln(os.Stderr, "保存任务失败", err)
			os.Exit(1)
		}

		fmt.Printf("已标记任务 %d 为已完成\n", taskID)
	case "delete":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "用法: todo delete <任务ID>")
			os.Exit(1)
		}

		taskID, err := strconv.Atoi(os.Args[2])
		if err != nil || taskID <= 0 {
			fmt.Fprintln(os.Stderr, "任务ID必须是正整数")
			os.Exit(1)
		}

		tasks, err := todo.LoadTasks(tasksPatch)
		if err != nil {
			fmt.Fprintln(os.Stderr, "加载任务失败:", err)
			os.Exit(1)
		}

		newTasks, found := todo.DeleteTask(tasks, taskID)

		if !found {
			fmt.Fprintf(os.Stderr, "找不到任务: %d\n", taskID)
			os.Exit(1)
		}

		if err := todo.SaveTasks(tasksPatch, newTasks); err != nil {
			fmt.Fprintln(os.Stderr, "保存任务失败", err)
			os.Exit(1)
		}

		fmt.Printf("已删除任务 %d\n", taskID)
	default:
		fmt.Fprintf(os.Stderr, "未知命令: %s\n", os.Args[1])
		os.Exit(1)
	}
}
