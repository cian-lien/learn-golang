package main

import (
	"sync"
)

type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// Store 负责保存待办数据。所有方法内部都会加锁，调用者不需要关心锁。
type Store struct {
	mu     sync.Mutex
	todos  []Todo
	nextID int
}

func NewStore() *Store {
	return &Store{
		todos:  []Todo{},
		nextID: 1,
	}
}

// List 返回待办列表的副本。done 为 nil 时返回全部，否则只返回 Done 等于 *done 的。
func (s *Store) List(done *bool) []Todo {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := []Todo{}
	for _, t := range s.todos {
		if done == nil || t.Done == *done {
			result = append(result, t)
		}
	}
	return result
}

func (s *Store) Create(title string) Todo {
	s.mu.Lock()
	defer s.mu.Unlock()

	todo := Todo{ID: s.nextID, Title: title, Done: false}
	s.nextID++
	s.todos = append(s.todos, todo)
	return todo
}
