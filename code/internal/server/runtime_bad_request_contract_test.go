package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"juyu-ai-platform/internal/types"
)

func TestRuntimeBadRequestContracts(t *testing.T) {
	srv := newTestServer(t, "", "")
	mux := http.NewServeMux()
	srv.Register(mux)

	t.Run("create session rejects malformed json", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", body(`{"input":`))
		mux.ServeHTTP(rec, req)
		assertBadRequest(t, rec)
	})

	t.Run("create session rejects unknown fields", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", body(`{
			"input": {
				"type": "product_url_learning",
				"url": "https://example.com/product/bad",
				"message": "bad request",
				"unexpected": true
			}
		}`))
		mux.ServeHTTP(rec, req)
		assertBadRequest(t, rec)
	})

	createRec := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", body(`{
		"input": {
			"type": "product_url_learning",
			"url": "https://example.com/product/bad-contract",
			"message": "setup valid session"
		}
	}`))
	mux.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("create session failed: %d %s", createRec.Code, createRec.Body.String())
	}
	sessionID := decodeAPIResponse(t, createRec).Data.(map[string]any)["session_id"].(string)

	t.Run("edit prd rejects malformed json", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/edit-prd", body(`{"patch":`))
		mux.ServeHTTP(rec, req)
		assertBadRequest(t, rec)
	})

	t.Run("edit prd rejects unknown fields", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/edit-prd", body(`{"patch":{"title":"x"},"nope":1}`))
		mux.ServeHTTP(rec, req)
		assertBadRequest(t, rec)
	})

	confirmPrdRec := httptest.NewRecorder()
	confirmPrdReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/confirm-prd", body(`{"comment":"ok"}`))
	mux.ServeHTTP(confirmPrdRec, confirmPrdReq)
	if confirmPrdRec.Code != http.StatusOK {
		t.Fatalf("confirm prd failed: %d %s", confirmPrdRec.Code, confirmPrdRec.Body.String())
	}

	t.Run("edit todo rejects malformed json", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/edit-todo", body(`{"items":`))
		mux.ServeHTTP(rec, req)
		assertBadRequest(t, rec)
	})

	t.Run("edit todo rejects unknown fields", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/edit-todo", body(`{"items":[],"extra":"bad"}`))
		mux.ServeHTTP(rec, req)
		assertBadRequest(t, rec)
	})

	t.Run("retry rejects malformed json", func(t *testing.T) {
		confirmTodoRec := httptest.NewRecorder()
		confirmTodoReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/confirm-todo", body(`{"comment":"run"}`))
		mux.ServeHTTP(confirmTodoRec, confirmTodoReq)
		if confirmTodoRec.Code != http.StatusOK {
			t.Fatalf("confirm todo failed: %d %s", confirmTodoRec.Code, confirmTodoRec.Body.String())
		}

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/executions/retry", body(`{"items":`))
		mux.ServeHTTP(rec, req)
		assertBadRequest(t, rec)
	})
}

func assertBadRequest(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	resp := decodeAPIResponse(t, rec)
	if resp.Code != types.ErrCodeBadRequest {
		t.Fatalf("expected BAD_REQUEST code, got %v", resp.Code)
	}
	if resp.Error == "" {
		t.Fatalf("expected non-empty error message")
	}
}
