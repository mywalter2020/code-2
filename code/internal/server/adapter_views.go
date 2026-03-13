package server

import (
	"net/http"

	"juyu-ai-platform/internal/adapters"
)

func (s *Server) handleAdapterDescriptors(w http.ResponseWriter, r *http.Request) {
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
		items = append(items, adapter.Descriptor())
	}
	writeAPI(w, http.StatusOK, true, "", "ok", "", map[string]any{"adapters": items})
}
