package server

import (
	"testing"

	"juyu-ai-platform/internal/types"
)

func TestRuntimeServicePickBindingByInferredScene(t *testing.T) {
	svc := newRuntimeService(nil, []types.BindingView{
		{MasterAgent: "product_ops", SceneType: "product", Abilities: []string{"content_gen"}},
		{MasterAgent: "content_publish", SceneType: "content", Abilities: []string{"image_gen"}},
		{MasterAgent: "competition_builder", SceneType: "competition", Abilities: []string{"proposal_gen"}},
	})

	cases := []struct {
		name       string
		sess       *types.RuntimeSession
		wantMaster string
	}{
		{
			name:       "product url learning goes to product binding",
			sess:       &types.RuntimeSession{Input: types.CreateSessionInput{Type: "product_url_learning", URL: "https://example.com/p"}},
			wantMaster: "product_ops",
		},
		{
			name:       "content publish goes to content binding",
			sess:       &types.RuntimeSession{Input: types.CreateSessionInput{Type: "content_publish_request", Message: "帮我做内容分发"}},
			wantMaster: "content_publish",
		},
		{
			name:       "competition request goes to competition binding",
			sess:       &types.RuntimeSession{Input: types.CreateSessionInput{Type: "competition_plan", Message: "生成比赛方案PPT"}},
			wantMaster: "competition_builder",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			binding := svc.pickBinding(tc.sess)
			if binding == nil {
				t.Fatal("expected binding")
			}
			if binding.MasterAgent != tc.wantMaster {
				t.Fatalf("expected %s, got %s", tc.wantMaster, binding.MasterAgent)
			}
		})
	}
}

func TestRuntimeRequestFromSessionUsesInferredScene(t *testing.T) {
	req := runtimeRequestFromSession(&types.RuntimeSession{Input: types.CreateSessionInput{Type: "competition_plan", Message: "做一个比赛方案"}})
	if req.Scene != "competition" {
		t.Fatalf("expected competition scene, got %s", req.Scene)
	}
	if got := req.Payload["type"]; got != "competition_plan" {
		t.Fatalf("expected payload type competition_plan, got %v", got)
	}
}

func TestEnrichBindingViewsInjectsSceneType(t *testing.T) {
	bindings := enrichBindingViews(
		[]types.MasterAgentMetadata{{Code: "content_publish", SceneType: "content"}},
		[]types.BindingView{{MasterAgent: "content_publish"}},
	)
	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}
	if bindings[0].SceneType != "content" {
		t.Fatalf("expected content scene type, got %q", bindings[0].SceneType)
	}
}
