package adapters

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"juyu-ai-platform/internal/types"
)

func TestAlibabaMD5SignStable(t *testing.T) {
	sig := alibabaMD5Sign("secret", map[string]string{
		"app_key":   "demo",
		"method":    "alibaba.item.publish",
		"timestamp": "2026-03-16 10:00:00",
		"v":         "2.0",
		"payload":   `{"k":"v"}`,
	})
	if sig == "" {
		t.Fatal("expected non-empty signature")
	}
	if sig != alibabaMD5Sign("secret", map[string]string{
		"payload":   `{"k":"v"}`,
		"v":         "2.0",
		"timestamp": "2026-03-16 10:00:00",
		"method":    "alibaba.item.publish",
		"app_key":   "demo",
	}) {
		t.Fatal("expected stable signature regardless of param order")
	}
}

func TestAlibabaLiveInvokeSendsSignedEnvelope(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/publish" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["sign"] == "" || body["method"] != "alibaba.item.publish" {
			t.Fatalf("unexpected signed payload: %+v", body)
		}
		item, ok := body["item"].(map[string]any)
		if !ok || item["title"] != "demo" {
			t.Fatalf("expected mapped item payload, got %+v", body["item"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "published", "remote_id": "ali-1"})
	}))
	defer ts.Close()

	store := NewCredentialStore()
	store.Set(types.AdapterCredentials{Platform: "alibaba", Fields: map[string]string{"app_key": "demo-key", "secret": "demo-secret"}})
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
	if resp.Status != "published" || resp.Mode != "live" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
