package agents

import "strings"

func stringFromPayload(payload map[string]any, key, fallback string) string {
	if payload == nil {
		return fallback
	}
	if v, ok := payload[key].(string); ok && strings.TrimSpace(v) != "" {
		return v
	}
	return fallback
}
