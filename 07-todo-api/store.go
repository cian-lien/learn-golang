package main

import (
	"errors"
	"slices"
	"sync"
)

type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// ErrNotFound 表示找不到指定的待办
var ErrNotFound = errors.New("todo 不存在")

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

func (s *Store) Get(id int) (Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, t := range s.todos {
		if t.ID == id {
			return t, nil
		}
	}
	return Todo{}, ErrNotFound
}

func (s *Store) UpdateDone(id int, done bool) (Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.todos {
		if s.todos[i].ID == id {
			s.todos[i].Done = done
			return s.todos[i], nil
		}
	}
	return Todo{}, ErrNotFound
}

func (s *Store) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, t := range s.todos {
		if t.ID == id {
			s.todos = slices.Delete(s.todos, i, i+1)
			return nil
		}
	}
	return ErrNotFound
}
