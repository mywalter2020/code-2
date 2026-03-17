package store

import (
	"encoding/json"

	"juyu-ai-platform/internal/types"
)

func cloneTask(task *types.Task) (*types.Task, error) {
	if task == nil {
		return nil, nil
	}
	b, err := json.Marshal(task)
	if err != nil {
		return nil, err
	}
	var cloned types.Task
	if err := json.Unmarshal(b, &cloned); err != nil {
		return nil, err
	}
	return &cloned, nil
}
