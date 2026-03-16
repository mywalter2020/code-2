package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"juyu-ai-platform/internal/adapters"
	"juyu-ai-platform/internal/agents"
	"juyu-ai-platform/internal/orchestrator"
	"juyu-ai-platform/internal/registry"
	"juyu-ai-platform/internal/router"
	"juyu-ai-platform/internal/store"
	"juyu-ai-platform/internal/types"
)

func newTestServer(t *testing.T, apiKey string) *Server {
	t.Helper()

	adapterRegistry := adapters.NewRegistry()
	adapterRegistry.Register(adapters.NewAlibabaAdapter(adapters.NewCredentialStore(), true))

	rg := registry.New()
	rg.Register(agents.NewContentAgent())
	rg.Register(agents.NewPageAgent())
	rg.Register(agents.NewReviewAgent())
	rg.Register(agents.NewPublishAgent(adapterRegistry))

	rt := router.New([]types.MasterAgent{{Code: "product_ops", SceneType: "product", Enabled: true}})
	st := store.NewMemoryStore()
	orc := orchestrator.New(rt, rg, st, []types.Binding{{
		MasterAgent: "product_ops",
		Workflow: []types.WorkflowStep{{
			Step: 1, Ability: "content_gen",
		}, {
			Step: 2, Ability: "page_gen",
		}, {
			Step: 3, Ability: "review_check", RequireHumanConfirm: true,
		}, {
			Step: 4, Ability: "publish_exec",
		}},
	}})

	srv := New(
		orc,
		rg,
		BuildAbilityMetadata([]types.AbilityConfig{{Code: "content_gen", Name: "Content", Enabled: true}}),
		BuildMasterMetadata([]types.MasterAgent{{Code: "product_ops", Name: "Product Ops", SceneType: "product", Enabled: true}}),
		BuildBindingViews([]types.Binding{{MasterAgent: "product_ops"}}),
		apiKey,
	)
	AttachAdapterRuntime(srv, adapterRegistry, adapters.NewCredentialStore())
	return srv
}

func TestExecuteRequiresAPIKey(t *testing.T) {
	srv := newTestServer(t, "secret")
	mux := http.NewServeMux()
	srv.Register(mux)

	body := bytes.NewBufferString(`{"scene":"product","input":"demo"}`)
	req := httptest.NewRequest(http.MethodPost, "/execute", body)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	var resp types.APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Code != types.ErrCodeUnauthorized {
		t.Fatalf("expected error code %q, got %q", types.ErrCodeUnauthorized, resp.Code)
	}
}

func TestExecuteWithAPIKeyCreatesTask(t *testing.T) {
	srv := newTestServer(t, "secret")
	mux := http.NewServeMux()
	srv.Register(mux)

	body := bytes.NewBufferString(`{"scene":"product","input":"demo","payload":{"platform":"alibaba"}}`)
	req := httptest.NewRequest(http.MethodPost, "/execute", body)
	req.Header.Set("X-API-Key", "secret")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp types.APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload, got %T", resp.Data)
	}
	if data["task_id"] == "" {
		t.Fatal("expected task_id in response")
	}
	if data["status"] != types.TaskStatusPendingConfirm {
		t.Fatalf("expected status %q, got %v", types.TaskStatusPendingConfirm, data["status"])
	}
}

func TestProviderAdminUpdatesRuntimeConfig(t *testing.T) {
	t.Setenv("CONTENT_GEN_PROVIDER", "stub")
	t.Setenv("PAGE_GEN_PROVIDER", "stub")

	srv := newTestServer(t, "secret")
	mux := http.NewServeMux()
	srv.Register(mux)

	body := bytes.NewBufferString(`{"content_gen":{"provider":"openai_compat","model":"gpt-test"},"page_gen":{"provider":"stub"}}`)
	req := httptest.NewRequest(http.MethodPut, "/admin/providers", body)
	req.Header.Set("X-API-Key", "secret")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := providerRuntimeInfo()["content_gen"].(map[string]any)["provider"]; got != "openai_compat" {
		t.Fatalf("expected runtime provider update, got %v", got)
	}
}
