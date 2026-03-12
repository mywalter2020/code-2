package config

import "os"

func ResolveConfigPath() string {
	if p := os.Getenv("JUYU_CONFIG"); p != "" {
		return p
	}
	candidates := []string{
		"configs/agents.yaml",
		"../configs/agents.yaml",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "configs/agents.yaml"
}
