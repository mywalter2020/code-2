package adapters

import "testing"

func TestTaobaoErrorClassificationUsesSharedParser(t *testing.T) {
	resp := parseAlibabaResponse(AdapterHTTPResponse{StatusCode: 429, JSON: map[string]any{
		"data": map[string]any{
			"error_code":    "LIMIT_EXCEEDED",
			"error_message": "slow down",
		},
	}}, "publish", "req-1")
	if resp.Data["error_class"] != "rate_limit" {
		t.Fatalf("expected rate_limit, got %+v", resp.Data)
	}
}
