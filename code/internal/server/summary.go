package server

import (
	"net/http"

	"juyu-ai-platform/internal/types"
)

func (s *Server) handleTaskSummary(w http.ResponseWriter, r *http.Request) {
	items, err := s.orc.ListTasks(types.TaskFilter{})
	if err != nil {
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
		return
	}
	summary := map[string]int{}
	for _, task := range items {
		summary[task.Status]++
	}
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{
		"total":   len(items),
		"summary": summary,
	})
}
