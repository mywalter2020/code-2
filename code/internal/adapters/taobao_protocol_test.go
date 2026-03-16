package adapters

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"juyu-ai-platform/internal/types"
)

func TestTaobaoLiveInvokeSendsSignedEnvelope(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["sign"] == "" || body["method"] != "taobao.item.publish" {
			t.Fatalf("unexpected body: %+v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "published", "remote_id": "tb-1"})
	}))
	defer ts.Close()
	store := NewCredentialStore()
	store.Set(types.AdapterCredentials{Platform: "taobao", Fields: map[string]string{"app_key": "demo-key", "secret": "demo-secret"}})
	adapter := &TaobaoAdapter{BaseAdapter: NewBaseAdapter("taobao", "Taobao", []string{"app_key", "secret"}, store, false, ts.URL)}
	resp, err := adapter.Publish(context.Background(), types.AdapterRequest{Platform: "taobao", Action: "publish", RequestID: "req-1", Payload: map[string]any{"product": types.Product{Title: "demo"}}})
	if err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	if resp.Platform != "taobao" || resp.Status != "published" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if _, ok := resp.Data["trace"]; !ok {
		t.Fatalf("expected trace in response: %+v", resp.Data)
	}
}
