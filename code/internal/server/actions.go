package server

import (
	"encoding/json"
	"net/http"
)

type taskActionRequest struct {
	Comment string `json:"comment,omitempty"`
}

func (s *Server) handleTaskCancel(w http.ResponseWriter, r *http.Request, taskID string) {
	var req taskActionRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	task, err := s.orc.Cancel(taskID, req.Comment)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (s *Server) handleTaskRetry(w http.ResponseWriter, r *http.Request, taskID string) {
	task, err := s.orc.Retry(r.Context(), taskID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error(), "task": task})
		return
	}
	writeJSON(w, http.StatusOK, task)
}
