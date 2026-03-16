package adapters

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"juyu-ai-platform/internal/types"
)

func TestAlibabaAdapterLiveCall(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/publish" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("X-Request-ID") != "req-1" {
			t.Fatalf("missing request id header")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "published", "remote_id": "abc-1"})
	}))
	defer ts.Close()

	store := NewCredentialStore()
	store.Set(types.AdapterCredentials{Platform: "alibaba", Fields: map[string]string{"app_key": "k", "secret": "s"}})
	adapter := &AlibabaAdapter{BaseAdapter: NewBaseAdapter("alibaba", "Alibaba", []string{"app_key", "secret"}, store, false, ts.URL)}

	resp, err := adapter.Publish(context.Background(), types.AdapterRequest{
		Platform:    "alibaba",
		Action:      "publish",
		RequestID:   "req-1",
		ExternalRef: "sku-1",
		Payload: map[string]any{
			"product": types.Product{Title: "demo"},
		},
	})
	if err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	if resp.Mode != "live" || resp.Status != "published" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
