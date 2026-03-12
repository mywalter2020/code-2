package server

import (
	"net/http"

	"juyu-ai-platform/internal/types"
)

func buildMasterMetadata(items []types.MasterAgent) []types.MasterAgentMetadata {
	result := make([]types.MasterAgentMetadata, 0, len(items))
	for _, item := range items {
		result = append(result, types.MasterAgentMetadata{
			Code:      item.Code,
			Name:      item.Name,
			SceneType: item.SceneType,
			Enabled:   item.Enabled,
		})
	}
	return result
}

func (s *Server) handleMasterMetadata(w http.ResponseWriter, r *http.Request) {
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{"agents": s.masterMetadata})
}

func (s *Server) handleBindings(w http.ResponseWriter, r *http.Request) {
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{"bindings": s.bindingViews})
}
