package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"juyu-ai-platform/internal/types"
)

func TestRuntimeExecutionNotFoundContracts(t *testing.T) {
	srv := newTestServer(t, "", "")
	mux := http.NewServeMux()
	srv.Register(mux)

	createRec := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", body(`{
		"input": {
			"type": "product_url_learning",
			"url": "https://example.com/product/not-found",
			"message": "setup for execution not found"
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

	t.Run("get missing execution returns 404", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+sessionID+"/executions/exec_missing", nil)
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
		}
		resp := decodeAPIResponse(t, rec)
		if resp.Code != types.ErrCodeNotFound {
			t.Fatalf("expected NOT_FOUND code, got %v", resp.Code)
		}
	})

	t.Run("get missing execution logs returns 404", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+sessionID+"/executions/exec_missing/logs?limit=10&offset=0", nil)
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
		}
		resp := decodeAPIResponse(t, rec)
		if resp.Code != types.ErrCodeNotFound {
			t.Fatalf("expected NOT_FOUND code, got %v", resp.Code)
		}
	})
}

func TestRuntimeQueryBoundaryContracts(t *testing.T) {
	srv := newTestServer(t, "", "")
	mux := http.NewServeMux()
	srv.Register(mux)

	for _, raw := range []string{
		`{"input":{"type":"product_url_learning","url":"https://example.com/product/q1","message":"first"}}`,
		`{"input":{"type":"product_url_learning","url":"https://example.com/product/q2","message":"second"}}`,
		`{"input":{"type":"product_url_learning","url":"https://example.com/product/q3","message":"third"}}`,
	} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", body(raw))
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("create session failed: %d %s", rec.Code, rec.Body.String())
		}
	}

	t.Run("list sessions non-numeric limit returns 400", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions?limit=abc&offset=0", nil)
		mux.ServeHTTP(rec, req)
		assertBadRequest(t, rec)
	})

	t.Run("list sessions offset beyond total returns empty items", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions?limit=2&offset=99", nil)
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		resp := decodeAPIResponse(t, rec)
		data := resp.Data.(map[string]any)
		if data["total"] != float64(3) {
			t.Fatalf("expected total 3, got %v", data["total"])
		}
		items := data["items"].([]any)
		if len(items) != 0 {
			t.Fatalf("expected empty items when offset beyond total, got %d", len(items))
		}
	})

	createExecRec := httptest.NewRecorder()
	createExecReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", body(`{
		"input": {
			"type": "product_url_learning",
			"url": "https://example.com/product/log-edge",
			"message": "log edges"
		}
	}`))
	mux.ServeHTTP(createExecRec, createExecReq)
	if createExecRec.Code != http.StatusOK {
		t.Fatalf("create execution session failed: %d %s", createExecRec.Code, createExecRec.Body.String())
	}
	execSessionID := decodeAPIResponse(t, createExecRec).Data.(map[string]any)["session_id"].(string)

	confirmPrdRec := httptest.NewRecorder()
	confirmPrdReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+execSessionID+"/confirm-prd", body(`{"comment":"ok"}`))
	mux.ServeHTTP(confirmPrdRec, confirmPrdReq)
	if confirmPrdRec.Code != http.StatusOK {
		t.Fatalf("confirm prd failed: %d %s", confirmPrdRec.Code, confirmPrdRec.Body.String())
	}
	confirmTodoRec := httptest.NewRecorder()
	confirmTodoReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+execSessionID+"/confirm-todo", body(`{"comment":"run"}`))
	mux.ServeHTTP(confirmTodoRec, confirmTodoReq)
	if confirmTodoRec.Code != http.StatusOK {
		t.Fatalf("confirm todo failed: %d %s", confirmTodoRec.Code, confirmTodoRec.Body.String())
	}

	t.Run("session logs non-numeric limit returns 400", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+execSessionID+"/logs?limit=abc&offset=0", nil)
		mux.ServeHTTP(rec, req)
		assertBadRequest(t, rec)
	})

	t.Run("session logs negative offset returns 400", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+execSessionID+"/logs?limit=10&offset=-1", nil)
		mux.ServeHTTP(rec, req)
		assertBadRequest(t, rec)
	})

	t.Run("list sessions negative limit returns 400", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions?limit=-1&offset=0", nil)
		mux.ServeHTTP(rec, req)
		assertBadRequest(t, rec)
	})

	t.Run("session logs offset beyond total returns empty items", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+execSessionID+"/logs?limit=10&offset=999", nil)
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		resp := decodeAPIResponse(t, rec)
		data := resp.Data.(map[string]any)
		items := data["items"].([]any)
		if len(items) != 0 {
			t.Fatalf("expected empty logs items with huge offset, got %d", len(items))
		}
		if data["total"].(float64) == 0 {
			t.Fatalf("expected non-zero total logs")
		}
	})
}
