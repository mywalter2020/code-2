package adapters

import "strings"

func redactMap(m map[string]any, keys ...string) map[string]any {
	if m == nil {
		return nil
	}
	redacted := map[string]any{}
	sensitive := map[string]struct{}{}
	for _, k := range keys {
		sensitive[strings.ToLower(strings.TrimSpace(k))] = struct{}{}
	}
	for k, v := range m {
		if _, ok := sensitive[strings.ToLower(k)]; ok {
			redacted[k] = "***REDACTED***"
			continue
		}
		switch child := v.(type) {
		case map[string]any:
			redacted[k] = redactMap(child, keys...)
		default:
			redacted[k] = v
		}
	}
	return redacted
}
