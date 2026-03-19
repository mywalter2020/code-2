package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"juyu-ai-platform/internal/store"
	"juyu-ai-platform/internal/types"
)

func TestRuntimePGBackedQueryAndFilterContractsIfConfigured(t *testing.T) {
	dsn := os.Getenv("JUYU_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("JUYU_TEST_PG_DSN not set")
	}

	runtimeStore, err := store.NewRuntimePostgresStore(dsn)
	if err != nil {
		t.Fatalf("new runtime postgres store: %v", err)
	}

	srv := newTestServerWithRuntimeStore(t, "", "", runtimeStore, "postgres")
	mux := http.NewServeMux()
	srv.Register(mux)

	createRec := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", body(`{
		"input": {
			"type": "product_url_learning",
			"url": "https://example.com/product/pg-query",
			"message": "pg query contract"
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

	editTodoRec := httptest.NewRecorder()
	editTodoReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/edit-todo", body(`{
		"items": [
			{"id":"todo_1","title":"生成产品标题","type":"title_generation","status":"pending","depends_on":[],"parallel_group":"content_assets"},
			{"id":"todo_2","title":"生成图片描述图内容","type":"feature_image_copy","status":"pending","depends_on":[],"parallel_group":"content_assets"},
			{"id":"todo_3","title":"生成产品轮播图内容","type":"carousel_generation","status":"pending","depends_on":[],"parallel_group":"content_assets"}
		],
		"comment":"pg-backed query setup"
	}`))
	mux.ServeHTTP(editTodoRec, editTodoReq)
	if editTodoRec.Code != http.StatusOK {
		t.Fatalf("edit todo failed: %d %s", editTodoRec.Code, editTodoRec.Body.String())
	}

	confirmTodoRec := httptest.NewRecorder()
	confirmTodoReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/confirm-todo", body(`{"comment":"run"}`))
	mux.ServeHTTP(confirmTodoRec, confirmTodoReq)
	if confirmTodoRec.Code != http.StatusOK {
		t.Fatalf("confirm todo failed: %d %s", confirmTodoRec.Code, confirmTodoRec.Body.String())
	}

	t.Run("executions filter by status works on pg-backed path", func(t *testing.T) {
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

	t.Run("executions filter by todo_id works on pg-backed path", func(t *testing.T) {
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

	t.Run("executions pagination works on pg-backed path", func(t *testing.T) {
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
		if len(data["items"].([]any)) != 1 {
			t.Fatalf("expected 1 paged item, got %d", len(data["items"].([]any)))
		}
	})

	t.Run("executions invalid query returns 400 on pg-backed path", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+sessionID+"/executions?limit=abc", nil)
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
		}
		resp := decodeAPIResponse(t, rec)
		if resp.Code != types.ErrCodeBadRequest {
			t.Fatalf("expected BAD_REQUEST, got %v", resp.Code)
		}
	})

	t.Run("logs combined filters work on pg-backed path", func(t *testing.T) {
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
	})

	t.Run("missing execution logs returns 404 on pg-backed path", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+sessionID+"/executions/exec_missing/logs?limit=10&offset=0", nil)
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
		}
		resp := decodeAPIResponse(t, rec)
		if resp.Code != types.ErrCodeNotFound {
			t.Fatalf("expected NOT_FOUND, got %v", resp.Code)
		}
	})
}
