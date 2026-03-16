package adapters

import "testing"

func TestRedactMap(t *testing.T) {
	out := redactMap(map[string]any{
		"sign":   "abc",
		"nested": map[string]any{"app_key": "demo"},
	}, "sign", "app_key")
	if out["sign"] != "***REDACTED***" {
		t.Fatalf("expected sign redacted: %+v", out)
	}
	nested := out["nested"].(map[string]any)
	if nested["app_key"] != "***REDACTED***" {
		t.Fatalf("expected nested app_key redacted: %+v", nested)
	}
}
