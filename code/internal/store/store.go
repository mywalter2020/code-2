package store

import "juyu-ai-platform/internal/types"

type TaskStore interface {
	Save(task *types.Task) error
	Get(id string) (*types.Task, error)
	List(filter types.TaskFilter) ([]types.Task, error)
}
