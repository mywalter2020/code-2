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

func newTestServer(t *testing.T, apiKey string, operatorTokens string) *Server {
	return newTestServerWithRuntimeStore(t, apiKey, operatorTokens, nil, "")
}

func newTestServerWithRuntimeStore(t *testing.T, apiKey string, operatorTokens string, runtimeStore store.RuntimeStore, runtimeDriver string) *Server {
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
		operatorTokens,
	)
	AttachAdapterRuntime(srv, adapterRegistry, adapters.NewCredentialStore())
	if runtimeStore != nil {
		AttachRuntimeStore(srv, runtimeStore, runtimeDriver)
	}
	return srv
}

func TestExecuteRequiresAPIKey(t *testing.T) {
	srv := newTestServer(t, "secret", "")
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
	srv := newTestServer(t, "secret", "")
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

func TestExecuteUsesOperatorIdentityHeader(t *testing.T) {
	srv := newTestServer(t, "secret", "alice:token-1")
	mux := http.NewServeMux()
	srv.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/execute", body(`{"scene":"product","input":"demo","payload":{"platform":"alibaba"}}`))
	req.Header.Set("X-API-Key", "secret")
	req.Header.Set("X-Operator-ID", "alice")
	req.Header.Set("X-Operator-Token", "token-1")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	resp := decodeAPIResponse(t, rec)
	data := resp.Data.(map[string]any)
	if data["task_id"] == "" {
		t.Fatal("expected task id")
	}

	detail := httptest.NewRecorder()
	mux.ServeHTTP(detail, httptest.NewRequest(http.MethodGet, "/tasks/"+data["task_id"].(string), nil))
	if detail.Code != http.StatusOK {
		t.Fatalf("expected 200 for task detail, got %d", detail.Code)
	}
	taskResp := decodeAPIResponse(t, detail)
	taskData := taskResp.Data.(map[string]any)
	if taskData["operator"] != "alice" {
		t.Fatalf("expected operator alice, got %v", taskData["operator"])
	}
}

func TestExecuteRejectsOperatorMismatch(t *testing.T) {
	srv := newTestServer(t, "secret", "alice:token-1")
	mux := http.NewServeMux()
	srv.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/execute", body(`{"scene":"product","input":"demo","operator":"bob","payload":{"platform":"alibaba"}}`))
	req.Header.Set("X-API-Key", "secret")
	req.Header.Set("X-Operator-ID", "alice")
	req.Header.Set("X-Operator-Token", "token-1")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestProviderAdminUpdatesRuntimeConfig(t *testing.T) {
	t.Setenv("CONTENT_GEN_PROVIDER", "stub")
	t.Setenv("PAGE_GEN_PROVIDER", "stub")

	srv := newTestServer(t, "secret", "")
	mux := http.NewServeMux()
	srv.Register(mux)

	req := httptest.NewRequest(http.MethodPut, "/admin/providers", body(`{"content_gen":{"provider":"openai_compat","model":"gpt-test"},"page_gen":{"provider":"stub"}}`))
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

func body(s string) *bytes.Buffer { return bytes.NewBufferString(s) }

func decodeAPIResponse(t *testing.T, rec *httptest.ResponseRecorder) types.APIResponse {
	t.Helper()
	var resp types.APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp
}

func createTaskForTest(t *testing.T, mux *http.ServeMux) string {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/execute", body(`{"scene":"product","input":"demo","payload":{"platform":"alibaba"}}`))
	req.Header.Set("X-API-Key", "secret")
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("create task failed: %d %s", rec.Code, rec.Body.String())
	}
	resp := decodeAPIResponse(t, rec)
	data := resp.Data.(map[string]any)
	return data["task_id"].(string)
}
