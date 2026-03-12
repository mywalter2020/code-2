package server

import "net/http"

func (s *Server) handleTaskList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}
	items := s.orc.ListTasks()
	writeJSON(w, http.StatusOK, map[string]any{"tasks": items, "count": len(items)})
}
