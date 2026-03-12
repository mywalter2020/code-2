package orchestrator

import (
	"context"
	"fmt"
	"sort"

	"juyu-ai-platform/internal/registry"
	"juyu-ai-platform/internal/router"
	"juyu-ai-platform/internal/types"
)

type Orchestrator struct {
	router    *router.Router
	registry  *registry.Registry
	bindings  map[string]types.Binding
}

func New(rt *router.Router, rg *registry.Registry, bindingList []types.Binding) *Orchestrator {
	bindings := make(map[string]types.Binding)
	for _, b := range bindingList {
		bindings[b.MasterAgent] = b
	}
	return &Orchestrator{router: rt, registry: rg, bindings: bindings}
}

func (o *Orchestrator) Execute(ctx context.Context, req types.Request) (types.ExecuteResponse, error) {
	master, ok := o.router.Route(req.Scene)
	if !ok {
		return types.ExecuteResponse{}, fmt.Errorf("no master agent for scene: %s", req.Scene)
	}

	binding, ok := o.bindings[master.Code]
	if !ok {
		return types.ExecuteResponse{}, fmt.Errorf("no binding for master agent: %s", master.Code)
	}

	results := make([]types.Response, 0)

	if len(binding.Workflow) > 0 {
		sort.Slice(binding.Workflow, func(i, j int) bool {
			return binding.Workflow[i].Step < binding.Workflow[j].Step
		})
		for _, step := range binding.Workflow {
			agent, err := o.registry.Get(step.Ability)
			if err != nil {
				return types.ExecuteResponse{}, err
			}
			resp, err := agent.Run(ctx, req)
			if err != nil {
				return types.ExecuteResponse{}, fmt.Errorf("run %s failed: %w", step.Ability, err)
			}
			if step.RequireHumanConfirm {
				if resp.Data == nil {
					resp.Data = map[string]any{}
				}
				resp.Data["require_human_confirm"] = true
			}
			results = append(results, resp)
		}
	} else {
		for _, code := range binding.Abilities {
			agent, err := o.registry.Get(code)
			if err != nil {
				return types.ExecuteResponse{}, err
			}
			resp, err := agent.Run(ctx, req)
			if err != nil {
				return types.ExecuteResponse{}, fmt.Errorf("run %s failed: %w", code, err)
			}
			results = append(results, resp)
		}
	}

	return types.ExecuteResponse{
		MasterAgent: master.Code,
		Scene:       req.Scene,
		Results:     results,
	}, nil
}
