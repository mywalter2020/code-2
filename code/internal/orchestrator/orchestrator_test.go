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

func TestExecuteProductFlow(t *testing.T) {
	adapterRegistry := adapters.NewRegistry()
	adapterRegistry.Register(adapters.NewAlibabaAdapter())

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
