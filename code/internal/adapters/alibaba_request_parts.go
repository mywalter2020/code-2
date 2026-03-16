package adapters

import (
	"strings"

	"juyu-ai-platform/internal/types"
)

func (a *AlibabaAdapter) buildRequestParts(action string, req types.AdapterRequest, body map[string]any) AdapterRequestParts {
	opts := a.liveOptions()
	parts := AdapterRequestParts{
		URL:         a.endpoint(a.livePath(action)),
		Headers:     map[string]string{},
		Body:        body,
		ContentType: "application/json",
	}
	if strings.EqualFold(opts.RequestFormat, "form") || strings.EqualFold(opts.RequestFormat, "x-www-form-urlencoded") {
		parts.ContentType = "application/x-www-form-urlencoded"
	}
	parts.Headers["X-Request-ID"] = req.RequestID
	parts.Headers["X-External-Ref"] = req.ExternalRef

	authPlacement := strings.ToLower(strings.TrimSpace(opts.AuthPlacement))
	if authPlacement == "header" {
		if sign, ok := body["sign"].(string); ok && sign != "" {
			parts.Headers["X-Alibaba-Sign"] = sign
		}
		if method, ok := body["method"].(string); ok && method != "" {
			parts.Headers["X-Alibaba-Method"] = method
		}
		if appKey, ok := body["app_key"].(string); ok && appKey != "" {
			parts.Headers["X-Alibaba-App-Key"] = appKey
			delete(body, "app_key")
		}
		delete(body, "sign")
		delete(body, "method")
	} else if authPlacement == "query" {
		query := map[string]string{}
		for _, key := range []string{"app_key", "method", "sign_method", "timestamp", "v", "sign"} {
			if s, ok := body[key].(string); ok && s != "" {
				query[key] = s
				delete(body, key)
			}
		}
		parts.URL = buildRequestURL(parts.URL, query)
	}
	return parts
}
