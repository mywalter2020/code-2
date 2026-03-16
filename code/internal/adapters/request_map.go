package adapters

import "juyu-ai-platform/internal/types"

func requestPayload(req types.AdapterRequest) map[string]any {
	return map[string]any{
		"platform":     req.Platform,
		"action":       req.Action,
		"request_id":   req.RequestID,
		"external_ref": req.ExternalRef,
		"operator":     req.Operator,
		"payload":      req.Payload,
	}
}
