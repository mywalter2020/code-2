package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"juyu-ai-platform/internal/types"
)

func TestRuntimeEndpointsReturnNotFoundForUnknownSession(t *testing.T) {
	srv := newTestServer(t, "", "")
	mux := http.NewServeMux()
	srv.Register(mux)

	for _, tc := range []struct {
		name string
		path string
	}{
		{name: "get session detail", path: "/api/v1/sessions/sess_missing"},
		{name: "get prd", path: "/api/v1/sessions/sess_missing/prd"},
		{name: "get todo", path: "/api/v1/sessions/sess_missing/todo"},
		{name: "get preview", path: "/api/v1/sessions/sess_missing/preview"},
		{name: "list executions", path: "/api/v1/sessions/sess_missing/executions"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
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
}

func TestRuntimeInvalidStageContracts(t *testing.T) {
	srv := newTestServer(t, "", "")
	mux := http.NewServeMux()
	srv.Register(mux)

	createRec := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", body(`{
		"input": {
			"type": "product_url_learning",
			"url": "https://example.com/product/invalid-stage",
			"message": "invalid stage contract"
		}
	}`))
	mux.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("create session failed: %d %s", createRec.Code, createRec.Body.String())
	}
	sessionID := decodeAPIResponse(t, createRec).Data.(map[string]any)["session_id"].(string)

	t.Run("edit todo before confirm prd returns 409", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/edit-todo", body(`{"items":[],"comment":"too early"}`))
		mux.ServeHTTP(rec, req)
		assertConflict(t, rec)
	})

	confirmPrdRec := httptest.NewRecorder()
	confirmPrdReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/confirm-prd", body(`{"comment":"ok"}`))
	mux.ServeHTTP(confirmPrdRec, confirmPrdReq)
	if confirmPrdRec.Code != http.StatusOK {
		t.Fatalf("confirm prd failed: %d %s", confirmPrdRec.Code, confirmPrdRec.Body.String())
	}

	t.Run("confirm prd twice returns 409", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/confirm-prd", body(`{"comment":"again"}`))
		mux.ServeHTTP(rec, req)
		assertConflict(t, rec)
	})

	confirmTodoRec := httptest.NewRecorder()
	confirmTodoReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/confirm-todo", body(`{"comment":"run"}`))
	mux.ServeHTTP(confirmTodoRec, confirmTodoReq)
	if confirmTodoRec.Code != http.StatusOK {
		t.Fatalf("confirm todo failed: %d %s", confirmTodoRec.Code, confirmTodoRec.Body.String())
	}

	t.Run("confirm todo twice returns 409", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/confirm-todo", body(`{"comment":"again"}`))
		mux.ServeHTTP(rec, req)
		assertConflict(t, rec)
	})

	t.Run("edit prd after execution started returns 409", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/edit-prd", body(`{"patch":{"title":"late change"}}`))
		mux.ServeHTTP(rec, req)
		assertConflict(t, rec)
	})

	t.Run("retry before executing on new session returns 409", func(t *testing.T) {
		freshCreateRec := httptest.NewRecorder()
		freshCreateReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", body(`{
			"input": {
				"type": "product_url_learning",
				"url": "https://example.com/product/retry-too-early",
				"message": "retry too early"
			}
		}`))
		mux.ServeHTTP(freshCreateRec, freshCreateReq)
		if freshCreateRec.Code != http.StatusOK {
			t.Fatalf("create fresh session failed: %d %s", freshCreateRec.Code, freshCreateRec.Body.String())
		}
		freshID := decodeAPIResponse(t, freshCreateRec).Data.(map[string]any)["session_id"].(string)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+freshID+"/executions/retry", body(`{"items":[{"todo_id":"todo_1"}],"reason":"too early"}`))
		mux.ServeHTTP(rec, req)
		assertConflict(t, rec)
	})
}

func assertConflict(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	resp := decodeAPIResponse(t, rec)
	if resp.Code != types.ErrCodeConflict {
		t.Fatalf("expected CONFLICT code, got %v", resp.Code)
	}
}
