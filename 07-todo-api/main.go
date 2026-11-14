package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	store := NewStore()
	store.Create("学习 Go 基础语法")
	store.Create("写一个 HTTP 接口")
	store.Create("学习 JSON 处理")
	store.UpdateDone(1, true)

	s := &server{store: store}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /todos", s.listTodos)
	mux.HandleFunc("POST /todos", s.createTodo)
	mux.HandleFunc("GET /todos/{id}", s.getTodo)
	mux.HandleFunc("PATCH /todos/{id}", s.updateTodo)
	mux.HandleFunc("DELETE /todos/{id}", s.deleteTodo)

	fmt.Println("服务已启动：http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
