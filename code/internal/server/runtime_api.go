package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"juyu-ai-platform/internal/types"
)

func (s *Server) handleRuntimeSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPI(w, http.StatusMethodNotAllowed, false, statusCodeToErr(http.StatusMethodNotAllowed), "", "method not allowed", nil)
		return
	}
	var req types.CreateSessionRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
		return
	}
	sess, err := s.runtimeStore.CreateSession(req.Input)
	if err != nil {
		writeAPI(w, http.StatusInternalServerError, false, statusCodeToErr(http.StatusInternalServerError), "", err.Error(), nil)
		return
	}
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{
		"session_id":   sess.SessionID,
		"stage":        sess.CurrentStage,
		"status":       sess.Status,
		"message":      sess.Message,
		"artifacts":    map[string]any{"prd": sess.PRD},
		"next_actions": sess.NextActions,
	})
}

func (s *Server) handleRuntimeSessionRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/sessions/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", "session id required", nil)
		return
	}
	sessionID := parts[0]
	if len(parts) == 1 && r.Method == http.MethodGet {
		s.handleRuntimeGetSession(w, sessionID)
		return
	}
	if len(parts) == 2 {
		switch parts[1] {
		case "prd":
			if r.Method == http.MethodGet {
				s.handleRuntimeGetPrd(w, sessionID)
				return
			}
		case "edit-prd":
			if r.Method == http.MethodPost {
				s.handleRuntimeEditPrd(w, r, sessionID)
				return
			}
		case "confirm-prd":
			if r.Method == http.MethodPost {
				s.handleRuntimeConfirmPrd(w, r, sessionID)
				return
			}
		case "todo":
			if r.Method == http.MethodGet {
				s.handleRuntimeGetTodo(w, sessionID)
				return
			}
		case "edit-todo":
			if r.Method == http.MethodPost {
				s.handleRuntimeEditTodo(w, r, sessionID)
				return
			}
		case "confirm-todo":
			if r.Method == http.MethodPost {
				s.handleRuntimeConfirmTodo(w, r, sessionID)
				return
			}
		case "executions":
			if r.Method == http.MethodGet {
				s.handleRuntimeListExecutions(w, sessionID)
				return
			}
		case "logs":
			if r.Method == http.MethodGet {
				s.handleRuntimeListLogs(w, r, sessionID, "")
				return
			}
		case "preview":
			if r.Method == http.MethodGet {
				s.handleRuntimeGetPreview(w, sessionID)
				return
			}
		case "cancel":
			if r.Method == http.MethodPost {
				s.handleRuntimeCancelSession(w, r, sessionID)
				return
			}
		}
	}
	if len(parts) == 3 && parts[1] == "executions" && parts[2] == "retry" && r.Method == http.MethodPost {
		s.handleRuntimeRetryExecutions(w, r, sessionID)
		return
	}
	if len(parts) == 3 && parts[1] == "executions" && r.Method == http.MethodGet {
		s.handleRuntimeGetExecution(w, sessionID, parts[2])
		return
	}
	if len(parts) == 4 && parts[1] == "executions" && parts[3] == "logs" && r.Method == http.MethodGet {
		s.handleRuntimeListLogs(w, r, sessionID, parts[2])
		return
	}
	writeAPI(w, http.StatusMethodNotAllowed, false, statusCodeToErr(http.StatusMethodNotAllowed), "", "unsupported runtime operation", nil)
}

func (s *Server) handleRuntimeGetSession(w http.ResponseWriter, sessionID string) {
	sess, err := s.runtimeStore.GetSession(sessionID)
	if err != nil {
		writeAPI(w, http.StatusNotFound, false, statusCodeToErr(http.StatusNotFound), "", err.Error(), nil)
		return
	}
	writeAPI(w, http.StatusOK, true, "", "ok", "", sess)
}

func (s *Server) handleRuntimeGetPrd(w http.ResponseWriter, sessionID string) {
	sess, err := s.runtimeStore.GetSession(sessionID)
	if err != nil {
		writeAPI(w, http.StatusNotFound, false, statusCodeToErr(http.StatusNotFound), "", err.Error(), nil)
		return
	}
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{"session_id": sessionID, "prd": sess.PRD})
}

func (s *Server) handleRuntimeEditPrd(w http.ResponseWriter, r *http.Request, sessionID string) {
	var req types.EditPrdRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
		return
	}
	sess, err := s.runtimeStore.EditPrd(sessionID, req.Patch, req.Comment)
	if err != nil {
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
		return
	}
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{"session_id": sessionID, "stage": sess.CurrentStage, "status": sess.Status, "message": sess.Message, "artifacts": map[string]any{"prd": sess.PRD}, "next_actions": sess.NextActions})
}

func (s *Server) handleRuntimeConfirmPrd(w http.ResponseWriter, r *http.Request, sessionID string) {
	var req types.ConfirmPrdRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
		return
	}
	sess, err := s.runtimeStore.ConfirmPrd(sessionID, req.Comment)
	if err != nil {
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
		return
	}
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{"session_id": sessionID, "stage": sess.CurrentStage, "status": sess.Status, "message": sess.Message, "artifacts": map[string]any{"todo": sess.Todo}, "next_actions": sess.NextActions})
}

func (s *Server) handleRuntimeGetTodo(w http.ResponseWriter, sessionID string) {
	todo, err := s.runtimeStore.GetTodo(sessionID)
	if err != nil {
		writeAPI(w, http.StatusNotFound, false, statusCodeToErr(http.StatusNotFound), "", err.Error(), nil)
		return
	}
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{"session_id": sessionID, "todo": todo})
}

func (s *Server) handleRuntimeEditTodo(w http.ResponseWriter, r *http.Request, sessionID string) {
	var req types.EditTodoRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
		return
	}
	sess, err := s.runtimeStore.EditTodo(sessionID, req.Items, req.Comment)
	if err != nil {
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
		return
	}
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{"session_id": sessionID, "stage": sess.CurrentStage, "status": sess.Status, "message": sess.Message, "artifacts": map[string]any{"todo": sess.Todo}, "next_actions": sess.NextActions})
}

func (s *Server) handleRuntimeConfirmTodo(w http.ResponseWriter, r *http.Request, sessionID string) {
	var req types.ConfirmTodoRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
		return
	}
	sess, err := s.runtimeStore.ConfirmTodo(sessionID, req.Comment)
	if err != nil {
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
		return
	}
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{"session_id": sessionID, "stage": sess.CurrentStage, "status": sess.Status, "message": sess.Message, "next_actions": sess.NextActions})
}

func (s *Server) handleRuntimeListExecutions(w http.ResponseWriter, sessionID string) {
	items, err := s.runtimeStore.ListExecutions(sessionID)
	if err != nil {
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
		return
	}
	sess, _ := s.runtimeStore.GetSession(sessionID)
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{"session_id": sessionID, "status": sess.Status, "items": items})
}

func (s *Server) handleRuntimeGetExecution(w http.ResponseWriter, sessionID, executionID string) {
	item, err := s.runtimeStore.GetExecution(sessionID, executionID)
	if err != nil {
		writeAPI(w, http.StatusNotFound, false, statusCodeToErr(http.StatusNotFound), "", err.Error(), nil)
		return
	}
	logs, _, _ := s.runtimeStore.ListLogs(sessionID, executionID, "", "", 100, 0)
	payload := map[string]any{}
	b, _ := json.Marshal(item)
	_ = json.Unmarshal(b, &payload)
	payload["logs"] = logs
	writeAPI(w, http.StatusOK, true, "", "ok", "", payload)
}

func (s *Server) handleRuntimeRetryExecutions(w http.ResponseWriter, r *http.Request, sessionID string) {
	var req types.RetryExecutionsRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
		return
	}
	items, sess, err := s.runtimeStore.RetryExecutions(sessionID, req)
	if err != nil {
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
		return
	}
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{"session_id": sessionID, "status": sess.Status, "message": "retry scheduled", "items": items})
}

func (s *Server) handleRuntimeListLogs(w http.ResponseWriter, r *http.Request, sessionID, executionID string) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, total, err := s.runtimeStore.ListLogs(sessionID, executionID, r.URL.Query().Get("todo_id"), r.URL.Query().Get("level"), limit, offset)
	if err != nil {
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
		return
	}
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{"session_id": sessionID, "execution_id": executionID, "items": items, "total": total})
}

func (s *Server) handleRuntimeGetPreview(w http.ResponseWriter, sessionID string) {
	preview, sess, err := s.runtimeStore.GetPreview(sessionID)
	if err != nil {
		writeAPI(w, http.StatusNotFound, false, statusCodeToErr(http.StatusNotFound), "", err.Error(), nil)
		return
	}
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{"session_id": sessionID, "status": sess.Status, "preview": preview})
}

func (s *Server) handleRuntimeCancelSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	var req types.CancelSessionRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
		return
	}
	sess, err := s.runtimeStore.CancelSession(sessionID, req.Reason)
	if err != nil {
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
		return
	}
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{"session_id": sessionID, "status": sess.Status, "message": sess.Message})
}

func decodeJSONBody(r *http.Request, dst any) error {
	if r.Body == nil {
		return nil
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}
