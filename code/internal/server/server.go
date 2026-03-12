package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"juyu-ai-platform/internal/orchestrator"
	"juyu-ai-platform/internal/registry"
	"juyu-ai-platform/internal/types"
)

type Server struct {
	orc *orchestrator.Orchestrator
	rg  *registry.Registry
}

func New(orc *orchestrator.Orchestrator, rg *registry.Registry) *Server {
	return &Server{orc: orc, rg: rg}
}

func (s *Server) Register(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/abilities", s.handleAbilities)
	mux.HandleFunc("/execute", s.handleExecute)
	mux.HandleFunc("/tasks", s.handleTaskList)
	mux.HandleFunc("/tasks/", s.handleTasks)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (s *Server) handleAbilities(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"abilities": s.rg.ListCodes()})
}

func (s *Server) handleExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}

	var req types.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	resp, err := s.orc.Execute(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error":   err.Error(),
			"task_id": resp.TaskID,
			"status":  resp.Status,
			"preview": resp.Preview,
			"results": resp.Results,
		})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "task id required"})
		return
	}
	taskID := parts[0]

	if len(parts) == 1 && r.Method == http.MethodGet {
		task, err := s.orc.GetTask(taskID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, task)
		return
	}
	if len(parts) == 2 && parts[1] == "preview" && r.Method == http.MethodGet {
		s.handleTaskPreview(w, taskID)
		return
	}
	if len(parts) == 2 && parts[1] == "logs" && r.Method == http.MethodGet {
		s.handleTaskLogs(w, taskID)
		return
	}

	if len(parts) == 2 && parts[1] == "confirm" && r.Method == http.MethodPost {
		var req types.ConfirmRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		task, err := s.orc.Confirm(r.Context(), taskID, req.Approved, req.Comment)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error(), "task": task})
			return
		}
		writeJSON(w, http.StatusOK, task)
		return
	}
	if len(parts) == 2 && parts[1] == "cancel" && r.Method == http.MethodPost {
		s.handleTaskCancel(w, r, taskID)
		return
	}
	if len(parts) == 2 && parts[1] == "retry" && r.Method == http.MethodPost {
		s.handleTaskRetry(w, r, taskID)
		return
	}

	writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "unsupported task operation"})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
