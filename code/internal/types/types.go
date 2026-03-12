package types

import "context"

type Request struct {
	Scene   string
	Input   string
	Payload map[string]any
}

type Response struct {
	Agent   string         `json:"agent"`
	Success bool           `json:"success"`
	Data    map[string]any `json:"data"`
}

type AbilityAgent interface {
	Code() string
	Run(ctx context.Context, req Request) (Response, error)
}

type MasterAgent struct {
	Code      string
	Name      string
	SceneType string
	Enabled   bool
}

type AbilityConfig struct {
	Code    string
	Name    string
	Type    string
	Enabled bool
}

type Binding struct {
	MasterAgent string
	Abilities   []string
}

type Config struct {
	MasterAgents  []MasterAgent  `yaml:"master_agents"`
	AbilityAgents []AbilityConfig `yaml:"ability_agents"`
	Bindings      []Binding      `yaml:"bindings"`
}
