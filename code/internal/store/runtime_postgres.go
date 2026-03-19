package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

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
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		input_type TEXT NOT NULL,
		status TEXT NOT NULL,
		current_stage TEXT NOT NULL,
		user_input JSONB NOT NULL DEFAULT '{}'::jsonb,
		prd_version INTEGER NOT NULL DEFAULT 0,
		todo_version INTEGER NOT NULL DEFAULT 0,
		preview_version INTEGER NOT NULL DEFAULT 0,
		latest_error_code TEXT,
		latest_error_message TEXT,
		created_by TEXT,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		canceled_at TIMESTAMPTZ,
		completed_at TIMESTAMPTZ
	);
	CREATE TABLE IF NOT EXISTS session_artifacts (
		id BIGSERIAL PRIMARY KEY,
		session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
		artifact_type TEXT NOT NULL,
		version INTEGER NOT NULL,
		content_format TEXT NOT NULL DEFAULT 'json',
		content JSONB NOT NULL DEFAULT '{}'::jsonb,
		markdown_content TEXT,
		summary TEXT,
		created_by_agent TEXT,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		is_current BOOLEAN NOT NULL DEFAULT true,
		UNIQUE(session_id, artifact_type, version)
	);
	CREATE TABLE IF NOT EXISTS todo_items (
		id TEXT NOT NULL,
		session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
		version INTEGER NOT NULL,
		title TEXT NOT NULL,
		task_type TEXT NOT NULL,
		description TEXT,
		status TEXT NOT NULL,
		priority INTEGER NOT NULL DEFAULT 100,
		parallel_group TEXT,
		depends_on JSONB NOT NULL DEFAULT '[]'::jsonb,
		acceptance_criteria JSONB NOT NULL DEFAULT '[]'::jsonb,
		assigned_agent TEXT,
		result_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
		sort_order INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		finished_at TIMESTAMPTZ,
		PRIMARY KEY(id, session_id, version)
	);
	CREATE TABLE IF NOT EXISTS execution_records (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
		todo_id TEXT NOT NULL,
		todo_version INTEGER NOT NULL,
		parent_execution_id TEXT REFERENCES execution_records(id) ON DELETE SET NULL,
		retry_of_execution_id TEXT REFERENCES execution_records(id) ON DELETE SET NULL,
		executor_agent TEXT NOT NULL,
		skill_code TEXT,
		status TEXT NOT NULL,
		attempt INTEGER NOT NULL DEFAULT 1,
		queue_reason TEXT,
		input_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
		output_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
		metrics JSONB NOT NULL DEFAULT '{}'::jsonb,
		reasoning_summary TEXT,
		error_code TEXT,
		error_message TEXT,
		created_by TEXT,
		started_at TIMESTAMPTZ,
		finished_at TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		UNIQUE(session_id, todo_id, todo_version, attempt),
		FOREIGN KEY (todo_id, session_id, todo_version)
		  REFERENCES todo_items(id, session_id, version)
		  ON DELETE CASCADE
	);
	CREATE TABLE IF NOT EXISTS logs (
		id BIGSERIAL PRIMARY KEY,
		session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
		execution_id TEXT REFERENCES execution_records(id) ON DELETE CASCADE,
		todo_id TEXT,
		level TEXT NOT NULL,
		source_type TEXT NOT NULL,
		source_code TEXT NOT NULL,
		event_type TEXT NOT NULL,
		message TEXT NOT NULL,
		data JSONB NOT NULL DEFAULT '{}'::jsonb,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now()
	);
	CREATE INDEX IF NOT EXISTS idx_sessions_updated_at ON sessions(updated_at DESC);
	CREATE INDEX IF NOT EXISTS idx_session_artifacts_type_current ON session_artifacts(session_id, artifact_type, is_current);
	CREATE INDEX IF NOT EXISTS idx_todo_items_session_version ON todo_items(session_id, version);
	CREATE INDEX IF NOT EXISTS idx_execution_records_todo ON execution_records(session_id, todo_id, todo_version);
	CREATE INDEX IF NOT EXISTS idx_logs_session_created_at ON logs(session_id, created_at);
	CREATE INDEX IF NOT EXISTS idx_logs_execution_created_at ON logs(execution_id, created_at);
	`)
	return err
}

func (s *RuntimePostgresStore) CreateSession(input types.CreateSessionInput) (*types.RuntimeSession, error) {
	mem := NewRuntimeMemoryStore()
	sess, err := mem.CreateSession(input)
	if err != nil {
		return nil, err
	}
	return sess, s.persistFromMemory(mem, sess.SessionID)
}

func (s *RuntimePostgresStore) ListSessions(status, inputType string, limit, offset int) ([]types.RuntimeSession, int, error) {
	query := `SELECT id FROM sessions WHERE 1=1`
	args := make([]any, 0)
	idx := 1
	if status != "" {
		query += fmt.Sprintf(` AND status = $%d`, idx)
		args = append(args, status)
		idx++
	}
	if inputType != "" {
		query += fmt.Sprintf(` AND input_type = $%d`, idx)
		args = append(args, inputType)
		idx++
	}
	query += ` ORDER BY updated_at DESC`
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]types.RuntimeSession, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, 0, err
		}
		sess, err := s.rebuildSession(id)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *sess)
	}
	total := len(items)
	if offset > total {
		return []types.RuntimeSession{}, total, nil
	}
	if limit <= 0 {
		limit = 20
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return append([]types.RuntimeSession{}, items[offset:end]...), total, nil
}

func (s *RuntimePostgresStore) GetSession(sessionID string) (*types.RuntimeSession, error) {
	return s.rebuildSession(sessionID)
}

func (s *RuntimePostgresStore) EditPrd(sessionID string, patch map[string]any, comment string) (*types.RuntimeSession, error) {
	mem, err := s.loadMemoryFromDB(sessionID)
	if err != nil {
		return nil, err
	}
	sess, err := mem.EditPrd(sessionID, patch, comment)
	if err != nil {
		return nil, err
	}
	return sess, s.persistFromMemory(mem, sessionID)
}

func (s *RuntimePostgresStore) ConfirmPrd(sessionID string, comment string) (*types.RuntimeSession, error) {
	mem, err := s.loadMemoryFromDB(sessionID)
	if err != nil {
		return nil, err
	}
	sess, err := mem.ConfirmPrd(sessionID, comment)
	if err != nil {
		return nil, err
	}
	return sess, s.persistFromMemory(mem, sessionID)
}

func (s *RuntimePostgresStore) GetTodo(sessionID string) (*types.RuntimeTodoArtifact, error) {
	sess, err := s.rebuildSession(sessionID)
	if err != nil {
		return nil, err
	}
	if sess.Todo == nil {
		return nil, fmt.Errorf("todo not found for session: %s", sessionID)
	}
	return cloneRuntimeTodo(sess.Todo), nil
}

func (s *RuntimePostgresStore) EditTodo(sessionID string, items []types.RuntimeTodoItem, comment string) (*types.RuntimeSession, error) {
	mem, err := s.loadMemoryFromDB(sessionID)
	if err != nil {
		return nil, err
	}
	sess, err := mem.EditTodo(sessionID, items, comment)
	if err != nil {
		return nil, err
	}
	return sess, s.persistFromMemory(mem, sessionID)
}

func (s *RuntimePostgresStore) ConfirmTodo(sessionID string, comment string) (*types.RuntimeSession, error) {
	mem, err := s.loadMemoryFromDB(sessionID)
	if err != nil {
		return nil, err
	}
	sess, err := mem.ConfirmTodo(sessionID, comment)
	if err != nil {
		return nil, err
	}
	return sess, s.persistFromMemory(mem, sessionID)
}

func (s *RuntimePostgresStore) CancelSession(sessionID string, reason string) (*types.RuntimeSession, error) {
	mem, err := s.loadMemoryFromDB(sessionID)
	if err != nil {
		return nil, err
	}
	sess, err := mem.CancelSession(sessionID, reason)
	if err != nil {
		return nil, err
	}
	return sess, s.persistFromMemory(mem, sessionID)
}

func (s *RuntimePostgresStore) ListExecutions(sessionID, status, todoID string, limit, offset int) ([]types.RuntimeExecution, int, error) {
	items, err := s.loadExecutions(sessionID)
	if err != nil {
		return nil, 0, err
	}
	filtered := make([]types.RuntimeExecution, 0, len(items))
	for _, item := range items {
		if status != "" && string(item.Status) != status {
			continue
		}
		if todoID != "" && item.TodoID != todoID {
			continue
		}
		filtered = append(filtered, item)
	}
	total := len(filtered)
	if offset > total {
		return []types.RuntimeExecution{}, total, nil
	}
	if limit <= 0 {
		limit = 100
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return append([]types.RuntimeExecution{}, filtered[offset:end]...), total, nil
}

func (s *RuntimePostgresStore) GetExecution(sessionID, executionID string) (*types.RuntimeExecution, error) {
	items, err := s.loadExecutions(sessionID)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.ExecutionID == executionID {
			cp := item
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("execution not found: %s", executionID)
}

func (s *RuntimePostgresStore) RetryExecutions(sessionID string, req types.RetryExecutionsRequest) ([]types.RetryResultItem, *types.RuntimeSession, error) {
	mem, err := s.loadMemoryFromDB(sessionID)
	if err != nil {
		return nil, nil, err
	}
	items, sess, err := mem.RetryExecutions(sessionID, req)
	if err != nil {
		return nil, nil, err
	}
	if err := s.persistFromMemory(mem, sessionID); err != nil {
		return nil, nil, err
	}
	return items, sess, nil
}

func (s *RuntimePostgresStore) ListLogs(sessionID, executionID, todoID, level string, limit, offset int) ([]types.RuntimeLogEntry, int, error) {
	query := `SELECT id, execution_id, todo_id, level, source_type, source_code, event_type, message, data, created_at FROM logs WHERE session_id = $1`
	args := []any{sessionID}
	idx := 2
	if executionID != "" {
		query += fmt.Sprintf(" AND execution_id = $%d", idx)
		args = append(args, executionID)
		idx++
	}
	if todoID != "" {
		query += fmt.Sprintf(" AND todo_id = $%d", idx)
		args = append(args, todoID)
		idx++
	}
	if level != "" {
		query += fmt.Sprintf(" AND level = $%d", idx)
		args = append(args, strings.ToLower(level))
		idx++
	}
	query += " ORDER BY created_at ASC"
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]types.RuntimeLogEntry, 0)
	for rows.Next() {
		var item types.RuntimeLogEntry
		var data []byte
		if err := rows.Scan(&item.LogID, &item.ExecutionID, &item.TodoID, &item.Level, &item.SourceType, &item.SourceCode, &item.EventType, &item.Message, &data, &item.Time); err != nil {
			return nil, 0, err
		}
		item.SessionID = sessionID
		if len(data) > 0 {
			_ = json.Unmarshal(data, &item.Data)
		}
		items = append(items, item)
	}
	total := len(items)
	if offset > total {
		return []types.RuntimeLogEntry{}, total, nil
	}
	if limit <= 0 {
		limit = 100
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return append([]types.RuntimeLogEntry{}, items[offset:end]...), total, nil
}

func (s *RuntimePostgresStore) GetPreview(sessionID string) (map[string]any, *types.RuntimeSession, error) {
	sess, err := s.rebuildSession(sessionID)
	if err != nil {
		return nil, nil, err
	}
	preview := map[string]any{}
	for k, v := range sess.Preview {
		preview[k] = v
	}
	return preview, sess, nil
}

func (s *RuntimePostgresStore) loadMemoryFromDB(sessionID string) (*RuntimeMemoryStore, error) {
	sess, err := s.rebuildSession(sessionID)
	if err != nil {
		return nil, err
	}
	executions, err := s.loadExecutions(sessionID)
	if err != nil {
		return nil, err
	}
	logs, _, err := s.ListLogs(sessionID, "", "", "", 10000, 0)
	if err != nil {
		return nil, err
	}
	mem := NewRuntimeMemoryStore()
	mem.sessions[sessionID] = cloneRuntimeSession(sess)
	mem.executions[sessionID] = executions
	mem.logs[sessionID] = logs
	mem.sessionCounter = maxSessionCounterID(sessionID)
	mem.execCounter = maxExecutionCounter(executions)
	mem.logCounter = maxLogCounter(logs)
	return mem, nil
}

func (s *RuntimePostgresStore) rebuildSession(sessionID string) (*types.RuntimeSession, error) {
	row := s.db.QueryRow(`SELECT input_type, status, current_stage, user_input, created_at, updated_at FROM sessions WHERE id = $1`, sessionID)
	var inputType, status, stage string
	var userInput []byte
	var createdAt, updatedAt time.Time
	if err := row.Scan(&inputType, &status, &stage, &userInput, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	var input types.CreateSessionInput
	_ = json.Unmarshal(userInput, &input)
	input.Type = inputType
	prd, _ := s.loadArtifactPRD(sessionID)
	preview, _ := s.loadArtifactMap(sessionID, "preview")
	todo, _ := s.loadTodo(sessionID)
	return &types.RuntimeSession{
		SessionID:    sessionID,
		Status:       types.SessionStatus(status),
		CurrentStage: stage,
		Input:        input,
		PRD:          prd,
		Todo:         todo,
		Preview:      preview,
		Message:      messageForStatus(types.SessionStatus(status)),
		NextActions:  nextActionsForStatus(types.SessionStatus(status)),
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, nil
}

func (s *RuntimePostgresStore) persistFromMemory(mem *RuntimeMemoryStore, sessionID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	sess, ok := mem.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	inputJSON, _ := json.Marshal(sess.Input)
	_, err = tx.Exec(`
	INSERT INTO sessions(id, input_type, status, current_stage, user_input, prd_version, todo_version, preview_version, latest_error_code, latest_error_message, created_at, updated_at)
	VALUES($1,$2,$3,$4,$5::jsonb,$6,$7,$8,$9,$10,$11,$12)
	ON CONFLICT(id) DO UPDATE SET
	input_type=EXCLUDED.input_type,
	status=EXCLUDED.status,
	current_stage=EXCLUDED.current_stage,
	user_input=EXCLUDED.user_input,
	prd_version=EXCLUDED.prd_version,
	todo_version=EXCLUDED.todo_version,
	preview_version=EXCLUDED.preview_version,
	latest_error_code=EXCLUDED.latest_error_code,
	latest_error_message=EXCLUDED.latest_error_message,
	updated_at=EXCLUDED.updated_at
	`, sess.SessionID, sess.Input.Type, string(sess.Status), sess.CurrentStage, string(inputJSON), versionForArtifact(sess.PRD), versionForTodo(sess.Todo), versionForPreview(sess.Preview), latestErrorCode(mem.executions[sessionID]), latestErrorMessage(mem.executions[sessionID]), nonZeroTime(sess.CreatedAt), time.Now())
	if err != nil {
		return err
	}
	if err := s.replaceArtifactsTx(tx, sess); err != nil {
		return err
	}
	if err := s.replaceTodoItemsTx(tx, sessionID, sess.Todo); err != nil {
		return err
	}
	if err := s.replaceExecutionsTx(tx, sessionID, mem.executions[sessionID]); err != nil {
		return err
	}
	if err := s.replaceLogsTx(tx, sessionID, mem.logs[sessionID]); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *RuntimePostgresStore) replaceArtifactsTx(tx *sql.Tx, sess *types.RuntimeSession) error {
	if _, err := tx.Exec(`DELETE FROM session_artifacts WHERE session_id = $1`, sess.SessionID); err != nil {
		return err
	}
	if sess.PRD != nil {
		content, _ := json.Marshal(sess.PRD)
		_, err := tx.Exec(`INSERT INTO session_artifacts(session_id, artifact_type, version, content_format, content, markdown_content, summary, created_by_agent, is_current) VALUES($1,'prd',$2,'mixed',$3::jsonb,$4,$5,'analyst',true)`, sess.SessionID, versionForArtifact(sess.PRD), string(content), sess.PRD.Markdown, sess.PRD.Title)
		if err != nil {
			return err
		}
	}
	if sess.Todo != nil {
		content, _ := json.Marshal(sess.Todo)
		_, err := tx.Exec(`INSERT INTO session_artifacts(session_id, artifact_type, version, content_format, content, summary, created_by_agent, is_current) VALUES($1,'todo',$2,'json',$3::jsonb,$4,'planner',true)`, sess.SessionID, versionForTodo(sess.Todo), string(content), "todo list")
		if err != nil {
			return err
		}
	}
	if len(sess.Preview) > 0 {
		content, _ := json.Marshal(sess.Preview)
		_, err := tx.Exec(`INSERT INTO session_artifacts(session_id, artifact_type, version, content_format, content, summary, created_by_agent, is_current) VALUES($1,'preview',1,'json',$2::jsonb,$3,'executor',true)`, sess.SessionID, string(content), "preview payload")
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *RuntimePostgresStore) replaceTodoItemsTx(tx *sql.Tx, sessionID string, todo *types.RuntimeTodoArtifact) error {
	if _, err := tx.Exec(`DELETE FROM todo_items WHERE session_id = $1`, sessionID); err != nil {
		return err
	}
	if todo == nil {
		return nil
	}
	for i, item := range todo.Items {
		dependsOn, _ := json.Marshal(item.DependsOn)
		criteria, _ := json.Marshal(item.AcceptanceCriteria)
		result, _ := json.Marshal(item.Result)
		_, err := tx.Exec(`INSERT INTO todo_items(id, session_id, version, title, task_type, description, status, parallel_group, depends_on, acceptance_criteria, result_snapshot, sort_order, updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,$10::jsonb,$11::jsonb,$12,now())`, item.ID, sessionID, versionForTodo(todo), item.Title, item.Type, item.Description, string(item.Status), nullIfEmptyString(item.ParallelGroup), string(dependsOn), string(criteria), string(result), i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *RuntimePostgresStore) replaceExecutionsTx(tx *sql.Tx, sessionID string, items []types.RuntimeExecution) error {
	if _, err := tx.Exec(`DELETE FROM execution_records WHERE session_id = $1`, sessionID); err != nil {
		return err
	}
	for _, item := range items {
		in, _ := json.Marshal(item.Input)
		out, _ := json.Marshal(item.Output)
		metrics, _ := json.Marshal(item.Metrics)
		_, err := tx.Exec(`INSERT INTO execution_records(id, session_id, todo_id, todo_version, parent_execution_id, retry_of_execution_id, executor_agent, skill_code, status, attempt, queue_reason, input_payload, output_payload, metrics, reasoning_summary, error_code, error_message, created_by, started_at, finished_at, created_at, updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12::jsonb,$13::jsonb,$14::jsonb,$15,$16,$17,$18,$19,$20,$21,now())`, item.ExecutionID, sessionID, item.TodoID, coalesceInt(item.TodoVersion, 1), nullIfEmptyString(item.ParentExecutionID), nullIfEmptyString(item.RetryOfExecutionID), nullIfEmptyString(item.Executor), nullIfEmptyString(item.SkillCode), string(item.Status), coalesceInt(item.Attempt, 1), nullIfEmptyString(item.QueueReason), string(in), string(out), string(metrics), item.ReasoningSummary, nullIfEmptyString(item.ErrorCode), nullIfEmptyString(item.ErrorMessage), "system", nullableTime(item.StartedAt), nullableTime(item.FinishedAt), nonZeroTime(item.CreatedAt))
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *RuntimePostgresStore) replaceLogsTx(tx *sql.Tx, sessionID string, items []types.RuntimeLogEntry) error {
	if _, err := tx.Exec(`DELETE FROM logs WHERE session_id = $1`, sessionID); err != nil {
		return err
	}
	for _, item := range items {
		data, _ := json.Marshal(item.Data)
		_, err := tx.Exec(`INSERT INTO logs(session_id, execution_id, todo_id, level, source_type, source_code, event_type, message, data, created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,$10)`, sessionID, nullIfEmptyString(item.ExecutionID), nullIfEmptyString(item.TodoID), strings.ToLower(item.Level), item.SourceType, item.SourceCode, nonEmpty(item.EventType, "runtime.event"), item.Message, string(data), nonZeroTime(item.Time))
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *RuntimePostgresStore) loadArtifactPRD(sessionID string) (*types.RuntimePRD, error) {
	m, err := s.loadArtifactMap(sessionID, "prd")
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, nil
	}
	blob, _ := json.Marshal(m)
	var prd types.RuntimePRD
	if err := json.Unmarshal(blob, &prd); err != nil {
		return nil, err
	}
	return &prd, nil
}

func (s *RuntimePostgresStore) loadArtifactMap(sessionID, artifactType string) (map[string]any, error) {
	row := s.db.QueryRow(`SELECT content FROM session_artifacts WHERE session_id = $1 AND artifact_type = $2 AND is_current = true ORDER BY version DESC LIMIT 1`, sessionID, artifactType)
	var content []byte
	if err := row.Scan(&content); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	out := map[string]any{}
	if len(content) > 0 {
		_ = json.Unmarshal(content, &out)
	}
	return out, nil
}

func (s *RuntimePostgresStore) loadTodo(sessionID string) (*types.RuntimeTodoArtifact, error) {
	rows, err := s.db.Query(`SELECT id, version, title, task_type, description, status, parallel_group, depends_on, acceptance_criteria, result_snapshot FROM todo_items WHERE session_id = $1 AND version = (SELECT COALESCE(MAX(version), 1) FROM todo_items WHERE session_id = $1) ORDER BY sort_order ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.RuntimeTodoItem, 0)
	version := 0
	for rows.Next() {
		var item types.RuntimeTodoItem
		var rowVersion int
		var dependsOn, criteria, result []byte
		if err := rows.Scan(&item.ID, &rowVersion, &item.Title, &item.Type, &item.Description, &item.Status, &item.ParallelGroup, &dependsOn, &criteria, &result); err != nil {
			return nil, err
		}
		version = rowVersion
		_ = json.Unmarshal(dependsOn, &item.DependsOn)
		_ = json.Unmarshal(criteria, &item.AcceptanceCriteria)
		_ = json.Unmarshal(result, &item.Result)
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &types.RuntimeTodoArtifact{Version: version, Items: items}, nil
}

func (s *RuntimePostgresStore) loadExecutions(sessionID string) ([]types.RuntimeExecution, error) {
	rows, err := s.db.Query(`SELECT id, parent_execution_id, retry_of_execution_id, todo_id, todo_version, executor_agent, skill_code, status, attempt, queue_reason, input_payload, output_payload, metrics, reasoning_summary, error_code, error_message, created_at, started_at, finished_at FROM execution_records WHERE session_id = $1 ORDER BY created_at ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.RuntimeExecution, 0)
	for rows.Next() {
		var item types.RuntimeExecution
		var in, out, metrics []byte
		if err := rows.Scan(&item.ExecutionID, &item.ParentExecutionID, &item.RetryOfExecutionID, &item.TodoID, &item.TodoVersion, &item.Executor, &item.SkillCode, &item.Status, &item.Attempt, &item.QueueReason, &in, &out, &metrics, &item.ReasoningSummary, &item.ErrorCode, &item.ErrorMessage, &item.CreatedAt, &item.StartedAt, &item.FinishedAt); err != nil {
			return nil, err
		}
		item.SessionID = sessionID
		_ = json.Unmarshal(in, &item.Input)
		_ = json.Unmarshal(out, &item.Output)
		_ = json.Unmarshal(metrics, &item.Metrics)
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.Before(items[j].CreatedAt) })
	return items, nil
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

func nonZeroTime(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now()
	}
	return t
}

func nullableTime(t *time.Time) any {
	if t == nil || t.IsZero() {
		return nil
	}
	return *t
}

func coalesceInt(v, fallback int) int {
	if v == 0 {
		return fallback
	}
	return v
}

func nullIfEmptyString(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}

func nonEmpty(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func versionForArtifact(prd *types.RuntimePRD) int {
	if prd == nil {
		return 0
	}
	if prd.Version <= 0 {
		return 1
	}
	return prd.Version
}

func versionForTodo(todo *types.RuntimeTodoArtifact) int {
	if todo == nil {
		return 0
	}
	if todo.Version <= 0 {
		return 1
	}
	return todo.Version
}

func versionForPreview(preview map[string]any) int {
	if len(preview) == 0 {
		return 0
	}
	return 1
}

func latestErrorCode(items []types.RuntimeExecution) string {
	for i := len(items) - 1; i >= 0; i-- {
		if items[i].ErrorCode != "" {
			return items[i].ErrorCode
		}
	}
	return ""
}

func latestErrorMessage(items []types.RuntimeExecution) string {
	for i := len(items) - 1; i >= 0; i-- {
		if items[i].ErrorMessage != "" {
			return items[i].ErrorMessage
		}
	}
	return ""
}

func messageForStatus(status types.SessionStatus) string {
	switch status {
	case types.SessionStatusWaitingPrdConfirm:
		return "PRD 已生成，请确认后进入 Todo 阶段"
	case types.SessionStatusWaitingTodo:
		return "Todo 已生成，请确认后开始执行"
	case types.SessionStatusExecuting:
		return "执行已启动"
	case types.SessionStatusCanceled:
		return "session canceled"
	case types.SessionStatusDone:
		return "session done"
	case types.SessionStatusFailed:
		return "session failed"
	default:
		return string(status)
	}
}

func nextActionsForStatus(status types.SessionStatus) []string {
	switch status {
	case types.SessionStatusWaitingPrdConfirm:
		return []string{"confirm_prd", "edit_prd", "cancel"}
	case types.SessionStatusWaitingTodo:
		return []string{"confirm_todo", "edit_todo", "cancel"}
	case types.SessionStatusExecuting:
		return []string{"list_executions", "view_logs", "cancel"}
	default:
		return nil
	}
}
