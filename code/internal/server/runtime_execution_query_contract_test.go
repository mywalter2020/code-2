package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRuntimeExecutionsListFilterAndPaginationContracts(t *testing.T) {
	srv := newTestServer(t, "", "")
	mux := http.NewServeMux()
	srv.Register(mux)

	createRec := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", body(`{
		"input": {
			"type": "product_url_learning",
			"url": "https://example.com/product/executions-query",
			"message": "executions list query contract"
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
	confirmTodoReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/confirm-todo", body(`{"comment":"run"}`))
	mux.ServeHTTP(confirmTodoRec, confirmTodoReq)
	if confirmTodoRec.Code != http.StatusOK {
		t.Fatalf("confirm todo failed: %d %s", confirmTodoRec.Code, confirmTodoRec.Body.String())
	}

	t.Run("executions list filter by status", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+sessionID+"/executions?status=failed", nil)
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		resp := decodeAPIResponse(t, rec)
		data := resp.Data.(map[string]any)
		items := data["items"].([]any)
		if len(items) != 1 {
			t.Fatalf("expected 1 failed execution, got %d", len(items))
		}
		if items[0].(map[string]any)["status"] != "failed" {
			t.Fatalf("expected failed execution, got %+v", items[0])
		}
	})

	t.Run("executions list filter by todo_id", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+sessionID+"/executions?todo_id=todo_2", nil)
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		resp := decodeAPIResponse(t, rec)
		items := resp.Data.(map[string]any)["items"].([]any)
		if len(items) != 1 {
			t.Fatalf("expected 1 execution for todo_2, got %d", len(items))
		}
		if items[0].(map[string]any)["todo_id"] != "todo_2" {
			t.Fatalf("expected todo_2 execution, got %+v", items[0])
		}
	})

	t.Run("executions list pagination returns total and paged items", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+sessionID+"/executions?limit=1&offset=1", nil)
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		resp := decodeAPIResponse(t, rec)
		data := resp.Data.(map[string]any)
		if data["total"].(float64) < 3 {
			t.Fatalf("expected total >= 3, got %v", data["total"])
		}
		items := data["items"].([]any)
		if len(items) != 1 {
			t.Fatalf("expected 1 paged item, got %d", len(items))
		}
	})

	t.Run("executions list invalid limit returns 400", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+sessionID+"/executions?limit=abc", nil)
		mux.ServeHTTP(rec, req)
		assertBadRequest(t, rec)
	})

	t.Run("executions list negative offset returns 400", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+sessionID+"/executions?offset=-1", nil)
		mux.ServeHTTP(rec, req)
		assertBadRequest(t, rec)
	})
}

func TestRuntimeLogsFilterContractsRemainConsistent(t *testing.T) {
	srv := newTestServer(t, "", "")
	mux := http.NewServeMux()
	srv.Register(mux)

	createRec := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", body(`{
		"input": {
			"type": "product_url_learning",
			"url": "https://example.com/product/log-filter-consistency",
			"message": "logs filter contract"
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
	confirmTodoReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/confirm-todo", body(`{"comment":"run"}`))
	mux.ServeHTTP(confirmTodoRec, confirmTodoReq)
	if confirmTodoRec.Code != http.StatusOK {
		t.Fatalf("confirm todo failed: %d %s", confirmTodoRec.Code, confirmTodoRec.Body.String())
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+sessionID+"/logs?level=error&todo_id=todo_2&limit=10&offset=0", nil)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	resp := decodeAPIResponse(t, rec)
	items := resp.Data.(map[string]any)["items"].([]any)
	if len(items) == 0 {
		t.Fatalf("expected filtered logs")
	}
	for _, raw := range items {
		item := raw.(map[string]any)
		if item["level"] != "error" {
			t.Fatalf("expected error level, got %+v", item)
		}
		if item["todo_id"] != "todo_2" {
			t.Fatalf("expected todo_2 filter, got %+v", item)
		}
	}
}
