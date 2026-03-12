package store

import (
	"fmt"
	"sort"
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

func (s *MemoryStore) Save(task *types.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[task.ID] = task
	return nil
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

func (s *MemoryStore) List(filter types.TaskFilter) ([]types.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]types.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		if filter.Status != "" && task.Status != filter.Status {
			continue
		}
		if filter.Scene != "" && task.Request.Scene != filter.Scene {
			continue
		}
		items = append(items, *task)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	start := filter.Offset
	if start > len(items) {
		return []types.Task{}, nil
	}
	end := len(items)
	if filter.Limit > 0 && start+filter.Limit < end {
		end = start + filter.Limit
	}
	return items[start:end], nil
}
