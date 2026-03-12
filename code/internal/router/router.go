package router

import "juyu-ai-platform/internal/types"

type Router struct {
	masters map[string]types.MasterAgent
}

func New(masters []types.MasterAgent) *Router {
	items := make(map[string]types.MasterAgent)
	for _, m := range masters {
		if m.Enabled {
			items[m.SceneType] = m
		}
	}
	return &Router{masters: items}
}

func (r *Router) Route(scene string) (types.MasterAgent, bool) {
	m, ok := r.masters[scene]
	return m, ok
}
