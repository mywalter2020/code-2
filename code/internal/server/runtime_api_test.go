package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRuntimeAPIPrdVersionIncrementsAfterEdit(t *testing.T) {
	srv := newTestServer(t, "", "")
	mux := http.NewServeMux()
	srv.Register(mux)

	createRec := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", body(`{
		"input": {
			"type": "product_url_learning",
			"url": "https://example.com/product/1",
			"message": "learn this product"
		}
	}`))
	mux.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("create session failed: %d %s", createRec.Code, createRec.Body.String())
	}
	createResp := decodeAPIResponse(t, createRec)
	createData := createResp.Data.(map[string]any)
	sessionID := createData["session_id"].(string)
	artifacts := createData["artifacts"].(map[string]any)
	prd := artifacts["prd"].(map[string]any)
	if prd["version"] != float64(1) {
		t.Fatalf("expected initial prd version 1, got %v", prd["version"])
	}

	editRec := httptest.NewRecorder()
	editReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/edit-prd", body(`{
		"patch": {"title": "updated prd"},
		"comment": "bump version"
	}`))
	mux.ServeHTTP(editRec, editReq)
	if editRec.Code != http.StatusOK {
		t.Fatalf("edit prd failed: %d %s", editRec.Code, editRec.Body.String())
	}
	editResp := decodeAPIResponse(t, editRec)
	editData := editResp.Data.(map[string]any)
	editArtifacts := editData["artifacts"].(map[string]any)
	editedPrd := editArtifacts["prd"].(map[string]any)
	if editedPrd["version"] != float64(2) {
		t.Fatalf("expected edited prd version 2, got %v", editedPrd["version"])
	}
}

func TestRuntimeAPITodoVersionIncrementsAfterEdit(t *testing.T) {
	srv := newTestServer(t, "", "")
	mux := http.NewServeMux()
	srv.Register(mux)

	createRec := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", body(`{
		"input": {
			"type": "product_url_learning",
			"url": "https://example.com/product/2",
			"message": "learn this product"
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
	confirmPrdResp := decodeAPIResponse(t, confirmPrdRec)
	confirmPrdData := confirmPrdResp.Data.(map[string]any)
	confirmArtifacts := confirmPrdData["artifacts"].(map[string]any)
	todo := confirmArtifacts["todo"].(map[string]any)
	if todo["version"] != float64(1) {
		t.Fatalf("expected initial todo version 1, got %v", todo["version"])
	}
	items := todo["items"].([]any)
	if len(items) < 2 {
		t.Fatalf("expected todo items, got %d", len(items))
	}

	editTodoRec := httptest.NewRecorder()
	editTodoReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+sessionID+"/edit-todo", body(`{
		"items": [
			{"id":"todo_1","title":"生成产品标题","type":"title_generation","status":"pending","depends_on":[],"parallel_group":"content_assets"},
			{"id":"todo_2","title":"生成图片描述图内容","type":"feature_image_copy","status":"pending","depends_on":[],"parallel_group":"content_assets"}
		],
		"comment": "trim and bump"
	}`))
	mux.ServeHTTP(editTodoRec, editTodoReq)
	if editTodoRec.Code != http.StatusOK {
		t.Fatalf("edit todo failed: %d %s", editTodoRec.Code, editTodoRec.Body.String())
	}
	editTodoResp := decodeAPIResponse(t, editTodoRec)
	editTodoData := editTodoResp.Data.(map[string]any)
	editArtifacts := editTodoData["artifacts"].(map[string]any)
	editedTodo := editArtifacts["todo"].(map[string]any)
	if editedTodo["version"] != float64(2) {
		t.Fatalf("expected edited todo version 2, got %v", editedTodo["version"])
	}
}
