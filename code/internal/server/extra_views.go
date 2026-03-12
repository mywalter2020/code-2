package server

import "net/http"

func (s *Server) handleTaskPreview(w http.ResponseWriter, taskID string) {
	task, err := s.orc.GetTask(taskID)
	if err != nil {
		writeAPI(w, http.StatusNotFound, false, statusCodeToErr(http.StatusNotFound), "", err.Error(), nil)
		return
	}
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{"task_id": task.ID, "preview": task.Preview})
}

func (s *Server) handleTaskLogs(w http.ResponseWriter, taskID string) {
	task, err := s.orc.GetTask(taskID)
	if err != nil {
		writeAPI(w, http.StatusNotFound, false, statusCodeToErr(http.StatusNotFound), "", err.Error(), nil)
		return
	}
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{"task_id": task.ID, "logs": task.Logs, "count": len(task.Logs)})
}
