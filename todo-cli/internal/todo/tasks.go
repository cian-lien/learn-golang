package todo

func AddTask(tasks []Task, title string) []Task {
	maxID := 0
	for _, t := range tasks {
		if t.ID > maxID {
			maxID = t.ID
		}
	}

	// 创建任务
	newTask := Task{
		ID:    maxID + 1,
		Title: title,
		Done:  false,
	}

	return append(tasks, newTask)

}

func MarkDone(tasks []Task, taskID int) bool {
	for i, task := range tasks {
		if task.ID == taskID {
			tasks[i].Done = true
			return true
		}
	}
	return false
}

func DeleteTask(tasks []Task, taskID int) ([]Task, bool) {
	found := false
	newTasks := []Task{}
	for _, task := range tasks {
		if task.ID == taskID {
			found = true
			continue
		}
		newTasks = append(newTasks, task)
	}
	return newTasks, found
}
