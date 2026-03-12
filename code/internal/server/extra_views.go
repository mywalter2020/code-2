package server

import "net/http"

func (s *Server) handleTaskPreview(w http.ResponseWriter, taskID string) {
	task, err := s.orc.GetTask(taskID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"task_id": task.ID, "preview": task.Preview})
}

func (s *Server) handleTaskLogs(w http.ResponseWriter, taskID string) {
	task, err := s.orc.GetTask(taskID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"task_id": task.ID, "logs": task.Logs, "count": len(task.Logs)})
}
