package orchestrator

import (
	"context"
	"fmt"
	"sort"
	"sync/atomic"

	"juyu-ai-platform/internal/registry"
	"juyu-ai-platform/internal/router"
	"juyu-ai-platform/internal/store"
	"juyu-ai-platform/internal/types"
)

type Orchestrator struct {
	router   *router.Router
	registry *registry.Registry
	store    *store.MemoryStore
	bindings map[string]types.Binding
	counter  atomic.Uint64
}

func New(rt *router.Router, rg *registry.Registry, st *store.MemoryStore, bindingList []types.Binding) *Orchestrator {
	bindings := make(map[string]types.Binding)
	for _, b := range bindingList {
		bindings[b.MasterAgent] = b
	}
	return &Orchestrator{router: rt, registry: rg, store: st, bindings: bindings}
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

	taskID := fmt.Sprintf("task-%06d", o.counter.Add(1))
	task := &types.Task{
		ID:          taskID,
		Request:     req,
		MasterAgent: master.Code,
		Status:      types.TaskStatusRunning,
	}

	results, currentStep, needsConfirm, pendingResults, status, err := o.runBinding(ctx, req, binding, 0, nil)
	if err != nil {
		return types.ExecuteResponse{}, err
	}

	task.Results = results
	task.PendingResults = pendingResults
	task.CurrentStep = currentStep
	task.NeedsConfirm = needsConfirm
	task.Status = status
	o.store.Save(task)

	return types.ExecuteResponse{
		TaskID:      task.ID,
		MasterAgent: master.Code,
		Scene:       req.Scene,
		Status:      task.Status,
		Results:     task.Results,
	}, nil
}

func (o *Orchestrator) Confirm(ctx context.Context, taskID string, approved bool) (types.Task, error) {
	task, err := o.store.Get(taskID)
	if err != nil {
		return types.Task{}, err
	}
	if !task.NeedsConfirm {
		return *task, nil
	}
	if !approved {
		task.Status = types.TaskStatusRejected
		task.NeedsConfirm = false
		task.PendingResults = nil
		o.store.Save(task)
		return *task, nil
	}

	binding, ok := o.bindings[task.MasterAgent]
	if !ok {
		return types.Task{}, fmt.Errorf("no binding for master agent: %s", task.MasterAgent)
	}

	results, currentStep, needsConfirm, pendingResults, status, err := o.runBinding(ctx, task.Request, binding, task.CurrentStep, task.Results)
	if err != nil {
		return types.Task{}, err
	}
	task.Results = results
	task.PendingResults = pendingResults
	task.CurrentStep = currentStep
	task.NeedsConfirm = needsConfirm
	task.Status = status
	o.store.Save(task)
	return *task, nil
}

func (o *Orchestrator) GetTask(taskID string) (types.Task, error) {
	t, err := o.store.Get(taskID)
	if err != nil {
		return types.Task{}, err
	}
	return *t, nil
}

func (o *Orchestrator) runBinding(ctx context.Context, req types.Request, binding types.Binding, startStep int, existing []types.Response) ([]types.Response, int, bool, []types.Response, string, error) {
	results := append([]types.Response{}, existing...)

	if len(binding.Workflow) > 0 {
		workflow := append([]types.WorkflowStep{}, binding.Workflow...)
		sort.Slice(workflow, func(i, j int) bool { return workflow[i].Step < workflow[j].Step })
		for _, step := range workflow {
			if step.Step <= startStep {
				continue
			}
			agent, err := o.registry.Get(step.Ability)
			if err != nil {
				return nil, 0, false, nil, "", err
			}
			resp, err := agent.Run(ctx, req)
			if err != nil {
				return nil, 0, false, nil, "", fmt.Errorf("run %s failed: %w", step.Ability, err)
			}
			if step.RequireHumanConfirm {
				if resp.Data == nil {
					resp.Data = map[string]any{}
				}
				resp.Data["require_human_confirm"] = true
				results = append(results, resp)
				return results, step.Step, true, []types.Response{resp}, types.TaskStatusPendingConfirm, nil
			}
			results = append(results, resp)
		}
		return results, len(workflow), false, nil, types.TaskStatusSuccess, nil
	}

	for _, code := range binding.Abilities {
		agent, err := o.registry.Get(code)
		if err != nil {
			return nil, 0, false, nil, "", err
		}
		resp, err := agent.Run(ctx, req)
		if err != nil {
			return nil, 0, false, nil, "", fmt.Errorf("run %s failed: %w", code, err)
		}
		results = append(results, resp)
	}
	return results, len(binding.Abilities), false, nil, types.TaskStatusSuccess, nil
}
