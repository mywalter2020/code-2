package server

import "juyu-ai-platform/internal/types"

func BuildAbilityMetadata(items []types.AbilityConfig) []types.AbilityMetadata {
	return buildAbilityMetadata(items)
}
