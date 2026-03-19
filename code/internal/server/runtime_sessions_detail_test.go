package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRuntimeSessionDetailIncludesPrdTodoPreviewAndNextActions(t *testing.T) {
	srv := newTestServer(t, "", "")
	mux := http.NewServeMux()
	srv.Register(mux)

	createRec := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", body(`{
		"input": {
			"type": "product_url_learning",
			"url": "https://example.com/product/detail",
			"message": "create detail session"
		}
	}`))
	mux.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("create session failed: %d %s", createRec.Code, createRec.Body.String())
	}
	sessionID := decodeAPIResponse(t, createRec).Data.(map[string]any)["session_id"].(string)

	confirmPrdRec := httptest.NewRecorder()
	confirmPrdReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/confirm-prd", body(`{"comment":"ok"}`))
	mux.ServeHTTP(confirmPrdRec, confirmPrdReq)
	if confirmPrdRec.Code != http.StatusOK {
		t.Fatalf("confirm prd failed: %d %s", confirmPrdRec.Code, confirmPrdRec.Body.String())
	}

	confirmTodoRec := httptest.NewRecorder()
	confirmTodoReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/confirm-todo", body(`{"comment":"go"}`))
	mux.ServeHTTP(confirmTodoRec, confirmTodoReq)
	if confirmTodoRec.Code != http.StatusOK {
		t.Fatalf("confirm todo failed: %d %s", confirmTodoRec.Code, confirmTodoRec.Body.String())
	}

	detailRec := httptest.NewRecorder()
	detailReq := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+sessionID, nil)
	mux.ServeHTTP(detailRec, detailReq)
	if detailRec.Code != http.StatusOK {
		t.Fatalf("get session detail failed: %d %s", detailRec.Code, detailRec.Body.String())
	}

	resp := decodeAPIResponse(t, detailRec)
	data := resp.Data.(map[string]any)
	if data["session_id"] != sessionID {
		t.Fatalf("expected session id %s, got %v", sessionID, data["session_id"])
	}
	if data["status"] != "executing" {
		t.Fatalf("expected executing status, got %v", data["status"])
	}
	if data["current_stage"] != "executing" {
		t.Fatalf("expected executing stage, got %v", data["current_stage"])
	}
	if _, ok := data["prd"].(map[string]any); !ok {
		t.Fatalf("expected prd object in detail, got %T", data["prd"])
	}
	todo, ok := data["todo"].(map[string]any)
	if !ok {
		t.Fatalf("expected todo object in detail, got %T", data["todo"])
	}
	if todo["version"] == nil {
		t.Fatalf("expected todo version in detail")
	}
	preview, ok := data["preview"].(map[string]any)
	if !ok {
		t.Fatalf("expected preview object in detail, got %T", data["preview"])
	}
	if preview["hero_title"] == nil {
		t.Fatalf("expected preview.hero_title in detail, got %+v", preview)
	}
	nextActions, ok := data["next_actions"].([]any)
	if !ok || len(nextActions) == 0 {
		t.Fatalf("expected next_actions in detail, got %+v", data["next_actions"])
	}
}

func TestRuntimeListSessionsSupportsFilterAndPagination(t *testing.T) {
	srv := newTestServer(t, "", "")
	mux := http.NewServeMux()
	srv.Register(mux)

	for _, raw := range []string{
		`{"input":{"type":"product_url_learning","url":"https://example.com/product/a","message":"A"}}`,
		`{"input":{"type":"product_url_learning","url":"https://example.com/product/b","message":"B"}}`,
		`{"input":{"type":"other_runtime_type","url":"https://example.com/product/c","message":"C"}}`,
	} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", body(raw))
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("create session failed: %d %s", rec.Code, rec.Body.String())
		}
	}

	filteredRec := httptest.NewRecorder()
	filteredReq := httptest.NewRequest(http.MethodGet, "/api/v1/sessions?status=waiting_prd_confirm&type=product_url_learning&limit=1&offset=0", nil)
	mux.ServeHTTP(filteredRec, filteredReq)
	if filteredRec.Code != http.StatusOK {
		t.Fatalf("filtered list failed: %d %s", filteredRec.Code, filteredRec.Body.String())
	}
	filteredResp := decodeAPIResponse(t, filteredRec)
	filteredData := filteredResp.Data.(map[string]any)
	if filteredData["total"] != float64(2) {
		t.Fatalf("expected total 2 for filtered list, got %v", filteredData["total"])
	}
	items := filteredData["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected 1 paged item, got %d", len(items))
	}
	first := items[0].(map[string]any)
	if first["type"] != "product_url_learning" {
		t.Fatalf("expected filtered type product_url_learning, got %v", first["type"])
	}
	if first["status"] != "waiting_prd_confirm" {
		t.Fatalf("expected filtered status waiting_prd_confirm, got %v", first["status"])
	}

	nextPageRec := httptest.NewRecorder()
	nextPageReq := httptest.NewRequest(http.MethodGet, "/api/v1/sessions?status=waiting_prd_confirm&type=product_url_learning&limit=1&offset=1", nil)
	mux.ServeHTTP(nextPageRec, nextPageReq)
	if nextPageRec.Code != http.StatusOK {
		t.Fatalf("next page list failed: %d %s", nextPageRec.Code, nextPageRec.Body.String())
	}
	nextPageResp := decodeAPIResponse(t, nextPageRec)
	nextItems := nextPageResp.Data.(map[string]any)["items"].([]any)
	if len(nextItems) != 1 {
		t.Fatalf("expected 1 item on second page, got %d", len(nextItems))
	}
}
