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

	s := &server{store: store}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /todos", s.listTodos)
	mux.HandleFunc("POST /todos", s.createTodo)

	fmt.Println("服务已启动：http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
