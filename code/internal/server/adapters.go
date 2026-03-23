package server

import (
	"context"
	"net/http"

	"juyu-ai-platform/internal/adapters"
	"juyu-ai-platform/internal/types"
)

func (s *Server) handleAdapterHealth(w http.ResponseWriter, r *http.Request) {
	writeAPI(w, http.StatusOK, true, "", "ok", "", adapterRuntimeInfo(s))
}

func adapterRuntimeInfo(s *Server) map[string]any {
	items := make([]any, 0)
	summary := map[string]any{
		"total":             0,
		"healthy":           0,
		"configured":        0,
		"live_ready":        0,
		"dry_run":           0,
		"misconfigured_live": 0,
	}
	if s == nil || s.adapterRegistry == nil {
		return map[string]any{"summary": summary, "adapters": items}
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
		summary["total"] = summary["total"].(int) + 1
		health, err := adapter.Health(context.Background())
		if err != nil {
			items = append(items, map[string]any{"platform": name, "healthy": false, "message": err.Error(), "live_ready": false})
			continue
		}
		if health.Healthy {
			summary["healthy"] = summary["healthy"].(int) + 1
		}
		if health.Configured {
			summary["configured"] = summary["configured"].(int) + 1
		}
		if health.LiveReady {
			summary["live_ready"] = summary["live_ready"].(int) + 1
		}
		if health.DryRun {
			summary["dry_run"] = summary["dry_run"].(int) + 1
		}
		if !health.DryRun && !health.LiveReady {
			summary["misconfigured_live"] = summary["misconfigured_live"].(int) + 1
		}
		items = append(items, health)
	}
	return map[string]any{"summary": summary, "adapters": items}
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
