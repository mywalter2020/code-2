package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"juyu-ai-platform/internal/types"
)

func TestTaskEndpointsAfterExecute(t *testing.T) {
	srv := newTestServer(t, "secret")
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
	srv := newTestServer(t, "secret")
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
