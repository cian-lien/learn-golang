package todo

import (
	"slices"
	"testing"
)

func TestAddTaskAssignsIDAfterMaximum(t *testing.T) {
	tasks := []Task{
		{ID: 3, Title: "学习 Go", Done: true},
		{ID: 1, Title: "买牛奶"},
	}
	want := []Task{tasks[0], tasks[1], {ID: 4, Title: "写测试"}}

	if got := AddTask(tasks, "写测试"); !slices.Equal(got, want) {
		t.Fatalf("添加后的任务 = %+v，期望 %+v", got, want)
	}
}

func TestMarkDoneChangesOnlyMatchingTask(t *testing.T) {
	tasks := []Task{{ID: 1}, {ID: 3}}
	if !MarkDone(tasks, 3) {
		t.Fatal("应找到任务 3")
	}
	want := []Task{{ID: 1}, {ID: 3, Done: true}}
	if !slices.Equal(tasks, want) {
		t.Fatalf("标记后的任务 = %+v，期望 %+v", tasks, want)
	}
	if MarkDone(tasks, 99) || !slices.Equal(tasks, want) {
		t.Fatalf("找不到任务 99 时应返回 false 并保留任务，实际任务 = %+v", tasks)
	}
}

func TestDeleteTaskPreservesRemainingTasks(t *testing.T) {
	tasks := []Task{{ID: 1, Title: "买牛奶"}, {ID: 2}, {ID: 3, Done: true}}
	want := []Task{tasks[0], tasks[2]}

	got, found := DeleteTask(tasks, 2)
	if !found || !slices.Equal(got, want) {
		t.Fatalf("删除任务 2 后得到 %+v，found = %t，期望 %+v 和 true", got, found, want)
	}

	got, found = DeleteTask(want, 99)
	if found || !slices.Equal(got, want) {
		t.Fatalf("删除不存在的任务后得到 %+v，found = %t，期望 %+v 和 false", got, found, want)
	}
}

func TestDeleteLastTaskReturnsEmptyList(t *testing.T) {
	got, found := DeleteTask([]Task{{ID: 1}}, 1)
	if !found || got == nil || len(got) != 0 {
		t.Fatalf("删除最后一条任务后应返回非 nil 的空列表和 true，实际 = %#v, %t", got, found)
	}
}
