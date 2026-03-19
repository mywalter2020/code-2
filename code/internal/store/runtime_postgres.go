package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	"juyu-ai-platform/internal/types"
)

type RuntimePostgresStore struct {
	db *sql.DB
}

func NewRuntimePostgresStore(dsn string) (*RuntimePostgresStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	s := &RuntimePostgresStore{db: db}
	if err := s.init(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *RuntimePostgresStore) init() error {
	_, err := s.db.Exec(`
	CREATE TABLE IF NOT EXISTS runtime_sessions (
		id TEXT PRIMARY KEY,
		status TEXT NOT NULL,
		payload JSONB NOT NULL,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
	);
	CREATE INDEX IF NOT EXISTS idx_runtime_sessions_status ON runtime_sessions(status);
	CREATE INDEX IF NOT EXISTS idx_runtime_sessions_updated_at ON runtime_sessions(updated_at DESC);

	CREATE TABLE IF NOT EXISTS runtime_logs (
		id BIGSERIAL PRIMARY KEY,
		session_id TEXT NOT NULL REFERENCES runtime_sessions(id) ON DELETE CASCADE,
		execution_id TEXT,
		todo_id TEXT,
		level TEXT NOT NULL,
		payload JSONB NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now()
	);
	CREATE INDEX IF NOT EXISTS idx_runtime_logs_session_created_at ON runtime_logs(session_id, created_at);
	CREATE INDEX IF NOT EXISTS idx_runtime_logs_execution_created_at ON runtime_logs(execution_id, created_at);
	`)
	return err
}

func (s *RuntimePostgresStore) CreateSession(input types.CreateSessionInput) (*types.RuntimeSession, error) {
	mem := NewRuntimeMemoryStore()
	sess, err := mem.CreateSession(input)
	if err != nil {
		return nil, err
	}
	if err := s.saveProjection(mem, sess.SessionID); err != nil {
		return nil, err
	}
	return sess, nil
}

func (s *RuntimePostgresStore) GetSession(sessionID string) (*types.RuntimeSession, error) {
	_, sess, err := s.loadProjection(sessionID)
	return sess, err
}

func (s *RuntimePostgresStore) EditPrd(sessionID string, patch map[string]any, comment string) (*types.RuntimeSession, error) {
	mem, _, err := s.loadProjection(sessionID)
	if err != nil {
		return nil, err
	}
	sess, err := mem.EditPrd(sessionID, patch, comment)
	if err != nil {
		return nil, err
	}
	return sess, s.saveProjection(mem, sessionID)
}

func (s *RuntimePostgresStore) ConfirmPrd(sessionID string, comment string) (*types.RuntimeSession, error) {
	mem, _, err := s.loadProjection(sessionID)
	if err != nil {
		return nil, err
	}
	sess, err := mem.ConfirmPrd(sessionID, comment)
	if err != nil {
		return nil, err
	}
	return sess, s.saveProjection(mem, sessionID)
}

func (s *RuntimePostgresStore) GetTodo(sessionID string) (*types.RuntimeTodoArtifact, error) {
	mem, _, err := s.loadProjection(sessionID)
	if err != nil {
		return nil, err
	}
	return mem.GetTodo(sessionID)
}

func (s *RuntimePostgresStore) EditTodo(sessionID string, items []types.RuntimeTodoItem, comment string) (*types.RuntimeSession, error) {
	mem, _, err := s.loadProjection(sessionID)
	if err != nil {
		return nil, err
	}
	sess, err := mem.EditTodo(sessionID, items, comment)
	if err != nil {
		return nil, err
	}
	return sess, s.saveProjection(mem, sessionID)
}

func (s *RuntimePostgresStore) ConfirmTodo(sessionID string, comment string) (*types.RuntimeSession, error) {
	mem, _, err := s.loadProjection(sessionID)
	if err != nil {
		return nil, err
	}
	sess, err := mem.ConfirmTodo(sessionID, comment)
	if err != nil {
		return nil, err
	}
	return sess, s.saveProjection(mem, sessionID)
}

func (s *RuntimePostgresStore) CancelSession(sessionID string, reason string) (*types.RuntimeSession, error) {
	mem, _, err := s.loadProjection(sessionID)
	if err != nil {
		return nil, err
	}
	sess, err := mem.CancelSession(sessionID, reason)
	if err != nil {
		return nil, err
	}
	return sess, s.saveProjection(mem, sessionID)
}

func (s *RuntimePostgresStore) ListExecutions(sessionID string) ([]types.RuntimeExecution, error) {
	mem, _, err := s.loadProjection(sessionID)
	if err != nil {
		return nil, err
	}
	return mem.ListExecutions(sessionID)
}

func (s *RuntimePostgresStore) GetExecution(sessionID, executionID string) (*types.RuntimeExecution, error) {
	mem, _, err := s.loadProjection(sessionID)
	if err != nil {
		return nil, err
	}
	return mem.GetExecution(sessionID, executionID)
}

func (s *RuntimePostgresStore) RetryExecutions(sessionID string, req types.RetryExecutionsRequest) ([]types.RetryResultItem, *types.RuntimeSession, error) {
	mem, _, err := s.loadProjection(sessionID)
	if err != nil {
		return nil, nil, err
	}
	items, sess, err := mem.RetryExecutions(sessionID, req)
	if err != nil {
		return nil, nil, err
	}
	if err := s.saveProjection(mem, sessionID); err != nil {
		return nil, nil, err
	}
	return items, sess, nil
}

func (s *RuntimePostgresStore) ListLogs(sessionID, executionID, todoID, level string, limit, offset int) ([]types.RuntimeLogEntry, int, error) {
	mem, _, err := s.loadProjection(sessionID)
	if err != nil {
		return nil, 0, err
	}
	return mem.ListLogs(sessionID, executionID, todoID, level, limit, offset)
}

func (s *RuntimePostgresStore) GetPreview(sessionID string) (map[string]any, *types.RuntimeSession, error) {
	mem, _, err := s.loadProjection(sessionID)
	if err != nil {
		return nil, nil, err
	}
	return mem.GetPreview(sessionID)
}

func (s *RuntimePostgresStore) loadProjection(sessionID string) (*RuntimeMemoryStore, *types.RuntimeSession, error) {
	var payload []byte
	row := s.db.QueryRow(`SELECT payload FROM runtime_sessions WHERE id = $1`, sessionID)
	if err := row.Scan(&payload); err != nil {
		return nil, nil, err
	}
	var stored runtimeSessionPayload
	if err := json.Unmarshal(payload, &stored); err != nil {
		return nil, nil, err
	}
	mem := NewRuntimeMemoryStore()
	mem.sessions[sessionID] = cloneRuntimeSession(stored.Session)
	mem.executions[sessionID] = append([]types.RuntimeExecution{}, stored.Executions...)
	mem.logs[sessionID] = append([]types.RuntimeLogEntry{}, stored.Logs...)
	mem.sessionCounter = maxSessionCounterID(sessionID)
	mem.execCounter = maxExecutionCounter(stored.Executions)
	mem.logCounter = maxLogCounter(stored.Logs)
	return mem, cloneRuntimeSession(stored.Session), nil
}

func (s *RuntimePostgresStore) saveProjection(mem *RuntimeMemoryStore, sessionID string) error {
	sess, ok := mem.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	payload := runtimeSessionPayload{
		Session:    cloneRuntimeSession(sess),
		Executions: append([]types.RuntimeExecution{}, mem.executions[sessionID]...),
		Logs:       append([]types.RuntimeLogEntry{}, mem.logs[sessionID]...),
	}
	blob, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`
	INSERT INTO runtime_sessions(id, status, payload, updated_at)
	VALUES($1, $2, $3::jsonb, now())
	ON CONFLICT(id) DO UPDATE SET
	status = EXCLUDED.status,
	payload = EXCLUDED.payload,
	updated_at = EXCLUDED.updated_at
	`, sessionID, string(sess.Status), string(blob))
	if err != nil {
		return err
	}
	return s.replaceLogs(sessionID, mem.logs[sessionID])
}

func (s *RuntimePostgresStore) replaceLogs(sessionID string, logs []types.RuntimeLogEntry) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Exec(`DELETE FROM runtime_logs WHERE session_id = $1`, sessionID); err != nil {
		return err
	}
	for _, item := range logs {
		blob, err := json.Marshal(item)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`
			INSERT INTO runtime_logs(session_id, execution_id, todo_id, level, payload)
			VALUES($1,$2,$3,$4,$5::jsonb)
		`, sessionID, nullIfEmpty(item.ExecutionID), nullIfEmpty(item.TodoID), item.Level, string(blob)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

type runtimeSessionPayload struct {
	Session    *types.RuntimeSession    `json:"session"`
	Executions []types.RuntimeExecution `json:"executions"`
	Logs       []types.RuntimeLogEntry  `json:"logs"`
}

func nullIfEmpty(v string) any {
	if v == "" {
		return nil
	}
	return v
}

func maxExecutionCounter(items []types.RuntimeExecution) int {
	max := 0
	for _, item := range items {
		var n int
		_, _ = fmt.Sscanf(strings.TrimSpace(item.ExecutionID), "exec_%d", &n)
		if n > max {
			max = n
		}
	}
	return max
}

func maxSessionCounterID(id string) int {
	var n int
	_, _ = fmt.Sscanf(strings.TrimSpace(id), "sess_%d", &n)
	return n
}

func maxLogCounter(items []types.RuntimeLogEntry) int64 {
	var max int64
	for _, item := range items {
		if item.LogID > max {
			max = item.LogID
		}
	}
	return max
}
