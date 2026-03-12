package server

import "net/http"

func (s *Server) handleTaskStatus(w http.ResponseWriter, taskID string) {
	task, err := s.orc.GetTask(taskID)
	if err != nil {
		writeAPI(w, http.StatusNotFound, false, statusCodeToErr(http.StatusNotFound), "", err.Error(), nil)
		return
	}
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{
		"task_id": task.ID,
		"status":  task.Status,
		"step":    task.CurrentStep,
		"updated": task.UpdatedAt,
	})
}
