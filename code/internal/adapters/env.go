package adapters

import (
	"strings"

	"juyu-ai-platform/internal/config"
	"juyu-ai-platform/internal/types"
)

func LoadCredentialsFromEnv() *CredentialStore {
	store := NewCredentialStore()
	store.Set(types.AdapterCredentials{Platform: "alibaba", Fields: map[string]string{
		"app_key": config.GetEnv("JUYU_ALIBABA_APP_KEY", ""),
		"secret":  config.GetEnv("JUYU_ALIBABA_SECRET", ""),
	}})
	store.Set(types.AdapterCredentials{Platform: "taobao", Fields: map[string]string{
		"app_key": config.GetEnv("JUYU_TAOBAO_APP_KEY", ""),
		"secret":  config.GetEnv("JUYU_TAOBAO_SECRET", ""),
	}})
	store.Set(types.AdapterCredentials{Platform: "douyin", Fields: map[string]string{
		"client_id":     config.GetEnv("JUYU_DOUYIN_CLIENT_ID", ""),
		"client_secret": config.GetEnv("JUYU_DOUYIN_CLIENT_SECRET", ""),
	}})
	return store
}

func DryRunFromEnv() bool {
	v := strings.TrimSpace(strings.ToLower(config.GetEnv("JUYU_ADAPTER_DRY_RUN", "true")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}
