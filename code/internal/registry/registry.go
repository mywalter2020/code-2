package registry

import (
	"fmt"

	"juyu-ai-platform/internal/types"
)

type Registry struct {
	agents map[string]types.AbilityAgent
}

func New() *Registry {
	return &Registry{agents: make(map[string]types.AbilityAgent)}
}

func (r *Registry) Register(agent types.AbilityAgent) {
	r.agents[agent.Code()] = agent
}

func (r *Registry) Get(code string) (types.AbilityAgent, error) {
	agent, ok := r.agents[code]
	if !ok {
		return nil, fmt.Errorf("ability agent not found: %s", code)
	}
	return agent, nil
}
