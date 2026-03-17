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
	cloned, err := cloneTask(task)
	if err != nil {
		return err
	}
	s.tasks[task.ID] = cloned
	return nil
}

func (s *MemoryStore) Get(id string) (*types.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, ok := s.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task not found: %s", id)
	}
	return cloneTask(task)
}

func (s *MemoryStore) List(filter types.TaskFilter) ([]types.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]types.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		if !matchTaskFilter(*task, filter) {
			continue
		}
		cloned, err := cloneTask(task)
		if err != nil {
			return nil, err
		}
		items = append(items, *cloned)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return paginate(items, filter), nil
}
