package store

import (
	"fmt"
	"sync"

	"juyu-ai-platform/internal/types"
)

type MemoryStore struct {
	mu    sync.RWMutex
	tasks map[string]*types.Task
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{tasks: make(map[string]*types.Task)}
}

func (s *MemoryStore) Save(task *types.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[task.ID] = task
}

func (s *MemoryStore) Get(id string) (*types.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, ok := s.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task not found: %s", id)
	}
	return task, nil
}
