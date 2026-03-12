package server

import (
	"context"
	"net/http"

	"juyu-ai-platform/internal/adapters"
	"juyu-ai-platform/internal/types"
)

func (s *Server) handleAdapterHealth(w http.ResponseWriter, r *http.Request) {
	items := make([]any, 0)
	if s.adapterRegistry == nil {
		writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{"adapters": items})
		return
	}
	for _, name := range s.adapterRegistry.ListNames() {
		adapterAny, err := s.adapterRegistry.Get(name)
		if err != nil {
			continue
		}
		adapter, ok := adapterAny.(adapters.PlatformAdapter)
		if !ok {
			continue
		}
		health, err := adapter.Health(context.Background())
		if err != nil {
			items = append(items, map[string]any{"platform": name, "healthy": false, "message": err.Error()})
			continue
		}
		items = append(items, health)
	}
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{"adapters": items})
}

func (s *Server) handleAdapterCredentials(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPI(w, http.StatusMethodNotAllowed, false, statusCodeToErr(http.StatusMethodNotAllowed), "", "method not allowed", nil)
		return
	}
	items := []types.AdapterCredentials{}
	if s.credentialStore != nil {
		items = s.credentialStore.List()
	}
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{"credentials": items})
}
