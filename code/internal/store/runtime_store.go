package store

import "juyu-ai-platform/internal/types"

type RuntimeStore interface {
	CreateSession(input types.CreateSessionInput) (*types.RuntimeSession, error)
	GetSession(sessionID string) (*types.RuntimeSession, error)
	EditPrd(sessionID string, patch map[string]any, comment string) (*types.RuntimeSession, error)
	ConfirmPrd(sessionID string, comment string) (*types.RuntimeSession, error)
	GetTodo(sessionID string) (*types.RuntimeTodoArtifact, error)
	EditTodo(sessionID string, items []types.RuntimeTodoItem, comment string) (*types.RuntimeSession, error)
	ConfirmTodo(sessionID string, comment string) (*types.RuntimeSession, error)
	CancelSession(sessionID string, reason string) (*types.RuntimeSession, error)
	ListExecutions(sessionID string) ([]types.RuntimeExecution, error)
	GetExecution(sessionID, executionID string) (*types.RuntimeExecution, error)
	RetryExecutions(sessionID string, req types.RetryExecutionsRequest) ([]types.RetryResultItem, *types.RuntimeSession, error)
	ListLogs(sessionID, executionID, todoID, level string, limit, offset int) ([]types.RuntimeLogEntry, int, error)
	GetPreview(sessionID string) (map[string]any, *types.RuntimeSession, error)
}
