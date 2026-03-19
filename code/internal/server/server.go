package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"juyu-ai-platform/internal/orchestrator"
	"juyu-ai-platform/internal/registry"
	"juyu-ai-platform/internal/store"
	"juyu-ai-platform/internal/types"
)

type Server struct {
	orc             *orchestrator.Orchestrator
	rg              *registry.Registry
	runtimeStore    store.RuntimeStore
	abilityMetadata []types.AbilityMetadata
	masterMetadata  []types.MasterAgentMetadata
	bindingViews    []types.BindingView
	adapterRegistry interface {
		ListNames() []string
		Get(string) (interface{}, error)
	}
	credentialStore interface {
		List() []types.AdapterCredentials
	}
	apiKey        string
	operatorAuth  operatorAuth
	runtimeDriver string
}

func New(orc *orchestrator.Orchestrator, rg *registry.Registry, abilityMetadata []types.AbilityMetadata, masterMetadata []types.MasterAgentMetadata, bindingViews []types.BindingView, apiKey string, operatorTokens string) *Server {
	return &Server{
		orc:             orc,
		rg:              rg,
		runtimeStore:    store.NewRuntimeMemoryStore(),
		runtimeDriver:   "memory",
		abilityMetadata: abilityMetadata,
		masterMetadata:  masterMetadata,
		bindingViews:    bindingViews,
		apiKey:          apiKey,
		operatorAuth:    newOperatorAuth(operatorTokens),
	}
}

func (s *Server) Register(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/abilities", s.handleAbilities)
	mux.HandleFunc("/abilities/metadata", s.handleAbilityMetadata)
	mux.HandleFunc("/agents/metadata", s.handleMasterMetadata)
	mux.HandleFunc("/bindings", s.handleBindings)
	mux.HandleFunc("/adapters/health", s.handleAdapterHealth)
	mux.HandleFunc("/adapters/descriptors", s.handleAdapterDescriptors)
	mux.HandleFunc("/adapters/credentials", s.handleAdapterCredentials)
	mux.HandleFunc("/admin/providers", s.handleProviderAdmin)
	mux.HandleFunc("/execute", s.handleExecute)
	mux.HandleFunc("/tasks/summary", s.handleTaskSummary)
	mux.HandleFunc("/tasks", s.handleTaskList)
	mux.HandleFunc("/tasks/", s.handleTasks)
	mux.HandleFunc("/api/v1/sessions", s.handleRuntimeSessions)
	mux.HandleFunc("/api/v1/sessions/", s.handleRuntimeSessionRoutes)
	s.registerUI(mux)
}

func (s *Server) handleExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPI(w, http.StatusMethodNotAllowed, false, statusCodeToErr(http.StatusMethodNotAllowed), "", "method not allowed", nil)
		return
	}
	if !s.requireWriteAuth(w, r) {
		return
	}
	identity, ok := s.requireOperator(w, r)
	if !ok {
		return
	}

	var req types.Request
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
		return
	}
	operator, err := pickOperator(req.Operator, identity)
	if err != nil {
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
		return
	}
	req.Operator = operator

	resp, err := s.orc.Execute(r.Context(), req)
	if err != nil {
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), map[string]any{
			"task_id": resp.TaskID,
			"status":  resp.Status,
			"preview": resp.Preview,
			"results": resp.Results,
		})
		return
	}
	writeAPI(w, http.StatusOK, true, "", "ok", "", resp)
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", "task id required", nil)
		return
	}
	taskID := parts[0]

	if len(parts) == 1 && r.Method == http.MethodGet {
		task, err := s.orc.GetTask(taskID)
		if err != nil {
			writeAPI(w, http.StatusNotFound, false, statusCodeToErr(http.StatusNotFound), "", err.Error(), nil)
			return
		}
		writeAPI(w, http.StatusOK, true, "", "ok", "", task)
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
	if len(parts) == 2 && parts[1] == "status" && r.Method == http.MethodGet {
		s.handleTaskStatus(w, taskID)
		return
	}

	if len(parts) == 2 && parts[1] == "confirm" && r.Method == http.MethodPost {
		if !s.requireWriteAuth(w, r) {
			return
		}
		identity, ok := s.requireOperator(w, r)
		if !ok {
			return
		}
		var req types.ConfirmRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
			return
		}
		approver, err := pickOperator(req.Approver, identity)
		if err != nil {
			writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
			return
		}
		task, err := s.orc.Confirm(r.Context(), taskID, req.Approved, req.Comment, approver)
		if err != nil {
			writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), task)
			return
		}
		writeAPI(w, http.StatusOK, true, "", "ok", "", task)
		return
	}
	if len(parts) == 2 && parts[1] == "cancel" && r.Method == http.MethodPost {
		if !s.requireWriteAuth(w, r) {
			return
		}
		if _, ok := s.requireOperator(w, r); !ok {
			return
		}
		s.handleTaskCancel(w, r, taskID)
		return
	}
	if len(parts) == 2 && parts[1] == "retry" && r.Method == http.MethodPost {
		if !s.requireWriteAuth(w, r) {
			return
		}
		if _, ok := s.requireOperator(w, r); !ok {
			return
		}
		s.handleTaskRetry(w, r, taskID)
		return
	}

	writeAPI(w, http.StatusMethodNotAllowed, false, statusCodeToErr(http.StatusMethodNotAllowed), "", "unsupported task operation", nil)
}
