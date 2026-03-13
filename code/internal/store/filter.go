package store

import (
	"strings"

	"juyu-ai-platform/internal/types"
)

func matchTaskFilter(task types.Task, filter types.TaskFilter) bool {
	if filter.Status != "" && task.Status != filter.Status {
		return false
	}
	if filter.Scene != "" && task.Request.Scene != filter.Scene {
		return false
	}
	if filter.Operator != "" && !strings.EqualFold(task.Operator, filter.Operator) && !strings.EqualFold(task.Request.Operator, filter.Operator) {
		return false
	}
	if filter.Platform != "" {
		platform := ""
		if task.Request.Payload != nil {
			if p, ok := task.Request.Payload["platform"].(string); ok {
				platform = p
			}
		}
		if !strings.EqualFold(platform, filter.Platform) {
			return false
		}
	}
	if q := strings.TrimSpace(strings.ToLower(filter.Query)); q != "" {
		blob := strings.ToLower(strings.Join([]string{
			task.ID,
			task.MasterAgent,
			task.Status,
			task.Request.Scene,
			task.Request.Input,
			task.Operator,
			task.Approver,
			task.ErrorMessage,
		}, " "))
		if !strings.Contains(blob, q) {
			return false
		}
	}
	return true
}

func paginate(items []types.Task, filter types.TaskFilter) []types.Task {
	start := filter.Offset
	if start > len(items) {
		return []types.Task{}
	}
	end := len(items)
	if filter.Limit > 0 && start+filter.Limit < end {
		end = start + filter.Limit
	}
	return items[start:end]
}
