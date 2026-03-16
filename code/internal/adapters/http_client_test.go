package adapters

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAdapterHTTPClientFormMode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); !strings.Contains(got, "application/x-www-form-urlencoded") {
			t.Fatalf("unexpected content-type: %s", got)
		}
		body, _ := io.ReadAll(r.Body)
		raw := string(body)
		if !strings.Contains(raw, "app_key=demo") || !strings.Contains(raw, "item.title=chair") {
			t.Fatalf("unexpected form body: %s", raw)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	client := NewAdapterHTTPClient(nil)
	_, err := client.DoJSON(context.Background(), AdapterHTTPRequest{
		Method:      http.MethodPost,
		URL:         ts.URL,
		ContentType: "application/x-www-form-urlencoded",
		Body: map[string]any{
			"app_key": "demo",
			"item": map[string]any{
				"title": "chair",
			},
		},
	})
	if err != nil {
		t.Fatalf("form request failed: %v", err)
	}
}
