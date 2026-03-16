package adapters

import "juyu-ai-platform/internal/types"

func attachTraceData(resp *types.AdapterResponse, trace AdapterTrace, baseURL, externalRef string) {
	resp.Data["base_url"] = baseURL
	if externalRef != "" {
		resp.Data["external_ref"] = externalRef
	}
	resp.Data["trace"] = trace
	resp.Data["request"] = map[string]any{
		"url":          trace.URL,
		"content_type": trace.ContentType,
		"headers":      trace.Headers,
		"body":         trace.Body,
		"attempts":     trace.Attempts,
	}
}
