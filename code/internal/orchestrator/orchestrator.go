package orchestrator

import (
	"context"
	"fmt"

	"juyu-ai-platform/internal/registry"
	"juyu-ai-platform/internal/router"
	"juyu-ai-platform/internal/types"
)

type Orchestrator struct {
	router   *router.Router
	registry *registry.Registry
	bindings map[string][]string
}

func New(rt *router.Router, rg *registry.Registry, bindingList []types.Binding) *Orchestrator {
	bindings := make(map[string][]string)
	for _, b := range bindingList {
		bindings[b.MasterAgent] = b.Abilities
	}
	return &Orchestrator{router: rt, registry: rg, bindings: bindings}
}

func (o *Orchestrator) Execute(ctx context.Context, req types.Request) ([]types.Response, error) {
	master, ok := o.router.Route(req.Scene)
	if !ok {
		return nil, fmt.Errorf("no master agent for scene: %s", req.Scene)
	}

	abilityCodes := o.bindings[master.Code]
	if len(abilityCodes) == 0 {
		return nil, fmt.Errorf("no abilities bound to master agent: %s", master.Code)
	}

	results := make([]types.Response, 0, len(abilityCodes))
	for _, code := range abilityCodes {
		agent, err := o.registry.Get(code)
		if err != nil {
			return nil, err
		}
		resp, err := agent.Run(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("run %s failed: %w", code, err)
		}
		results = append(results, resp)
	}

	return results, nil
}
