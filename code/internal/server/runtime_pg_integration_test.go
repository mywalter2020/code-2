package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"juyu-ai-platform/internal/store"
)

func TestRuntimeServerFlowWithPostgresStoreIfConfigured(t *testing.T) {
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
			"url": "https://example.com/product/pg-server",
			"message": "pg-backed runtime server flow"
		}
	}`))
	mux.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("create session failed: %d %s", createRec.Code, createRec.Body.String())
	}
	createData := decodeAPIResponse(t, createRec).Data.(map[string]any)
	sessionID := createData["session_id"].(string)

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
			{"id":"todo_2","title":"生成图片描述图内容","type":"feature_image_copy","status":"pending","depends_on":[],"parallel_group":"content_assets"}
		],
		"comment":"pg todo version bump"
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

	detailRec := httptest.NewRecorder()
	detailReq := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+sessionID, nil)
	mux.ServeHTTP(detailRec, detailReq)
	if detailRec.Code != http.StatusOK {
		t.Fatalf("get session detail failed: %d %s", detailRec.Code, detailRec.Body.String())
	}
	detailData := decodeAPIResponse(t, detailRec).Data.(map[string]any)
	if detailData["status"] != "executing" {
		t.Fatalf("expected executing status, got %v", detailData["status"])
	}
	todo := detailData["todo"].(map[string]any)
	if todo["version"] != float64(2) {
		t.Fatalf("expected todo version 2 from pg-backed detail, got %v", todo["version"])
	}

	listExecRec := httptest.NewRecorder()
	listExecReq := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+sessionID+"/executions", nil)
	mux.ServeHTTP(listExecRec, listExecReq)
	if listExecRec.Code != http.StatusOK {
		t.Fatalf("list executions failed: %d %s", listExecRec.Code, listExecRec.Body.String())
	}
	items := decodeAPIResponse(t, listExecRec).Data.(map[string]any)["items"].([]any)
	if len(items) == 0 {
		t.Fatalf("expected executions from pg-backed handler")
	}

	var failedExecutionID string
	var failedTodoID string
	for _, raw := range items {
		item := raw.(map[string]any)
		if item["status"] == "failed" {
			failedExecutionID = item["execution_id"].(string)
			failedTodoID = item["todo_id"].(string)
			break
		}
	}
	if failedExecutionID == "" {
		t.Fatalf("expected failed execution from pg-backed flow")
	}

	retryRec := httptest.NewRecorder()
	retryReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/executions/retry", body(`{
		"items":[{"todo_id":"`+failedTodoID+`","latest_failed_execution_id":"`+failedExecutionID+`"}],
		"reason":"pg-backed retry"
	}`))
	mux.ServeHTTP(retryRec, retryReq)
	if retryRec.Code != http.StatusOK {
		t.Fatalf("retry execution failed: %d %s", retryRec.Code, retryRec.Body.String())
	}
	retryItem := decodeAPIResponse(t, retryRec).Data.(map[string]any)["items"].([]any)[0].(map[string]any)
	newExecutionID := retryItem["new_execution_id"].(string)
	if retryItem["new_attempt"] != float64(2) {
		t.Fatalf("expected retry attempt 2, got %v", retryItem["new_attempt"])
	}

	execDetailRec := httptest.NewRecorder()
	execDetailReq := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+sessionID+"/executions/"+newExecutionID, nil)
	mux.ServeHTTP(execDetailRec, execDetailReq)
	if execDetailRec.Code != http.StatusOK {
		t.Fatalf("get retried execution detail failed: %d %s", execDetailRec.Code, execDetailRec.Body.String())
	}
	execDetail := decodeAPIResponse(t, execDetailRec).Data.(map[string]any)
	if execDetail["retry_of_execution_id"] != failedExecutionID {
		t.Fatalf("expected retry_of_execution_id %s, got %v", failedExecutionID, execDetail["retry_of_execution_id"])
	}
	if execDetail["todo_version"] != float64(2) {
		t.Fatalf("expected retried execution todo_version 2, got %v", execDetail["todo_version"])
	}

	logsRec := httptest.NewRecorder()
	logsReq := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+sessionID+"/executions/"+failedExecutionID+"/logs?limit=100&offset=0", nil)
	mux.ServeHTTP(logsRec, logsReq)
	if logsRec.Code != http.StatusOK {
		t.Fatalf("get failed execution logs failed: %d %s", logsRec.Code, logsRec.Body.String())
	}
	logsData := decodeAPIResponse(t, logsRec).Data.(map[string]any)
	if logsData["execution_id"] != failedExecutionID {
		t.Fatalf("expected execution logs id %s, got %v", failedExecutionID, logsData["execution_id"])
	}
	if len(logsData["items"].([]any)) == 0 {
		t.Fatalf("expected pg-backed execution logs")
	}
}
