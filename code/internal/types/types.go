package types

import (
	"context"
	"time"
)

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
	TaskID      string     `json:"task_id"`
	MasterAgent string     `json:"master_agent"`
	Scene       string     `json:"scene"`
	Status      string     `json:"status"`
	Preview     *Preview   `json:"preview,omitempty"`
	Results     []Response `json:"results"`
}

type ConfirmRequest struct {
	Approved bool   `json:"approved"`
	Comment  string `json:"comment,omitempty"`
}

type Preview struct {
	Title   string         `json:"title"`
	Summary string         `json:"summary"`
	Fields  map[string]any `json:"fields"`
}

type TaskLog struct {
	Time    time.Time      `json:"time"`
	Step    int            `json:"step"`
	Agent   string         `json:"agent"`
	Action  string         `json:"action"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data,omitempty"`
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
	MasterAgents  []MasterAgent   `yaml:"master_agents"`
	AbilityAgents []AbilityConfig `yaml:"ability_agents"`
	Bindings      []Binding       `yaml:"bindings"`
}

type Task struct {
	ID               string     `json:"id"`
	Request          Request    `json:"request"`
	MasterAgent      string     `json:"master_agent"`
	Status           string     `json:"status"`
	CurrentStep      int        `json:"current_step"`
	NeedsConfirm     bool       `json:"needs_confirm"`
	Preview          *Preview   `json:"preview,omitempty"`
	Results          []Response `json:"results"`
	PendingResults   []Response `json:"pending_results,omitempty"`
	Logs             []TaskLog  `json:"logs,omitempty"`
	ErrorMessage     string     `json:"error_message,omitempty"`
	ConfirmComment   string     `json:"confirm_comment,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	LastConfirmedAt  *time.Time `json:"last_confirmed_at,omitempty"`
}

const (
	TaskStatusPendingConfirm = "pending_confirm"
	TaskStatusRunning        = "running"
	TaskStatusSuccess        = "success"
	TaskStatusRejected       = "rejected"
	TaskStatusFailed         = "failed"
)
