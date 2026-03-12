package server

import (
	"net/http"

	"juyu-ai-platform/internal/types"
)

func buildAbilityMetadata(items []types.AbilityConfig) []types.AbilityMetadata {
	result := make([]types.AbilityMetadata, 0, len(items))
	for _, item := range items {
		result = append(result, types.AbilityMetadata{
			Code:        item.Code,
			Name:        item.Name,
			Type:        item.Type,
			Description: item.Description,
			Tags:        item.Tags,
			Version:     item.Version,
			Enabled:     item.Enabled,
		})
	}
	return result
}

func (s *Server) handleAbilityMetadata(w http.ResponseWriter, r *http.Request) {
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{"abilities": s.abilityMetadata})
}
