package types

import "context"

type Request struct {
	Scene   string         `json:"scene"`
	Input   string         `json:"input"`
	Payload map[string]any `json:"payload"`
}

type Response struct {
	Agent   string         `json:"agent"`
	Success bool           `json:"success"`
	Data    map[string]any `json:"data"`
}

type ExecuteResponse struct {
	MasterAgent string     `json:"master_agent"`
	Scene       string     `json:"scene"`
	Results     []Response `json:"results"`
}

type AbilityAgent interface {
	Code() string
	Run(ctx context.Context, req Request) (Response, error)
}

type MasterAgent struct {
	Code      string `yaml:"code"`
	Name      string `yaml:"name"`
	SceneType string `yaml:"scene_type"`
	Enabled   bool   `yaml:"enabled"`
}

type AbilityConfig struct {
	Code    string `yaml:"code"`
	Name    string `yaml:"name"`
	Type    string `yaml:"type"`
	Enabled bool   `yaml:"enabled"`
}

type WorkflowStep struct {
	Step                int    `yaml:"step"`
	Ability             string `yaml:"ability"`
	RequireHumanConfirm bool   `yaml:"require_human_confirm"`
}

type Binding struct {
	MasterAgent string         `yaml:"master_agent"`
	Abilities   []string       `yaml:"abilities"`
	Workflow    []WorkflowStep `yaml:"workflow"`
}

type Config struct {
	MasterAgents  []MasterAgent  `yaml:"master_agents"`
	AbilityAgents []AbilityConfig `yaml:"ability_agents"`
	Bindings      []Binding      `yaml:"bindings"`
}
