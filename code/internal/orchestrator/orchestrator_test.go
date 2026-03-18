package orchestrator

import (
	"context"
	"testing"

	"juyu-ai-platform/internal/adapters"
	"juyu-ai-platform/internal/agents"
	"juyu-ai-platform/internal/registry"
	"juyu-ai-platform/internal/router"
	"juyu-ai-platform/internal/store"
	"juyu-ai-platform/internal/types"
)

type stubAgent struct {
	code string
	run  func(ctx context.Context, req types.Request) (types.Response, error)
}

func (a stubAgent) Code() string { return a.code }
func (a stubAgent) Run(ctx context.Context, req types.Request) (types.Response, error) {
	return a.run(ctx, req)
}

func TestExecuteProductFlow(t *testing.T) {
	adapterRegistry := adapters.NewRegistry()
	adapterRegistry.Register(adapters.NewAlibabaAdapter(adapters.NewCredentialStore(), true))

	rg := registry.New()
	rg.Register(agents.NewContentAgent())
	rg.Register(agents.NewPageAgent())
	rg.Register(agents.NewReviewAgent())
	rg.Register(agents.NewPublishAgent(adapterRegistry))
	rg.Register(agents.NewOnShelfAgent(adapterRegistry))

	rt := router.New([]types.MasterAgent{{Code: "product_ops", SceneType: "product", Enabled: true}})
	st := store.NewMemoryStore()
	orc := New(rt, rg, st, []types.Binding{{MasterAgent: "product_ops", Workflow: []types.WorkflowStep{{Step: 1, Ability: "content_gen"}, {Step: 2, Ability: "page_gen"}, {Step: 3, Ability: "review_check", RequireHumanConfirm: true}, {Step: 4, Ability: "publish_exec"}}}})

	resp, err := orc.Execute(context.Background(), types.Request{Scene: "product", Input: "demo"})
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if resp.Status != types.TaskStatusPendingConfirm {
		t.Fatalf("expected pending_confirm, got %s", resp.Status)
	}
	if resp.TaskID == "" {
		t.Fatal("expected task id")
	}
}

func TestWorkflowSupportsConditionalBranch(t *testing.T) {
	rg := registry.New()
	rg.Register(stubAgent{code: "content_gen", run: func(ctx context.Context, req types.Request) (types.Response, error) {
		return types.Response{Agent: "content_gen", Success: true, Data: map[string]any{"content": "ok"}}, nil
	}})
	rg.Register(stubAgent{code: "publish_exec", run: func(ctx context.Context, req types.Request) (types.Response, error) {
		return types.Response{Agent: "publish_exec", Success: true, Data: map[string]any{"publish_status": "done"}}, nil
	}})
	rg.Register(stubAgent{code: "onshelf_exec", run: func(ctx context.Context, req types.Request) (types.Response, error) {
		return types.Response{Agent: "onshelf_exec", Success: true, Data: map[string]any{"shelf_status": "done"}}, nil
	}})

	rt := router.New([]types.MasterAgent{{Code: "product_ops", SceneType: "product", Enabled: true}})
	st := store.NewMemoryStore()
	orc := New(rt, rg, st, []types.Binding{{MasterAgent: "product_ops", Workflow: []types.WorkflowStep{
		{ID: "draft", Step: 1, Ability: "content_gen"},
		{ID: "publish", Step: 2, Ability: "publish_exec", DependsOn: []string{"draft"}, When: []types.WorkflowCondition{{Source: "request", Path: "mode", Equals: "publish"}}},
		{ID: "shelf", Step: 3, Ability: "onshelf_exec", DependsOn: []string{"draft"}, When: []types.WorkflowCondition{{Source: "request", Path: "mode", Equals: "onshelf"}}},
	}}})

	resp, err := orc.Execute(context.Background(), types.Request{Scene: "product", Input: "demo", Payload: map[string]any{"mode": "publish"}})
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if resp.Status != types.TaskStatusSuccess {
		t.Fatalf("expected success, got %s", resp.Status)
	}
	if len(resp.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(resp.Results))
	}
	if resp.Results[1].Agent != "publish_exec" {
		t.Fatalf("expected publish_exec, got %s", resp.Results[1].Agent)
	}
}

func TestWorkflowSupportsInvokeBinding(t *testing.T) {
	rg := registry.New()
	rg.Register(stubAgent{code: "content_gen", run: func(ctx context.Context, req types.Request) (types.Response, error) {
		return types.Response{Agent: "content_gen", Success: true, Data: map[string]any{"content": "generated"}}, nil
	}})
	rg.Register(stubAgent{code: "page_gen", run: func(ctx context.Context, req types.Request) (types.Response, error) {
		biz, ok := req.Payload["_business_context"].(types.BusinessContext)
		if !ok {
			t.Fatalf("expected _business_context in payload")
		}
		return types.Response{Agent: "page_gen", Success: true, Data: map[string]any{"page": map[string]any{"title": biz.Product.Title}}}, nil
	}})

	rt := router.New([]types.MasterAgent{{Code: "main_ops", SceneType: "product", Enabled: true}, {Code: "content_flow", SceneType: "content", Enabled: true}})
	st := store.NewMemoryStore()
	orc := New(rt, rg, st, []types.Binding{
		{MasterAgent: "content_flow", Workflow: []types.WorkflowStep{{ID: "draft", Step: 1, Ability: "content_gen"}}},
		{MasterAgent: "main_ops", Workflow: []types.WorkflowStep{{ID: "sub", Step: 1, InvokeBinding: "content_flow"}, {ID: "page", Step: 2, Ability: "page_gen", DependsOn: []string{"sub"}}}},
	})

	resp, err := orc.Execute(context.Background(), types.Request{Scene: "product", Input: "demo title"})
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if len(resp.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(resp.Results))
	}
	if resp.Results[0].Agent != "binding:content_flow" {
		t.Fatalf("expected binding response, got %s", resp.Results[0].Agent)
	}
	if resp.Results[1].Agent != "page_gen" {
		t.Fatalf("expected page_gen, got %s", resp.Results[1].Agent)
	}
}
