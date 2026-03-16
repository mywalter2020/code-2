package adapters

import (
	"testing"

	"juyu-ai-platform/internal/types"
)

func TestAlibabaItemPayloadMapping(t *testing.T) {
	payload, err := alibabaItemPayload(types.AdapterRequest{
		ExternalRef: "ext-1",
		Payload: map[string]any{
			"product": types.Product{
				Title:       "demo title",
				Subtitle:    "sub",
				Description: "desc",
				Category:    "cat",
				Brand:       "brand",
				SKU:         "sku-1",
				Price:       99.5,
				Currency:    "CNY",
				Images:      []string{"a.jpg"},
				Tags:        []string{"hot"},
			},
			"publish": types.PublishRequest{
				Channel:     "retail",
				Content:     "content body",
				PublishMode: "manual_confirm",
				Page:        map[string]any{"title": "page"},
				Metadata:    map[string]any{"k": "v"},
			},
		},
	})
	if err != nil {
		t.Fatalf("mapping error: %v", err)
	}
	if payload.Title != "demo title" || payload.OutBizID != "ext-1" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if payload.Metadata["k"] != "v" || payload.Metadata["external_ref"] != "ext-1" {
		t.Fatalf("unexpected metadata: %+v", payload.Metadata)
	}
}

func TestParseAlibabaResponse(t *testing.T) {
	resp := parseAlibabaResponse(AdapterHTTPResponse{StatusCode: 200, JSON: map[string]any{
		"response": map[string]any{
			"status":      "published",
			"remote_id":   "ali-123",
			"sub_msg":     "ok",
			"extra_field": true,
		},
	}}, "publish", "req-1")
	if resp.Status != "published" {
		t.Fatalf("expected published, got %+v", resp)
	}
	if resp.Data["remote_id"] != "ali-123" {
		t.Fatalf("expected remote id, got %+v", resp.Data)
	}
}

func TestParseAlibabaResponseWithErrorCode(t *testing.T) {
	resp := parseAlibabaResponse(AdapterHTTPResponse{StatusCode: 200, JSON: map[string]any{
		"data": map[string]any{
			"error_code":    "INVALID_PARAM",
			"error_message": "bad title",
		},
	}}, "publish", "req-1")
	if resp.Data["error_code"] != "INVALID_PARAM" {
		t.Fatalf("expected error code, got %+v", resp.Data)
	}
	if resp.Data["error_class"] != "business" {
		t.Fatalf("expected business class, got %+v", resp.Data)
	}
	if resp.Status != "failed" {
		t.Fatalf("expected failed status, got %+v", resp)
	}
}
