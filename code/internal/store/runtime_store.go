package store

import "juyu-ai-platform/internal/types"

type RuntimeStore interface {
	CreateSession(input types.CreateSessionInput) (*types.RuntimeSession, error)
	ListSessions(status, inputType string, limit, offset int) ([]types.RuntimeSession, int, error)
	GetSession(sessionID string) (*types.RuntimeSession, error)
	UpdatePrd(sessionID string, patch map[string]any) (*types.RuntimeSession, error)
	EditPrd(sessionID string, patch map[string]any, comment string) (*types.RuntimeSession, error)
	ConfirmPrd(sessionID string, comment string) (*types.RuntimeSession, error)
	GetTodo(sessionID string) (*types.RuntimeTodoArtifact, error)
	UpdateTodo(sessionID string, items []types.RuntimeTodoItem) (*types.RuntimeSession, error)
	EditTodo(sessionID string, items []types.RuntimeTodoItem, comment string) (*types.RuntimeSession, error)
	ConfirmTodo(sessionID string, comment string) (*types.RuntimeSession, error)
	CancelSession(sessionID string, reason string) (*types.RuntimeSession, error)
	ListExecutions(sessionID, status, todoID string, limit, offset int) ([]types.RuntimeExecution, int, error)
	GetExecution(sessionID, executionID string) (*types.RuntimeExecution, error)
	RetryExecutions(sessionID string, req types.RetryExecutionsRequest) ([]types.RetryResultItem, *types.RuntimeSession, error)
	ListLogs(sessionID, executionID, todoID, level string, limit, offset int) ([]types.RuntimeLogEntry, int, error)
	UpdatePreview(sessionID string, preview map[string]any) (*types.RuntimeSession, error)
	GetPreview(sessionID string) (map[string]any, *types.RuntimeSession, error)
}
