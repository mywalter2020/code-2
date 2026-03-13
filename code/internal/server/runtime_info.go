package server

import "juyu-ai-platform/internal/llm"

func runtimeInfo() map[string]any {
	cfg := llm.LoadContentGenConfig()
	pageCfg := llm.LoadPageGenConfig()
	return map[string]any{
		"content_gen": map[string]any{
			"provider":    cfg.Provider,
			"enabled":     cfg.Enabled,
			"model":       cfg.Model,
			"temperature": cfg.Temperature,
			"max_tokens":  cfg.MaxTokens,
		},
		"page_gen": map[string]any{
			"provider":    pageCfg.Provider,
			"enabled":     pageCfg.Enabled,
			"model":       pageCfg.Model,
			"temperature": pageCfg.Temperature,
			"max_tokens":  pageCfg.MaxTokens,
		},
		"providers": []string{"stub", "nvidia", "openai_compat"},
	}
}
