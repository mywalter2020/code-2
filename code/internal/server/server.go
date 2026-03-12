package server

import (
	"encoding/json"
	"net/http"

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
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
	})
}

func (s *Server) handleAbilities(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"abilities": s.rg.ListCodes(),
	})
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
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
