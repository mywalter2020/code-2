package server

import "juyu-ai-platform/internal/llm"

func runtimeInfo() map[string]any {
	cfg := llm.LoadContentGenConfig()
	return map[string]any{
		"content_gen": map[string]any{
			"provider":    cfg.Provider,
			"enabled":     cfg.Enabled,
			"model":       cfg.Model,
			"temperature": cfg.Temperature,
			"max_tokens":  cfg.MaxTokens,
		},
		"providers": []string{"stub", "nvidia"},
	}
}
