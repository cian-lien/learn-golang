package main

import (
	"encoding/json"
	"fmt"
)

type Task struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

func main() {
	var task Task
	if err := json.Unmarshal([]byte(`{"title":"读书","done":false}`), &task); err != nil {
		panic(err)
	}
	task.Done = true
	data, err := json.Marshal(task)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}
