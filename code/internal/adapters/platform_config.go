package adapters

import "juyu-ai-platform/internal/config"

type PlatformConfig struct {
	BaseURL string
}

func LoadPlatformConfigsFromEnv() map[string]PlatformConfig {
	return map[string]PlatformConfig{
		"alibaba": {BaseURL: config.GetEnv("JUYU_ALIBABA_BASE_URL", "")},
		"taobao":  {BaseURL: config.GetEnv("JUYU_TAOBAO_BASE_URL", "")},
		"douyin":  {BaseURL: config.GetEnv("JUYU_DOUYIN_BASE_URL", "")},
	}
}
