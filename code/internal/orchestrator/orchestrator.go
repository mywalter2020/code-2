package orchestrator

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
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
		return o.runWorkflow(ctx, req, binding, startStep, results, logs)
	}

	for idx, code := range binding.Abilities {
		step := idx + 1
		agent, err := o.registry.Get(code)
		if err != nil {
			return results, step, false, nil, types.TaskStatusFailed, nil, logs, err.Error(), err
		}
		resp, err := agent.Run(ctx, enrichRequest(req, results))
		if err != nil {
			return results, step, false, nil, types.TaskStatusFailed, nil, logs, err.Error(), fmt.Errorf("run %s failed: %w", code, err)
		}
		results = append(results, resp)
	}
	return results, len(binding.Abilities), false, nil, types.TaskStatusSuccess, buildPreview(req, results), logs, "", nil
}

func (o *Orchestrator) runWorkflow(ctx context.Context, req types.Request, binding types.Binding, startStep int, existing []types.Response, logs []types.TaskLog) ([]types.Response, int, bool, []types.Response, string, *types.Preview, []types.TaskLog, string, error) {
	workflow := append([]types.WorkflowStep{}, binding.Workflow...)
	sort.Slice(workflow, func(i, j int) bool { return workflow[i].Step < workflow[j].Step })

	results := append([]types.Response{}, existing...)
	completed := map[string]struct{}{}
	for _, step := range workflow {
		if step.Step <= startStep {
			completed[stepKey(step)] = struct{}{}
		}
	}

	for len(completed) < len(workflow) {
		ready := make([]types.WorkflowStep, 0)
		for _, step := range workflow {
			key := stepKey(step)
			if _, ok := completed[key]; ok {
				continue
			}
			if !depsSatisfied(step, workflow, completed, startStep) {
				continue
			}
			ok, err := evaluateConditions(step, req, results)
			if err != nil {
				return results, step.Step, false, nil, types.TaskStatusFailed, nil, logs, err.Error(), err
			}
			if !ok {
				completed[key] = struct{}{}
				logs = append(logs, types.TaskLog{Time: time.Now(), Step: step.Step, Agent: displayStepAgent(step), Action: "step_skipped", Message: "workflow step skipped by condition"})
				continue
			}
			ready = append(ready, step)
		}

		if len(ready) == 0 {
			return results, startStep, false, nil, types.TaskStatusFailed, nil, logs, "workflow deadlock or unmet dependencies", fmt.Errorf("workflow deadlock or unmet dependencies")
		}
		sort.Slice(ready, func(i, j int) bool { return ready[i].Step < ready[j].Step })

		batchResults, batchLogs, currentStep, needsConfirm, pendingResults, errMsg, err := o.runReadyBatch(ctx, req, ready, results)
		logs = append(logs, batchLogs...)
		if err != nil {
			return results, currentStep, false, nil, types.TaskStatusFailed, nil, logs, errMsg, err
		}
		results = append(results, batchResults...)
		for _, step := range ready {
			completed[stepKey(step)] = struct{}{}
		}
		if needsConfirm {
			preview := buildPreview(req, results)
			return results, currentStep, true, pendingResults, types.TaskStatusPendingConfirm, preview, logs, "", nil
		}
		startStep = currentStep
	}

	return results, maxWorkflowStep(workflow), false, nil, types.TaskStatusSuccess, buildPreview(req, results), logs, "", nil
}

type readyStepResult struct {
	step         types.WorkflowStep
	resp         types.Response
	logs         []types.TaskLog
	err          error
	errMsg       string
	needsConfirm bool
}

func (o *Orchestrator) runReadyBatch(ctx context.Context, req types.Request, steps []types.WorkflowStep, existing []types.Response) ([]types.Response, []types.TaskLog, int, bool, []types.Response, string, error) {
	logs := make([]types.TaskLog, 0)
	results := make([]types.Response, 0, len(steps))
	items := make([]readyStepResult, len(steps))

	var wg sync.WaitGroup
	for i, step := range steps {
		logs = append(logs, types.TaskLog{Time: time.Now(), Step: step.Step, Agent: displayStepAgent(step), Action: "step_started", Message: "workflow step started"})
		wg.Add(1)
		go func(i int, step types.WorkflowStep) {
			defer wg.Done()
			resp, needsConfirm, stepLogs, errMsg, err := o.runStep(ctx, req, step, existing)
			items[i] = readyStepResult{step: step, resp: resp, logs: stepLogs, err: err, errMsg: errMsg, needsConfirm: needsConfirm}
		}(i, step)
	}
	wg.Wait()

	for _, item := range items {
		logs = append(logs, item.logs...)
		if item.err != nil {
			return results, logs, item.step.Step, false, nil, item.errMsg, item.err
		}
		results = append(results, item.resp)
	}

	pending := make([]types.Response, 0)
	confirmStep := 0
	for i, item := range items {
		if item.needsConfirm {
			if results[i].Data == nil {
				results[i].Data = map[string]any{}
			}
			results[i].Data["require_human_confirm"] = true
			pending = append(pending, results[i])
			if item.step.Step > confirmStep {
				confirmStep = item.step.Step
			}
		}
	}
	if len(pending) > 0 {
		logs = append(logs, types.TaskLog{Time: time.Now(), Step: confirmStep, Agent: "orchestrator", Action: "awaiting_confirmation", Message: "waiting for human confirmation"})
		return results, logs, confirmStep, true, pending, "", nil
	}

	return results, logs, maxReadyStep(steps), false, nil, "", nil
}

func (o *Orchestrator) runStep(ctx context.Context, req types.Request, step types.WorkflowStep, existing []types.Response) (types.Response, bool, []types.TaskLog, string, error) {
	if step.InvokeBinding != "" {
		binding, ok := o.bindings[step.InvokeBinding]
		if !ok {
			return types.Response{}, false, nil, fmt.Sprintf("unknown binding: %s", step.InvokeBinding), fmt.Errorf("unknown binding: %s", step.InvokeBinding)
		}
		childResults, _, childConfirm, _, _, childPreview, childLogs, errMsg, err := o.runBinding(ctx, req, binding, 0, existing)
		resp := types.Response{
			Agent:   displayStepAgent(step),
			Success: err == nil,
			Data: map[string]any{
				"binding": binding.MasterAgent,
				"results": childResults,
			},
		}
		if childPreview != nil {
			resp.Data["preview"] = childPreview
		}
		logs := append([]types.TaskLog{}, childLogs...)
		if err != nil {
			logs = append(logs, types.TaskLog{Time: time.Now(), Step: step.Step, Agent: displayStepAgent(step), Action: "step_failed", Message: err.Error()})
			return resp, false, logs, errMsg, fmt.Errorf("invoke binding %s failed: %w", step.InvokeBinding, err)
		}
		logs = append(logs, types.TaskLog{Time: time.Now(), Step: step.Step, Agent: displayStepAgent(step), Action: "step_completed", Message: "workflow sub-binding completed", Data: resp.Data})
		return resp, step.RequireHumanConfirm || childConfirm, logs, "", nil
	}

	agent, err := o.registry.Get(step.Ability)
	if err != nil {
		return types.Response{}, false, []types.TaskLog{{Time: time.Now(), Step: step.Step, Agent: displayStepAgent(step), Action: "step_failed", Message: err.Error()}}, err.Error(), err
	}
	resp, err := agent.Run(ctx, enrichRequest(req, existing))
	if err != nil {
		logs := []types.TaskLog{{Time: time.Now(), Step: step.Step, Agent: displayStepAgent(step), Action: "step_failed", Message: err.Error()}}
		return types.Response{}, false, logs, err.Error(), fmt.Errorf("run %s failed: %w", step.Ability, err)
	}
	logs := []types.TaskLog{{Time: time.Now(), Step: step.Step, Agent: displayStepAgent(step), Action: "step_completed", Message: "workflow step completed", Data: resp.Data}}
	return resp, step.RequireHumanConfirm, logs, "", nil
}

func enrichRequest(req types.Request, results []types.Response) types.Request {
	out := req
	payload := map[string]any{}
	for k, v := range req.Payload {
		payload[k] = v
	}
	if len(results) > 0 {
		payload["_results"] = results
		payload["_business_context"] = types.BuildBusinessContext(req, results)
	}
	out.Payload = payload
	return out
}

func depsSatisfied(step types.WorkflowStep, workflow []types.WorkflowStep, completed map[string]struct{}, startStep int) bool {
	if len(step.DependsOn) > 0 {
		for _, dep := range step.DependsOn {
			if _, ok := completed[dep]; !ok {
				return false
			}
		}
		return true
	}
	for _, candidate := range workflow {
		if candidate.Step >= step.Step {
			break
		}
		if candidate.Step <= startStep {
			continue
		}
		if _, ok := completed[stepKey(candidate)]; !ok {
			return false
		}
	}
	return true
}

func evaluateConditions(step types.WorkflowStep, req types.Request, results []types.Response) (bool, error) {
	for _, cond := range step.When {
		value, ok := resolveConditionValue(cond.Source, cond.Path, req, results)
		if cond.Exists != nil {
			if ok != *cond.Exists {
				return false, nil
			}
			continue
		}
		if !ok {
			return false, nil
		}
		if cond.Equals != nil && fmt.Sprint(value) != fmt.Sprint(cond.Equals) {
			return false, nil
		}
		if cond.NotEquals != nil && fmt.Sprint(value) == fmt.Sprint(cond.NotEquals) {
			return false, nil
		}
	}
	return true, nil
}

func resolveConditionValue(source, path string, req types.Request, results []types.Response) (any, bool) {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "request", "payload", "req":
		return walkValue(req.Payload, path)
	case "result", "results":
		parts := splitPath(path)
		if len(parts) == 0 {
			return nil, false
		}
		agent := parts[0]
		for i := len(results) - 1; i >= 0; i-- {
			if results[i].Agent == agent {
				if len(parts) == 1 {
					return results[i].Data, true
				}
				return walkValue(results[i].Data, strings.Join(parts[1:], "."))
			}
		}
	}
	return nil, false
}

func walkValue(root map[string]any, path string) (any, bool) {
	if len(root) == 0 {
		return nil, false
	}
	parts := splitPath(path)
	if len(parts) == 0 {
		return root, true
	}
	var current any = root
	for _, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = m[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func splitPath(path string) []string {
	items := strings.Split(strings.TrimSpace(path), ".")
	out := make([]string, 0, len(items))
	for _, item := range items {
		if s := strings.TrimSpace(item); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func stepKey(step types.WorkflowStep) string {
	if s := strings.TrimSpace(step.ID); s != "" {
		return s
	}
	return fmt.Sprintf("step:%d", step.Step)
}

func displayStepAgent(step types.WorkflowStep) string {
	if step.InvokeBinding != "" {
		return "binding:" + step.InvokeBinding
	}
	return step.Ability
}

func maxWorkflowStep(workflow []types.WorkflowStep) int {
	max := 0
	for _, step := range workflow {
		if step.Step > max {
			max = step.Step
		}
	}
	return max
}

func maxReadyStep(steps []types.WorkflowStep) int {
	max := 0
	for _, step := range steps {
		if step.Step > max {
			max = step.Step
		}
	}
	return max
}

func buildPreview(req types.Request, results []types.Response) *types.Preview {
	ctx := types.BuildBusinessContext(req, results)
	fields := map[string]any{
		"scene":        req.Scene,
		"input":        req.Input,
		"product":      ctx.Product,
		"publish":      ctx.Publish,
		"confirmation": ctx.Confirmation,
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
