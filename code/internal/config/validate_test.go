package config

import (
	"testing"

	"juyu-ai-platform/internal/types"
)

func validConfig() types.Config {
	return types.Config{
		MasterAgents:  []types.MasterAgent{{Code: "product_ops", SceneType: "product", Enabled: true}},
		AbilityAgents: []types.AbilityConfig{{Code: "content_gen", Enabled: true}, {Code: "review_check", Enabled: true}},
		Bindings:      []types.Binding{{MasterAgent: "product_ops", Workflow: []types.WorkflowStep{{Step: 1, Ability: "content_gen"}, {Step: 2, Ability: "review_check"}}}},
	}
}

func TestValidateConfig(t *testing.T) {
	if err := Validate(validConfig()); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}
}

func TestValidateConfigRejectsUnknownAbility(t *testing.T) {
	cfg := validConfig()
	cfg.Bindings[0].Workflow[0].Ability = "missing"
	if err := Validate(cfg); err == nil {
		t.Fatal("expected validation error for unknown ability")
	}
}

func TestValidateRuntime(t *testing.T) {
	cases := []struct {
		name      string
		store     string
		sqlite    string
		pg        string
		apiKey    string
		dryRun    bool
		wantError bool
	}{
		{name: "memory ok", store: "memory", dryRun: true},
		{name: "sqlite ok", store: "sqlite", sqlite: "juyu.db", dryRun: true},
		{name: "postgres ok", store: "postgres", pg: "host=postgres", dryRun: true},
		{name: "sqlite missing path", store: "sqlite", dryRun: true, wantError: true},
		{name: "postgres missing dsn", store: "postgres", dryRun: true, wantError: true},
		{name: "live mode needs api key", store: "memory", dryRun: false, wantError: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateRuntime(tc.store, tc.sqlite, tc.pg, tc.apiKey, tc.dryRun)
			if tc.wantError && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.wantError && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
