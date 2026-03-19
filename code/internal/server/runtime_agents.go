package server

import (
	"net/http"
	"strings"
)

func (s *Server) handleRuntimeAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPI(w, http.StatusMethodNotAllowed, false, statusCodeToErr(http.StatusMethodNotAllowed), "", "method not allowed", nil)
		return
	}
	items := make([]map[string]any, 0, len(s.masterMetadata))
	for _, item := range s.masterMetadata {
		items = append(items, map[string]any{
			"code": item.Code,
			"name": item.Name,
			"role": item.SceneType,
			"profile": map[string]any{
				"identity":       item.Name,
				"memory_summary": "runtime draft profile placeholder",
			},
		})
	}
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{"items": items})
}

func (s *Server) handleRuntimeAgentRoutes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPI(w, http.StatusMethodNotAllowed, false, statusCodeToErr(http.StatusMethodNotAllowed), "", "method not allowed", nil)
		return
	}
	code := strings.TrimPrefix(r.URL.Path, "/api/v1/agents/")
	code = strings.Trim(code, "/")
	if code == "" {
		writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", "agent code required", nil)
		return
	}
	for _, item := range s.masterMetadata {
		if item.Code == code {
			writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{
				"code": item.Code,
				"name": item.Name,
				"role": item.SceneType,
				"profile": map[string]any{
					"identity":       item.Name,
					"memory_summary": "runtime draft profile placeholder",
				},
			})
			return
		}
	}
	writeAPI(w, http.StatusNotFound, false, statusCodeToErr(http.StatusNotFound), "", "agent not found", nil)
}
