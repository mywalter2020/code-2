package orchestrator

import (
	"context"
	"fmt"
	"sort"
	"sync/atomic"
	"time"

	"juyu-ai-platform/internal/registry"
	"juyu-ai-platform/internal/router"
	"juyu-ai-platform/internal/store"
	"juyu-ai-platform/internal/types"
)

type Orchestrator struct {
	router   *router.Router
	registry *registry.Registry
	store    store.TaskStore
	bindings map[string]types.Binding
	counter  atomic.Uint64
}

func New(rt *router.Router, rg *registry.Registry, st store.TaskStore, bindingList []types.Binding) *Orchestrator {
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

	now := time.Now()
	taskID := fmt.Sprintf("task-%06d", o.counter.Add(1))
	task := &types.Task{
		ID:          taskID,
		Request:     req,
		MasterAgent: master.Code,
		Status:      types.TaskStatusRunning,
		Operator:    req.Operator,
		CreatedAt:   now,
		UpdatedAt:   now,
		Logs: []types.TaskLog{{
			Time:    now,
			Step:    0,
			Agent:   master.Code,
			Action:  "task_created",
			Message: "task created",
		}},
	}

	results, currentStep, needsConfirm, pendingResults, status, preview, logs, errMsg, err := o.runBinding(ctx, req, binding, 0, nil)
	task.Results = results
	task.PendingResults = pendingResults
	task.CurrentStep = currentStep
	task.NeedsConfirm = needsConfirm
	task.Status = status
	task.Preview = preview
	task.ErrorMessage = errMsg
	task.Logs = append(task.Logs, logs...)
	task.UpdatedAt = time.Now()
	_ = o.store.Save(task)

	if err != nil {
		return types.ExecuteResponse{
			TaskID:      task.ID,
			MasterAgent: master.Code,
			Scene:       req.Scene,
			Status:      task.Status,
			Preview:     task.Preview,
			Results:     task.Results,
		}, err
	}

	return types.ExecuteResponse{
		TaskID:      task.ID,
		MasterAgent: master.Code,
		Scene:       req.Scene,
		Status:      task.Status,
		Preview:     task.Preview,
		Results:     task.Results,
	}, nil
}

func (o *Orchestrator) Confirm(ctx context.Context, taskID string, approved bool, comment string, approver string) (types.Task, error) {
	task, err := o.store.Get(taskID)
	if err != nil {
		return types.Task{}, err
	}
	if !task.NeedsConfirm {
		return *task, nil
	}
	if !approved {
		now := time.Now()
		task.Status = types.TaskStatusRejected
		task.NeedsConfirm = false
		task.PendingResults = nil
		task.ConfirmComment = comment
		task.Approver = approver
		task.UpdatedAt = now
		task.LastConfirmedAt = &now
		task.Logs = append(task.Logs, types.TaskLog{
			Time:    now,
			Step:    task.CurrentStep,
			Agent:   task.MasterAgent,
			Action:  "task_rejected",
			Message: "task rejected by user",
			Data:    map[string]any{"comment": comment},
		})
		_ = o.store.Save(task)
		return *task, nil
	}

	binding, ok := o.bindings[task.MasterAgent]
	if !ok {
		return types.Task{}, fmt.Errorf("no binding for master agent: %s", task.MasterAgent)
	}

	results, currentStep, needsConfirm, pendingResults, status, preview, logs, errMsg, err := o.runBinding(ctx, task.Request, binding, task.CurrentStep, task.Results)
	now := time.Now()
	task.Results = results
	task.PendingResults = pendingResults
	task.CurrentStep = currentStep
	task.NeedsConfirm = needsConfirm
	task.Status = status
	task.Preview = preview
	task.ErrorMessage = errMsg
	task.ConfirmComment = comment
	task.Approver = approver
	task.UpdatedAt = now
	task.LastConfirmedAt = &now
	task.Logs = append(task.Logs, types.TaskLog{
		Time:    now,
		Step:    task.CurrentStep,
		Agent:   task.MasterAgent,
		Action:  "task_confirmed",
		Message: "task approved by user",
		Data:    map[string]any{"comment": comment},
	})
	task.Logs = append(task.Logs, logs...)
	_ = o.store.Save(task)
	if err != nil {
		return *task, err
	}
	return *task, nil
}

func (o *Orchestrator) GetTask(taskID string) (types.Task, error) {
	t, err := o.store.Get(taskID)
	if err != nil {
		return types.Task{}, err
	}
	return *t, nil
}

func (o *Orchestrator) ListTasks(filter types.TaskFilter) ([]types.Task, error) {
	return o.store.List(filter)
}

func (o *Orchestrator) Cancel(taskID string, comment string) (types.Task, error) {
	task, err := o.store.Get(taskID)
	if err != nil {
		return types.Task{}, err
	}
	task.Status = types.TaskStatusCanceled
	task.NeedsConfirm = false
	task.PendingResults = nil
	task.ConfirmComment = comment
	task.UpdatedAt = time.Now()
	task.Logs = append(task.Logs, types.TaskLog{Time: task.UpdatedAt, Step: task.CurrentStep, Agent: task.MasterAgent, Action: "task_canceled", Message: "task canceled by user", Data: map[string]any{"comment": comment}})
	_ = o.store.Save(task)
	return *task, nil
}

func (o *Orchestrator) Retry(ctx context.Context, taskID string) (types.Task, error) {
	task, err := o.store.Get(taskID)
	if err != nil {
		return types.Task{}, err
	}
	binding, ok := o.bindings[task.MasterAgent]
	if !ok {
		return types.Task{}, fmt.Errorf("no binding for master agent: %s", task.MasterAgent)
	}
	results, currentStep, needsConfirm, pendingResults, status, preview, logs, errMsg, runErr := o.runBinding(ctx, task.Request, binding, 0, nil)
	task.Results = results
	task.CurrentStep = currentStep
	task.NeedsConfirm = needsConfirm
	task.PendingResults = pendingResults
	task.Status = status
	task.Preview = preview
	task.ErrorMessage = errMsg
	task.UpdatedAt = time.Now()
	task.Logs = append(task.Logs, types.TaskLog{Time: task.UpdatedAt, Step: 0, Agent: task.MasterAgent, Action: "task_retried", Message: "task retried"})
	task.Logs = append(task.Logs, logs...)
	_ = o.store.Save(task)
	if runErr != nil {
		return *task, runErr
	}
	return *task, nil
}

func (o *Orchestrator) runBinding(ctx context.Context, req types.Request, binding types.Binding, startStep int, existing []types.Response) ([]types.Response, int, bool, []types.Response, string, *types.Preview, []types.TaskLog, string, error) {
	results := append([]types.Response{}, existing...)
	logs := make([]types.TaskLog, 0)

	if len(binding.Workflow) > 0 {
		workflow := append([]types.WorkflowStep{}, binding.Workflow...)
		sort.Slice(workflow, func(i, j int) bool { return workflow[i].Step < workflow[j].Step })
		for _, step := range workflow {
			if step.Step <= startStep {
				continue
			}
			logs = append(logs, types.TaskLog{
				Time:    time.Now(),
				Step:    step.Step,
				Agent:   step.Ability,
				Action:  "step_started",
				Message: "workflow step started",
			})
			agent, err := o.registry.Get(step.Ability)
			if err != nil {
				return results, step.Step, false, nil, types.TaskStatusFailed, nil, logs, err.Error(), err
			}
			resp, err := agent.Run(ctx, req)
			if err != nil {
				logs = append(logs, types.TaskLog{
					Time:    time.Now(),
					Step:    step.Step,
					Agent:   step.Ability,
					Action:  "step_failed",
					Message: err.Error(),
				})
				return results, step.Step, false, nil, types.TaskStatusFailed, nil, logs, err.Error(), fmt.Errorf("run %s failed: %w", step.Ability, err)
			}
			results = append(results, resp)
			logs = append(logs, types.TaskLog{
				Time:    time.Now(),
				Step:    step.Step,
				Agent:   step.Ability,
				Action:  "step_completed",
				Message: "workflow step completed",
				Data:    resp.Data,
			})
			if step.RequireHumanConfirm {
				if resp.Data == nil {
					resp.Data = map[string]any{}
				}
				resp.Data["require_human_confirm"] = true
				results[len(results)-1] = resp
				preview := buildPreview(req, results)
				logs = append(logs, types.TaskLog{
					Time:    time.Now(),
					Step:    step.Step,
					Agent:   step.Ability,
					Action:  "awaiting_confirmation",
					Message: "waiting for human confirmation",
				})
				return results, step.Step, true, []types.Response{resp}, types.TaskStatusPendingConfirm, preview, logs, "", nil
			}
		}
		return results, len(workflow), false, nil, types.TaskStatusSuccess, buildPreview(req, results), logs, "", nil
	}

	for idx, code := range binding.Abilities {
		step := idx + 1
		agent, err := o.registry.Get(code)
		if err != nil {
			return results, step, false, nil, types.TaskStatusFailed, nil, logs, err.Error(), err
		}
		resp, err := agent.Run(ctx, req)
		if err != nil {
			return results, step, false, nil, types.TaskStatusFailed, nil, logs, err.Error(), fmt.Errorf("run %s failed: %w", code, err)
		}
		results = append(results, resp)
	}
	return results, len(binding.Abilities), false, nil, types.TaskStatusSuccess, buildPreview(req, results), logs, "", nil
}

func buildPreview(req types.Request, results []types.Response) *types.Preview {
	fields := map[string]any{
		"scene": req.Scene,
		"input": req.Input,
	}
	for _, r := range results {
		switch r.Agent {
		case "content_gen":
			if content, ok := r.Data["content"]; ok {
				fields["content"] = content
			}
		case "page_gen":
			if page, ok := r.Data["page"]; ok {
				fields["page"] = page
			}
		case "publish_exec":
			fields["publish_status"] = r.Data["publish_status"]
		case "onshelf_exec":
			fields["shelf_status"] = r.Data["shelf_status"]
		}
	}
	return &types.Preview{
		Title:   fmt.Sprintf("%s 场景预览", req.Scene),
		Summary: "系统已生成预览数据，可用于展示或人工确认",
		Fields:  fields,
	}
}
