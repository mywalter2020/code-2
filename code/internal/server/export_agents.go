package server

import "juyu-ai-platform/internal/types"

func BuildMasterMetadata(items []types.MasterAgent) []types.MasterAgentMetadata {
	return buildMasterMetadata(items)
}

func BuildBindingViews(items []types.Binding) []types.BindingView {
	result := make([]types.BindingView, 0, len(items))
	for _, item := range items {
		result = append(result, types.BindingView{
			MasterAgent: item.MasterAgent,
			Abilities:   item.Abilities,
			Workflow:    item.Workflow,
		})
	}
	return result
}
