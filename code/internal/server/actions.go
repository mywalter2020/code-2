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
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
		return
	}
	writeAPI(w, http.StatusOK, true, "", "ok", "", task)
}

func (s *Server) handleTaskRetry(w http.ResponseWriter, r *http.Request, taskID string) {
	task, err := s.orc.Retry(r.Context(), taskID)
	if err != nil {
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), task)
		return
	}
	writeAPI(w, http.StatusOK, true, "", "ok", "", task)
}
