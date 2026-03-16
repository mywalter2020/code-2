package adapters

import "juyu-ai-platform/internal/types"

func paramsWithItem(params map[string]string, req types.AdapterRequest, item AlibabaItemPayload) map[string]any {
	out := map[string]any{
		"method":       params["method"],
		"sign_method":  params["sign_method"],
		"timestamp":    params["timestamp"],
		"sign":         params["sign"],
		"request_id":   req.RequestID,
		"external_ref": req.ExternalRef,
		"item":         item,
	}
	if params["app_key"] != "" {
		out["app_key"] = params["app_key"]
	}
	if params["v"] != "" {
		out["v"] = params["v"]
	}
	return out
}
