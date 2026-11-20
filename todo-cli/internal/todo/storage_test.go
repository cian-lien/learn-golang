package todo

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestSaveAndLoadTasks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	want := []Task{
		{ID: 1, Title: "买牛奶"},
		{ID: 3, Title: "学习 Go", Done: true},
	}

	if err := SaveTasks(path, want); err != nil {
		t.Fatalf("保存任务失败: %v", err)
	}

	got, err := LoadTasks(path)
	if err != nil {
		t.Fatalf("读取任务失败: %v", err)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("读取结果 = %+v，期望 %+v", got, want)
	}
}

func TestLoadTasksMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")

	tasks, err := LoadTasks(path)
	if err != nil {
		t.Fatalf("文件不存在时不应报错: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("文件不存在时应返回空列表，实际有 %d 条任务", len(tasks))
	}
}

func TestLoadTasksRejectsMalformedJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	if err := os.WriteFile(path, []byte("["), 0o600); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	if _, err := LoadTasks(path); err == nil {
		t.Fatal("JSON 损坏时应返回错误")
	}
}
