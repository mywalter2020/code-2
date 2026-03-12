package registry

import (
	"fmt"
	"sort"

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

func (r *Registry) ListCodes() []string {
	codes := make([]string, 0, len(r.agents))
	for code := range r.agents {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes
}
