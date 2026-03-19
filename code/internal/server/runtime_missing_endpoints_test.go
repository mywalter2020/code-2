package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRuntimeListSessionsAndAgentsEndpoints(t *testing.T) {
	srv := newTestServer(t, "", "")
	mux := http.NewServeMux()
	srv.Register(mux)

	createRec := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", body(`{
		"input": {
			"type": "product_url_learning",
			"url": "https://example.com/product/list",
			"message": "create one session"
		}
	}`))
	mux.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("create session failed: %d %s", createRec.Code, createRec.Body.String())
	}

	listRec := httptest.NewRecorder()
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/sessions?status=waiting_prd_confirm&type=product_url_learning", nil)
	mux.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list sessions failed: %d %s", listRec.Code, listRec.Body.String())
	}
	listResp := decodeAPIResponse(t, listRec)
	items := listResp.Data.(map[string]any)["items"].([]any)
	if len(items) == 0 {
		t.Fatalf("expected runtime sessions list items")
	}

	agentsRec := httptest.NewRecorder()
	agentsReq := httptest.NewRequest(http.MethodGet, "/api/v1/agents", nil)
	mux.ServeHTTP(agentsRec, agentsReq)
	if agentsRec.Code != http.StatusOK {
		t.Fatalf("list agents failed: %d %s", agentsRec.Code, agentsRec.Body.String())
	}
	agentsResp := decodeAPIResponse(t, agentsRec)
	agentItems := agentsResp.Data.(map[string]any)["items"].([]any)
	if len(agentItems) == 0 {
		t.Fatalf("expected runtime agents list items")
	}

	agentRec := httptest.NewRecorder()
	agentReq := httptest.NewRequest(http.MethodGet, "/api/v1/agents/product_ops", nil)
	mux.ServeHTTP(agentRec, agentReq)
	if agentRec.Code != http.StatusOK {
		t.Fatalf("get agent failed: %d %s", agentRec.Code, agentRec.Body.String())
	}
	agentResp := decodeAPIResponse(t, agentRec)
	if agentResp.Data.(map[string]any)["code"] != "product_ops" {
		t.Fatalf("expected product_ops agent detail, got %+v", agentResp.Data)
	}
}
