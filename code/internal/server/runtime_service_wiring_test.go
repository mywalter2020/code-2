package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"juyu-ai-platform/internal/types"
)

func TestRuntimeServiceWiringEnrichesPrdAndPlansTodo(t *testing.T) {
	srv := newTestServer(t, "", "")
	srv.runtimeSvc = newRuntimeService(srv.rg, []types.BindingView{{MasterAgent: "product_ops", Abilities: []string{"content_gen", "page_gen", "review_check"}}})
	mux := http.NewServeMux()
	srv.Register(mux)

	createRec := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", body(`{
		"input": {
			"type": "product_url_learning",
			"url": "https://example.com/product/runtime-service",
			"message": "runtime service wiring"
		}
	}`))
	mux.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("create session failed: %d %s", createRec.Code, createRec.Body.String())
	}
	createData := decodeAPIResponse(t, createRec).Data.(map[string]any)
	sessionID := createData["session_id"].(string)
	prd := createData["artifacts"].(map[string]any)["prd"].(map[string]any)
	if prd["version"] != float64(1) {
		t.Fatalf("expected service-enriched prd to keep version 1, got %v", prd["version"])
	}
	if prd["markdown"] == nil || prd["markdown"] == "" {
		t.Fatalf("expected service-enriched prd markdown")
	}

	confirmPrdRec := httptest.NewRecorder()
	confirmPrdReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/confirm-prd", body(`{"comment":"ok"}`))
	mux.ServeHTTP(confirmPrdRec, confirmPrdReq)
	if confirmPrdRec.Code != http.StatusOK {
		t.Fatalf("confirm prd failed: %d %s", confirmPrdRec.Code, confirmPrdRec.Body.String())
	}
	confirmData := decodeAPIResponse(t, confirmPrdRec).Data.(map[string]any)
	todo := confirmData["artifacts"].(map[string]any)["todo"].(map[string]any)
	if todo["version"] != float64(1) {
		t.Fatalf("expected service-planned todo to keep version 1, got %v", todo["version"])
	}
	items := todo["items"].([]any)
	if len(items) == 0 {
		t.Fatalf("expected planned todo items")
	}
	first := items[0].(map[string]any)
	if first["type"] == "title_generation" {
		t.Fatalf("expected runtime service todo, not legacy hardcoded stub list: %+v", first)
	}
}
