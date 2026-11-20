package main

import (
	"fmt"
	"github/cian-lien/todo-cli/internal/todo"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{Use: "todo"}

	var showAll bool

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "列出任务",
		RunE: func(_ *cobra.Command, _ []string) error {
			homeDir, err := os.UserHomeDir()

			if err != nil {
				return fmt.Errorf("获取用户目录失败：%w", err)
			}

			tasks, err := todo.LoadTasks(filepath.Join(homeDir, ".todo.json"))

			if err != nil {
				return fmt.Errorf("加载任务失败：%w", err)
			}

			writer := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(writer, "ID\tSTATUS\t任务")

			for _, task := range tasks {
				if task.Done && !showAll {
					continue
				}

				status := "[ ]"
				if task.Done {
					status = "[x]"
				}
				fmt.Fprintf(writer, "%d\t%s\t%s\n", task.ID, status, task.Title)
			}
			return writer.Flush()

		},
	}

	addCmd := &cobra.Command{
		Use:   "add [任务描述]",
		Short: "添加任务",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			homeDir, err := os.UserHomeDir()

			if err != nil {
				return fmt.Errorf("获取用户目录失败：%w", err)
			}
			// 将任务描述连接成一个字符串
			task := strings.TrimSpace(strings.Join(args, " "))

			if task == "" {
				return fmt.Errorf("任务描述不能为空")
			}

			tasks, err := todo.LoadTasks(filepath.Join(homeDir, ".todo.json"))
			if err != nil {
				return fmt.Errorf("加载任务失败：%w", err)
			}

			updatedTasks := todo.AddTask(tasks, task)

			if err := todo.SaveTasks(filepath.Join(homeDir, ".todo.json"), updatedTasks); err != nil {
				return fmt.Errorf("保存任务失败：%w", err)
			}

			fmt.Println("已保存任务：", task)
			return nil
		},
	}

	doneCmd := &cobra.Command{
		Use:   "done [任务ID]",
		Short: "标记任务为已完成",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			homeDir, err := os.UserHomeDir()

			if err != nil {
				return fmt.Errorf("获取用户目录失败：%w", err)
			}

			tasksPatch := filepath.Join(homeDir, ".todo.json")

			// 将任务ID从字符串转换为整数
			taskID, err := strconv.Atoi(args[0])
			if err != nil || taskID <= 0 {
				return fmt.Errorf("任务ID必须是正整数")
			}

			tasks, err := todo.LoadTasks(tasksPatch)
			if err != nil {
				return fmt.Errorf("加载任务失败：%w", err)
			}

			if !todo.MarkDone(tasks, taskID) {
				return fmt.Errorf("找不到任务: %d", taskID)
			}

			if err := todo.SaveTasks(tasksPatch, tasks); err != nil {
				return fmt.Errorf("保存任务失败：%w", err)
			}

			fmt.Printf("已标记任务 %d 为已完成\n", taskID)
			return nil
		},
	}

	deleteCmd := &cobra.Command{
		Use:   "delete [任务ID]",
		Short: "删除任务",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			homeDir, err := os.UserHomeDir()

			if err != nil {
				return fmt.Errorf("获取用户目录失败：%w", err)
			}

			tasksPatch := filepath.Join(homeDir, ".todo.json")

			taskID, err := strconv.Atoi(args[0])
			if err != nil || taskID <= 0 {
				return fmt.Errorf("任务ID必须是正整数")
			}

			tasks, err := todo.LoadTasks(tasksPatch)
			if err != nil {
				return fmt.Errorf("加载任务失败：%w", err)
			}

			newTasks, found := todo.DeleteTask(tasks, taskID)

			if !found {
				return fmt.Errorf("找不到任务: %d", taskID)
			}

			if err := todo.SaveTasks(tasksPatch, newTasks); err != nil {
				return fmt.Errorf("保存任务失败：%w", err)
			}

			fmt.Printf("已删除任务 %d\n", taskID)
			return nil
		},
	}

	listCmd.Flags().BoolVar(&showAll, "all", false, "显示所有任务")
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(doneCmd)
	rootCmd.AddCommand(deleteCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
