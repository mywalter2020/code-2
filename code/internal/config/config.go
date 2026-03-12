package config

import (
	"os"

	"gopkg.in/yaml.v3"
	"juyu-ai-platform/internal/types"
)

func Load(path string) (types.Config, error) {
	var cfg types.Config
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
