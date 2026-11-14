package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// server 把 store 带给每个 handler，不再需要全局变量
type server struct {
	store *Store
}

type createTodoRequest struct {
	Title string `json:"title"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// ---------- 工具函数 ----------

func writeJSON(w http.ResponseWriter, status int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Println("JSON 编码失败：", err)
		http.Error(w, `{"error":"服务器内部错误"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}

// writeError 统一用 JSON 格式返回错误：{"error": "..."}
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

// ---------- handler ----------

func (s *server) listTodos(w http.ResponseWriter, r *http.Request) {
	var done *bool
	switch r.URL.Query().Get("done") {
	case "":
		// 不筛选，done 保持 nil
	case "true":
		v := true
		done = &v
	case "false":
		v := false
		done = &v
	default:
		writeError(w, http.StatusBadRequest, "参数 done 必须是 true 或 false")
		return
	}

	writeJSON(w, http.StatusOK, s.store.List(done))
}

func (s *server) createTodo(w http.ResponseWriter, r *http.Request) {
	var req createTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求体必须是合法的 JSON")
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		writeError(w, http.StatusBadRequest, "标题不能为空")
		return
	}
	if len([]rune(title)) > 50 {
		writeError(w, http.StatusBadRequest, "标题不能超过 50 个字")
		return
	}

	writeJSON(w, http.StatusCreated, s.store.Create(title))
}
