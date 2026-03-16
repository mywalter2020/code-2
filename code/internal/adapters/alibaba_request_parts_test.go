package adapters

import (
	"testing"

	"juyu-ai-platform/internal/types"
)

func TestAlibabaBuildRequestPartsQueryAuth(t *testing.T) {
	t.Setenv("JUYU_ALIBABA_AUTH_PLACEMENT", "query")
	t.Setenv("JUYU_ALIBABA_BASE_URL", "https://api.example.com")
	adapter := &AlibabaAdapter{BaseAdapter: NewBaseAdapter("alibaba", "Alibaba", []string{"app_key", "secret"}, NewCredentialStore(), false, "https://api.example.com")}
	parts := adapter.buildRequestParts("publish", types.AdapterRequest{RequestID: "req-1"}, map[string]any{
		"app_key": "k",
		"method":  "m",
		"sign":    "sig",
		"item":    map[string]any{"title": "demo"},
	})
	if parts.URL != "https://api.example.com/publish?app_key=k&method=m&sign=sig" {
		t.Fatalf("unexpected url: %s", parts.URL)
	}
}

func TestAlibabaBuildRequestPartsHeaderAuth(t *testing.T) {
	t.Setenv("JUYU_ALIBABA_AUTH_PLACEMENT", "header")
	adapter := &AlibabaAdapter{BaseAdapter: NewBaseAdapter("alibaba", "Alibaba", []string{"app_key", "secret"}, NewCredentialStore(), false, "https://api.example.com")}
	body := map[string]any{"app_key": "k", "method": "m", "sign": "sig", "item": map[string]any{"title": "demo"}}
	parts := adapter.buildRequestParts("publish", types.AdapterRequest{RequestID: "req-1"}, body)
	if parts.Headers["X-Alibaba-Sign"] != "sig" || parts.Headers["X-Alibaba-App-Key"] != "k" {
		t.Fatalf("unexpected headers: %+v", parts.Headers)
	}
	if _, ok := body["sign"]; ok {
		t.Fatalf("expected sign removed from body: %+v", body)
	}
}
