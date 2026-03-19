package types

import "time"

type SessionStatus string

type TodoStatus string

type ExecutionStatus string

const (
	SessionStatusInputReceived     SessionStatus = "input_received"
	SessionStatusAnalysisDone      SessionStatus = "analysis_done"
	SessionStatusWaitingPrdConfirm SessionStatus = "waiting_prd_confirm"
	SessionStatusTodoDone          SessionStatus = "todo_done"
	SessionStatusWaitingTodo       SessionStatus = "waiting_todo_confirm"
	SessionStatusExecuting         SessionStatus = "executing"
	SessionStatusDone              SessionStatus = "done"
	SessionStatusFailed            SessionStatus = "failed"
	SessionStatusCanceled          SessionStatus = "canceled"
)

const (
	TodoStatusPending  TodoStatus = "pending"
	TodoStatusReady    TodoStatus = "ready"
	TodoStatusRunning  TodoStatus = "running"
	TodoStatusBlocked  TodoStatus = "blocked"
	TodoStatusDone     TodoStatus = "done"
	TodoStatusFailed   TodoStatus = "failed"
	TodoStatusCanceled TodoStatus = "canceled"
)

const (
	ExecutionStatusQueued   ExecutionStatus = "queued"
	ExecutionStatusRunning  ExecutionStatus = "running"
	ExecutionStatusDone     ExecutionStatus = "done"
	ExecutionStatusFailed   ExecutionStatus = "failed"
	ExecutionStatusCanceled ExecutionStatus = "canceled"
)

type CreateSessionInput struct {
	Type     string `json:"type"`
	URL      string `json:"url"`
	Message  string `json:"message"`
	Platform string `json:"platform,omitempty"`
	Language string `json:"language,omitempty"`
	Style    string `json:"style,omitempty"`
}

type CreateSessionRequest struct {
	Input CreateSessionInput `json:"input"`
}

type RuntimePRD struct {
	Version       int      `json:"version,omitempty"`
	Title         string   `json:"title,omitempty"`
	Background    string   `json:"background,omitempty"`
	Goals         []string `json:"goals,omitempty"`
	Requirements  []string `json:"requirements,omitempty"`
	OpenQuestions []string `json:"open_questions,omitempty"`
	Markdown      string   `json:"markdown,omitempty"`
}

type RuntimeTodoItem struct {
	ID                 string         `json:"id"`
	Title              string         `json:"title"`
	Type               string         `json:"type"`
	Description        string         `json:"description,omitempty"`
	Status             TodoStatus     `json:"status"`
	DependsOn          []string       `json:"depends_on,omitempty"`
	ParallelGroup      string         `json:"parallel_group,omitempty"`
	AcceptanceCriteria []string       `json:"acceptance_criteria,omitempty"`
	Result             map[string]any `json:"result,omitempty"`
}

type RuntimeTodoArtifact struct {
	Version int               `json:"version,omitempty"`
	Items   []RuntimeTodoItem `json:"items"`
}

type RuntimeSession struct {
	SessionID    string               `json:"session_id"`
	Status       SessionStatus        `json:"status"`
	CurrentStage string               `json:"current_stage"`
	Input        CreateSessionInput   `json:"input"`
	PRD          *RuntimePRD          `json:"prd,omitempty"`
	Todo         *RuntimeTodoArtifact `json:"todo,omitempty"`
	Preview      map[string]any       `json:"preview,omitempty"`
	Message      string               `json:"message,omitempty"`
	NextActions  []string             `json:"next_actions,omitempty"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
}

type RuntimeExecution struct {
	ExecutionID        string          `json:"execution_id"`
	ParentExecutionID  string          `json:"parent_execution_id,omitempty"`
	RetryOfExecutionID string          `json:"retry_of_execution_id,omitempty"`
	SessionID          string          `json:"session_id,omitempty"`
	TodoID             string          `json:"todo_id"`
	TodoVersion        int             `json:"todo_version,omitempty"`
	Title              string          `json:"title,omitempty"`
	Executor           string          `json:"executor,omitempty"`
	SkillCode          string          `json:"skill_code,omitempty"`
	Status             ExecutionStatus `json:"status"`
	Attempt            int             `json:"attempt"`
	QueueReason        string          `json:"queue_reason,omitempty"`
	Input              map[string]any  `json:"input,omitempty"`
	Output             map[string]any  `json:"output,omitempty"`
	ReasoningSummary   string          `json:"reasoning_summary,omitempty"`
	ErrorCode          string          `json:"error_code,omitempty"`
	ErrorMessage       string          `json:"error_message,omitempty"`
	Metrics            map[string]any  `json:"metrics,omitempty"`
	CreatedAt          time.Time       `json:"created_at,omitempty"`
	StartedAt          *time.Time      `json:"started_at,omitempty"`
	FinishedAt         *time.Time      `json:"finished_at,omitempty"`
}

type RuntimeLogEntry struct {
	LogID       int64          `json:"log_id"`
	Time        time.Time      `json:"time"`
	Level       string         `json:"level"`
	SourceType  string         `json:"source_type"`
	SourceCode  string         `json:"source_code"`
	EventType   string         `json:"event_type,omitempty"`
	Message     string         `json:"message"`
	SessionID   string         `json:"session_id,omitempty"`
	ExecutionID string         `json:"execution_id,omitempty"`
	TodoID      string         `json:"todo_id,omitempty"`
	Data        map[string]any `json:"data,omitempty"`
}

type ConfirmPrdRequest struct {
	Comment string `json:"comment,omitempty"`
}

type EditPrdRequest struct {
	Patch   map[string]any `json:"patch,omitempty"`
	Comment string         `json:"comment,omitempty"`
}

type ConfirmTodoRequest struct {
	Comment string `json:"comment,omitempty"`
}

type EditTodoRequest struct {
	Items   []RuntimeTodoItem `json:"items"`
	Comment string            `json:"comment,omitempty"`
}

type RetryTarget struct {
	TodoID                  string `json:"todo_id"`
	LatestFailedExecutionID string `json:"latest_failed_execution_id,omitempty"`
	Force                   bool   `json:"force,omitempty"`
}

type RetryExecutionsRequest struct {
	Items  []RetryTarget `json:"items"`
	Reason string        `json:"reason,omitempty"`
}

type RetryResultItem struct {
	TodoID              string `json:"todo_id"`
	PreviousExecutionID string `json:"previous_execution_id,omitempty"`
	NewExecutionID      string `json:"new_execution_id,omitempty"`
	PreviousAttempt     int    `json:"previous_attempt,omitempty"`
	NewAttempt          int    `json:"new_attempt,omitempty"`
	Accepted            bool   `json:"accepted"`
	Reason              string `json:"reason,omitempty"`
}

type CancelSessionRequest struct {
	Reason string `json:"reason,omitempty"`
}
