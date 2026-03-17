package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"juyu-ai-platform/internal/types"
)

func TestTaskEndpointsAfterExecute(t *testing.T) {
	srv := newTestServer(t, "secret", "")
	mux := http.NewServeMux()
	srv.Register(mux)

	taskID := createTaskForTest(t, mux)

	for _, path := range []string{
		"/tasks/" + taskID,
		"/tasks/" + taskID + "/preview",
		"/tasks/" + taskID + "/logs",
		"/tasks/" + taskID + "/status",
	} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for %s, got %d", path, rec.Code)
		}
	}
}

func TestTaskConfirmFlow(t *testing.T) {
	srv := newTestServer(t, "secret", "")
	mux := http.NewServeMux()
	srv.Register(mux)

	taskID := createTaskForTest(t, mux)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/tasks/"+taskID+"/confirm", body(`{"approved":true,"comment":"ok","approver":"tester"}`))
	req.Header.Set("X-API-Key", "secret")
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	resp := decodeAPIResponse(t, rec)
	data := resp.Data.(map[string]any)
	if data["status"] != types.TaskStatusSuccess {
		t.Fatalf("expected task success after confirm, got %v", data["status"])
	}
}

func TestTaskConfirmUsesOperatorIdentityHeader(t *testing.T) {
	srv := newTestServer(t, "secret", "reviewer:approve-token")
	mux := http.NewServeMux()
	srv.Register(mux)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/execute", body(`{"scene":"product","input":"demo","payload":{"platform":"alibaba"}}`))
	req.Header.Set("X-API-Key", "secret")
	req.Header.Set("X-Operator-ID", "reviewer")
	req.Header.Set("X-Operator-Token", "approve-token")
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("execute failed: %d %s", rec.Code, rec.Body.String())
	}
	taskID := decodeAPIResponse(t, rec).Data.(map[string]any)["task_id"].(string)

	confirmRec := httptest.NewRecorder()
	confirmReq := httptest.NewRequest(http.MethodPost, "/tasks/"+taskID+"/confirm", body(`{"approved":true,"comment":"ship it"}`))
	confirmReq.Header.Set("X-API-Key", "secret")
	confirmReq.Header.Set("X-Operator-ID", "reviewer")
	confirmReq.Header.Set("X-Operator-Token", "approve-token")
	mux.ServeHTTP(confirmRec, confirmReq)

	if confirmRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", confirmRec.Code, confirmRec.Body.String())
	}
	data := decodeAPIResponse(t, confirmRec).Data.(map[string]any)
	if data["approver"] != "reviewer" {
		t.Fatalf("expected approver reviewer, got %v", data["approver"])
	}
}
