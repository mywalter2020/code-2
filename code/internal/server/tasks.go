package server

import (
	"net/http"
	"strconv"

	"juyu-ai-platform/internal/types"
)

func (s *Server) handleTaskList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, err := s.orc.ListTasks(types.TaskFilter{
		Status: r.URL.Query().Get("status"),
		Scene:  r.URL.Query().Get("scene"),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tasks": items, "count": len(items)})
}
