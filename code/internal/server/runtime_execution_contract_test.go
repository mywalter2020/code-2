package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRuntimeExecutionListAndDetailContractsAfterRetry(t *testing.T) {
	srv := newTestServer(t, "", "")
	mux := http.NewServeMux()
	srv.Register(mux)

	createRec := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", body(`{
		"input": {
			"type": "product_url_learning",
			"url": "https://example.com/product/execution",
			"message": "execution contract"
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
		"comment":"bump todo version before execution"
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

	listRec := httptest.NewRecorder()
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+sessionID+"/executions", nil)
	mux.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list executions failed: %d %s", listRec.Code, listRec.Body.String())
	}
	listResp := decodeAPIResponse(t, listRec)
	listData := listResp.Data.(map[string]any)
	items := listData["items"].([]any)
	if len(items) < 3 {
		t.Fatalf("expected at least 3 executions, got %d", len(items))
	}

	var failedExecutionID string
	var failedTodoID string
	for _, raw := range items {
		item := raw.(map[string]any)
		if item["status"] == "failed" {
			failedExecutionID = item["execution_id"].(string)
			failedTodoID = item["todo_id"].(string)
			if item["attempt"] != float64(1) {
				t.Fatalf("expected failed execution attempt 1, got %v", item["attempt"])
			}
			break
		}
	}
	if failedExecutionID == "" {
		t.Fatalf("expected one failed execution")
	}

	detailRec := httptest.NewRecorder()
	detailReq := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+sessionID+"/executions/"+failedExecutionID, nil)
	mux.ServeHTTP(detailRec, detailReq)
	if detailRec.Code != http.StatusOK {
		t.Fatalf("execution detail failed: %d %s", detailRec.Code, detailRec.Body.String())
	}
	detailResp := decodeAPIResponse(t, detailRec)
	detail := detailResp.Data.(map[string]any)
	if detail["execution_id"] != failedExecutionID {
		t.Fatalf("expected execution id %s, got %v", failedExecutionID, detail["execution_id"])
	}
	if detail["todo_id"] != failedTodoID {
		t.Fatalf("expected todo id %s, got %v", failedTodoID, detail["todo_id"])
	}
	if detail["todo_version"] != float64(2) {
		t.Fatalf("expected todo_version 2 after todo edit, got %v", detail["todo_version"])
	}
	if detail["queue_reason"] != "todo_confirmed" {
		t.Fatalf("expected queue_reason todo_confirmed, got %v", detail["queue_reason"])
	}
	if detail["error_code"] != "MODEL_TIMEOUT" {
		t.Fatalf("expected MODEL_TIMEOUT, got %v", detail["error_code"])
	}
	logs := detail["logs"].([]any)
	if len(logs) == 0 {
		t.Fatalf("expected execution detail logs")
	}
	firstLog := logs[0].(map[string]any)
	if firstLog["execution_id"] != failedExecutionID {
		t.Fatalf("expected detail log execution id %s, got %v", failedExecutionID, firstLog["execution_id"])
	}
	if firstLog["todo_id"] != failedTodoID {
		t.Fatalf("expected detail log todo id %s, got %v", failedTodoID, firstLog["todo_id"])
	}

	retryRec := httptest.NewRecorder()
	retryReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/executions/retry", body(`{
		"items":[{"todo_id":"`+failedTodoID+`","latest_failed_execution_id":"`+failedExecutionID+`"}],
		"reason":"retry contract"
	}`))
	mux.ServeHTTP(retryRec, retryReq)
	if retryRec.Code != http.StatusOK {
		t.Fatalf("retry execution failed: %d %s", retryRec.Code, retryRec.Body.String())
	}
	retryResp := decodeAPIResponse(t, retryRec)
	retryData := retryResp.Data.(map[string]any)
	retryItems := retryData["items"].([]any)
	if len(retryItems) != 1 {
		t.Fatalf("expected one retry item, got %d", len(retryItems))
	}
	retryItem := retryItems[0].(map[string]any)
	newExecutionID := retryItem["new_execution_id"].(string)
	if retryItem["previous_execution_id"] != failedExecutionID {
		t.Fatalf("expected previous execution id %s, got %v", failedExecutionID, retryItem["previous_execution_id"])
	}
	if retryItem["new_attempt"] != float64(2) {
		t.Fatalf("expected new attempt 2, got %v", retryItem["new_attempt"])
	}

	listAfterRetryRec := httptest.NewRecorder()
	listAfterRetryReq := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+sessionID+"/executions", nil)
	mux.ServeHTTP(listAfterRetryRec, listAfterRetryReq)
	if listAfterRetryRec.Code != http.StatusOK {
		t.Fatalf("list executions after retry failed: %d %s", listAfterRetryRec.Code, listAfterRetryRec.Body.String())
	}
	listAfterRetryResp := decodeAPIResponse(t, listAfterRetryRec)
	afterItems := listAfterRetryResp.Data.(map[string]any)["items"].([]any)
	foundRetry := false
	for _, raw := range afterItems {
		item := raw.(map[string]any)
		if item["execution_id"] == newExecutionID {
			foundRetry = true
			if item["attempt"] != float64(2) {
				t.Fatalf("expected retry execution attempt 2, got %v", item["attempt"])
			}
			if item["retry_of_execution_id"] != failedExecutionID {
				t.Fatalf("expected retry_of_execution_id %s, got %v", failedExecutionID, item["retry_of_execution_id"])
			}
		}
	}
	if !foundRetry {
		t.Fatalf("expected retried execution %s in execution list", newExecutionID)
	}
}

func TestRuntimeExecutionLogsFilteringContracts(t *testing.T) {
	srv := newTestServer(t, "", "")
	mux := http.NewServeMux()
	srv.Register(mux)

	createRec := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", body(`{
		"input": {
			"type": "product_url_learning",
			"url": "https://example.com/product/logs",
			"message": "logs filter"
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

	errorLogsRec := httptest.NewRecorder()
	errorLogsReq := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+sessionID+"/logs?level=error", nil)
	mux.ServeHTTP(errorLogsRec, errorLogsReq)
	if errorLogsRec.Code != http.StatusOK {
		t.Fatalf("session error logs failed: %d %s", errorLogsRec.Code, errorLogsRec.Body.String())
	}
	errorLogsResp := decodeAPIResponse(t, errorLogsRec)
	errorItems := errorLogsResp.Data.(map[string]any)["items"].([]any)
	if len(errorItems) == 0 {
		t.Fatalf("expected error logs")
	}
	logItem := errorItems[0].(map[string]any)
	if logItem["level"] != "error" {
		t.Fatalf("expected error log level, got %v", logItem["level"])
	}
	if logItem["event_type"] != "execution.failed" {
		t.Fatalf("expected execution.failed event type, got %v", logItem["event_type"])
	}
	failedExecutionID := logItem["execution_id"].(string)
	failedTodoID := logItem["todo_id"].(string)

	executionLogsRec := httptest.NewRecorder()
	executionLogsReq := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+sessionID+"/executions/"+failedExecutionID+"/logs?limit=100&offset=0", nil)
	mux.ServeHTTP(executionLogsRec, executionLogsReq)
	if executionLogsRec.Code != http.StatusOK {
		t.Fatalf("execution logs failed: %d %s", executionLogsRec.Code, executionLogsRec.Body.String())
	}
	executionLogsResp := decodeAPIResponse(t, executionLogsRec)
	executionData := executionLogsResp.Data.(map[string]any)
	if executionData["execution_id"] != failedExecutionID {
		t.Fatalf("expected execution_id %s, got %v", failedExecutionID, executionData["execution_id"])
	}
	executionItems := executionData["items"].([]any)
	if len(executionItems) == 0 {
		t.Fatalf("expected execution logs items")
	}
	for _, raw := range executionItems {
		item := raw.(map[string]any)
		if item["execution_id"] != failedExecutionID {
			t.Fatalf("expected filtered execution_id %s, got %v", failedExecutionID, item["execution_id"])
		}
		if item["todo_id"] != failedTodoID {
			t.Fatalf("expected filtered todo_id %s, got %v", failedTodoID, item["todo_id"])
		}
	}
}
