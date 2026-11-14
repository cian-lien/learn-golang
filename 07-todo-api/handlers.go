package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// server 把 store 带给每个 handler，不再需要全局变量
type server struct {
	store *Store
}

type createTodoRequest struct {
	Title string `json:"title"`
}

type updateTodoRequest struct {
	Done *bool `json:"done"`
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

func parseID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id 必须是整数")
		return 0, false
	}
	return id, true
}

// writeStoreError 把 Store 返回的错误转换成对应的 HTTP 状态码
func writeStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	log.Println("未知错误：", err)
	writeError(w, http.StatusInternalServerError, "服务器内部错误")
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

func (s *server) getTodo(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	todo, err := s.store.Get(id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, todo)
}

func (s *server) updateTodo(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	var req updateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求体必须是合法的 JSON")
		return
	}
	if req.Done == nil {
		writeError(w, http.StatusBadRequest, "请求体必须包含 done 字段")
		return
	}

	todo, err := s.store.UpdateDone(id, *req.Done)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, todo)
}

func (s *server) deleteTodo(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	if err := s.store.Delete(id); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
