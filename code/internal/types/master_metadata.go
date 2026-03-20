package types

type MasterAgentMetadata struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	SceneType string `json:"scene_type"`
	Enabled   bool   `json:"enabled"`
}

type BindingView struct {
	MasterAgent string         `json:"master_agent"`
	SceneType   string         `json:"scene_type,omitempty"`
	Abilities   []string       `json:"abilities,omitempty"`
	Workflow    []WorkflowStep `json:"workflow,omitempty"`
}
