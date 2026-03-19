package store

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"juyu-ai-platform/internal/types"
)

type RuntimeMemoryStore struct {
	mu             sync.RWMutex
	sessions       map[string]*types.RuntimeSession
	executions     map[string][]types.RuntimeExecution
	logs           map[string][]types.RuntimeLogEntry
	sessionCounter int
	execCounter    int
	logCounter     int64
}

func NewRuntimeMemoryStore() *RuntimeMemoryStore {
	return &RuntimeMemoryStore{
		sessions:   map[string]*types.RuntimeSession{},
		executions: map[string][]types.RuntimeExecution{},
		logs:       map[string][]types.RuntimeLogEntry{},
	}
}

func (s *RuntimeMemoryStore) CreateSession(input types.CreateSessionInput) (*types.RuntimeSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionCounter++
	now := time.Now()
	id := fmt.Sprintf("sess_%03d", s.sessionCounter)
	prd := &types.RuntimePRD{
		Version:    1,
		Title:      "产品学习与预览内容生成",
		Background: "用户希望基于 URL 理解产品并生成一版预览内容",
		Goals: []string{
			"理解产品定位与卖点",
			"生成预览页所需内容资产",
		},
		Requirements: []string{
			"输出标题、图片描述图内容、轮播图内容、视频脚本",
			"最终返回前端可消费的 preview payload",
		},
		OpenQuestions: []string{
			"是否需要视频脚本",
			"是否只做模拟预览",
		},
		Markdown: "# PRD\n\n- source url: " + input.URL,
	}
	sess := &types.RuntimeSession{
		SessionID:    id,
		Status:       types.SessionStatusWaitingPrdConfirm,
		CurrentStage: string(types.SessionStatusWaitingPrdConfirm),
		Input:        input,
		PRD:          prd,
		Message:      "PRD 已生成，请确认后进入 Todo 阶段",
		NextActions:  []string{"confirm_prd", "edit_prd", "cancel"},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.sessions[id] = cloneRuntimeSession(sess)
	s.appendLogLocked(id, "", "", "info", "system", "runtime", "session.created", "session created", map[string]any{"status": sess.Status})
	return cloneRuntimeSession(sess), nil
}

func (s *RuntimeMemoryStore) ListSessions(status, inputType string, limit, offset int) ([]types.RuntimeSession, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]types.RuntimeSession, 0, len(s.sessions))
	for _, sess := range s.sessions {
		if status != "" && string(sess.Status) != status {
			continue
		}
		if inputType != "" && sess.Input.Type != inputType {
			continue
		}
		items = append(items, *cloneRuntimeSession(sess))
	}
	sort.Slice(items, func(i, j int) bool { return items[i].UpdatedAt.After(items[j].UpdatedAt) })
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

func (s *RuntimeMemoryStore) GetSession(sessionID string) (*types.RuntimeSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	return cloneRuntimeSession(sess), nil
}

func (s *RuntimeMemoryStore) EditPrd(sessionID string, patch map[string]any, comment string) (*types.RuntimeSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, err := s.mustSessionLocked(sessionID)
	if err != nil {
		return nil, err
	}
	if sess.Status != types.SessionStatusWaitingPrdConfirm {
		return nil, invalidStagef("edit prd requires waiting_prd_confirm, got %s", sess.Status)
	}
	if sess.PRD == nil {
		sess.PRD = &types.RuntimePRD{}
	}
	sess.PRD.Version++
	if v, ok := patch["title"].(string); ok && v != "" {
		sess.PRD.Title = v
	}
	if v, ok := patch["background"].(string); ok && v != "" {
		sess.PRD.Background = v
	}
	if v, ok := patch["markdown"].(string); ok && v != "" {
		sess.PRD.Markdown = v
	}
	sess.Message = "PRD 已更新，等待确认"
	sess.NextActions = []string{"confirm_prd", "edit_prd", "cancel"}
	sess.UpdatedAt = time.Now()
	s.appendLogLocked(sessionID, "", "", "info", "agent", "analyst", "prd.edited", "prd edited", map[string]any{"comment": comment})
	return cloneRuntimeSession(sess), nil
}

func (s *RuntimeMemoryStore) ConfirmPrd(sessionID string, comment string) (*types.RuntimeSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, err := s.mustSessionLocked(sessionID)
	if err != nil {
		return nil, err
	}
	if sess.Status != types.SessionStatusWaitingPrdConfirm {
		return nil, invalidStagef("confirm prd requires waiting_prd_confirm, got %s", sess.Status)
	}
	todo := &types.RuntimeTodoArtifact{Version: 1, Items: []types.RuntimeTodoItem{
		{ID: "todo_1", Title: "生成产品标题", Type: "title_generation", Status: types.TodoStatusPending, ParallelGroup: "content_assets"},
		{ID: "todo_2", Title: "生成图片描述图内容", Type: "feature_image_copy", Status: types.TodoStatusPending, ParallelGroup: "content_assets"},
		{ID: "todo_3", Title: "生成产品轮播图内容", Type: "carousel_generation", Status: types.TodoStatusPending, ParallelGroup: "content_assets"},
		{ID: "todo_4", Title: "生成产品视频脚本", Type: "video_script_generation", Status: types.TodoStatusPending, ParallelGroup: "content_assets"},
		{ID: "todo_5", Title: "汇总预览页面数据", Type: "preview_payload_build", Status: types.TodoStatusPending, DependsOn: []string{"todo_1", "todo_2", "todo_3", "todo_4"}},
	}}
	sess.Todo = todo
	sess.Status = types.SessionStatusWaitingTodo
	sess.CurrentStage = string(types.SessionStatusWaitingTodo)
	sess.Message = "Todo 已生成，请确认后开始执行"
	sess.NextActions = []string{"confirm_todo", "edit_todo", "cancel"}
	sess.UpdatedAt = time.Now()
	s.appendLogLocked(sessionID, "", "", "info", "agent", "planner", "prd.confirmed", "prd confirmed", map[string]any{"comment": comment})
	return cloneRuntimeSession(sess), nil
}

func (s *RuntimeMemoryStore) GetTodo(sessionID string) (*types.RuntimeTodoArtifact, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[sessionID]
	if !ok || sess.Todo == nil {
		return nil, fmt.Errorf("todo not found for session: %s", sessionID)
	}
	return cloneRuntimeTodo(sess.Todo), nil
}

func (s *RuntimeMemoryStore) EditTodo(sessionID string, items []types.RuntimeTodoItem, comment string) (*types.RuntimeSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, err := s.mustSessionLocked(sessionID)
	if err != nil {
		return nil, err
	}
	if sess.Status != types.SessionStatusWaitingTodo {
		return nil, invalidStagef("edit todo requires waiting_todo_confirm, got %s", sess.Status)
	}
	version := 1
	if sess.Todo != nil && sess.Todo.Version > 0 {
		version = sess.Todo.Version + 1
	}
	sess.Todo = &types.RuntimeTodoArtifact{Version: version, Items: append([]types.RuntimeTodoItem{}, items...)}
	sess.Message = "Todo 已更新，等待确认"
	sess.NextActions = []string{"confirm_todo", "edit_todo", "cancel"}
	sess.UpdatedAt = time.Now()
	s.appendLogLocked(sessionID, "", "", "info", "agent", "planner", "todo.edited", "todo edited", map[string]any{"comment": comment})
	return cloneRuntimeSession(sess), nil
}

func (s *RuntimeMemoryStore) ConfirmTodo(sessionID string, comment string) (*types.RuntimeSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, err := s.mustSessionLocked(sessionID)
	if err != nil {
		return nil, err
	}
	if sess.Status != types.SessionStatusWaitingTodo {
		return nil, invalidStagef("confirm todo requires waiting_todo_confirm, got %s", sess.Status)
	}
	if sess.Todo == nil {
		return nil, fmt.Errorf("todo not found for session: %s", sessionID)
	}
	now := time.Now()
	sess.Status = types.SessionStatusExecuting
	sess.CurrentStage = string(types.SessionStatusExecuting)
	sess.Message = "Todo 已确认，执行已启动"
	sess.NextActions = []string{"list_executions", "view_logs", "cancel"}
	sess.UpdatedAt = now

	execs := make([]types.RuntimeExecution, 0, len(sess.Todo.Items))
	for i, item := range sess.Todo.Items {
		s.execCounter++
		execID := fmt.Sprintf("exec_%03d", s.execCounter)
		status := types.ExecutionStatusQueued
		var errCode, errMsg string
		if item.ID == "todo_2" {
			status = types.ExecutionStatusFailed
			errCode = "MODEL_TIMEOUT"
			errMsg = "upstream model timeout after 12s"
		} else if i == len(sess.Todo.Items)-1 {
			status = types.ExecutionStatusQueued
		} else {
			status = types.ExecutionStatusDone
		}
		startedAt := now
		finishedAt := now.Add(2 * time.Second)
		exec := types.RuntimeExecution{
			ExecutionID:      execID,
			SessionID:        sessionID,
			TodoID:           item.ID,
			TodoVersion:      sess.Todo.Version,
			Title:            item.Title,
			Executor:         "executor",
			SkillCode:        item.Type,
			Status:           status,
			Attempt:          1,
			QueueReason:      "todo_confirmed",
			Input:            map[string]any{"url": sess.Input.URL},
			Output:           map[string]any{},
			ReasoningSummary: summaryForExecution(status),
			ErrorCode:        errCode,
			ErrorMessage:     errMsg,
			Metrics:          map[string]any{"latency_ms": 2000},
			CreatedAt:        now,
			StartedAt:        &startedAt,
		}
		if status != types.ExecutionStatusQueued {
			exec.FinishedAt = &finishedAt
		}
		if status == types.ExecutionStatusDone {
			exec.Output = map[string]any{"ok": true}
		}
		if status == types.ExecutionStatusDone {
			s.updateTodoStatusLocked(sess, item.ID, types.TodoStatusDone)
		} else if status == types.ExecutionStatusFailed {
			s.updateTodoStatusLocked(sess, item.ID, types.TodoStatusFailed)
		} else {
			s.updateTodoStatusLocked(sess, item.ID, types.TodoStatusReady)
		}
		execs = append(execs, exec)
		s.appendLogLocked(sessionID, execID, item.ID, "info", "scheduler", "runtime", "execution.queued", "execution queued", map[string]any{"attempt": 1})
		if status == types.ExecutionStatusFailed {
			s.appendLogLocked(sessionID, execID, item.ID, "error", "skill", item.Type, "execution.failed", "execution failed", map[string]any{"error_code": errCode})
		}
	}
	s.executions[sessionID] = execs
	sess.Preview = map[string]any{
		"hero_title":     "生成中的商品预览",
		"hero_subtitle":  "执行已开始，部分内容待补全",
		"feature_blocks": []map[string]any{},
		"carousel":       []map[string]any{},
		"video_section":  map[string]any{},
	}
	return cloneRuntimeSession(sess), nil
}

func (s *RuntimeMemoryStore) CancelSession(sessionID string, reason string) (*types.RuntimeSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, err := s.mustSessionLocked(sessionID)
	if err != nil {
		return nil, err
	}
	sess.Status = types.SessionStatusCanceled
	sess.CurrentStage = string(types.SessionStatusCanceled)
	sess.Message = "session canceled"
	sess.NextActions = nil
	sess.UpdatedAt = time.Now()
	s.appendLogLocked(sessionID, "", "", "warn", "system", "runtime", "session.canceled", "session canceled", map[string]any{"reason": reason})
	return cloneRuntimeSession(sess), nil
}

func (s *RuntimeMemoryStore) ListExecutions(sessionID string) ([]types.RuntimeExecution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := s.executions[sessionID]
	out := append([]types.RuntimeExecution{}, items...)
	return out, nil
}

func (s *RuntimeMemoryStore) GetExecution(sessionID, executionID string) (*types.RuntimeExecution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.executions[sessionID] {
		if item.ExecutionID == executionID {
			cp := item
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("execution not found: %s", executionID)
}

func (s *RuntimeMemoryStore) RetryExecutions(sessionID string, req types.RetryExecutionsRequest) ([]types.RetryResultItem, *types.RuntimeSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, err := s.mustSessionLocked(sessionID)
	if err != nil {
		return nil, nil, err
	}
	if sess.Status != types.SessionStatusExecuting {
		return nil, nil, invalidStagef("retry executions requires executing, got %s", sess.Status)
	}
	results := make([]types.RetryResultItem, 0, len(req.Items))
	for _, item := range req.Items {
		prev, idx := latestExecutionForTodo(s.executions[sessionID], item.TodoID)
		if idx < 0 {
			results = append(results, types.RetryResultItem{TodoID: item.TodoID, Accepted: false, Reason: "todo execution not found"})
			continue
		}
		if prev.Status != types.ExecutionStatusFailed && !item.Force {
			results = append(results, types.RetryResultItem{TodoID: item.TodoID, PreviousExecutionID: prev.ExecutionID, PreviousAttempt: prev.Attempt, Accepted: false, Reason: "latest execution is not failed"})
			continue
		}
		s.execCounter++
		now := time.Now()
		newExec := prev
		newExec.ExecutionID = fmt.Sprintf("exec_%03d", s.execCounter)
		newExec.RetryOfExecutionID = prev.ExecutionID
		newExec.Status = types.ExecutionStatusQueued
		newExec.Attempt = prev.Attempt + 1
		newExec.QueueReason = "retry_requested"
		newExec.ErrorCode = ""
		newExec.ErrorMessage = ""
		newExec.CreatedAt = now
		newExec.StartedAt = nil
		newExec.FinishedAt = nil
		newExec.Output = map[string]any{}
		s.executions[sessionID] = append(s.executions[sessionID], newExec)
		s.updateTodoStatusLocked(sess, item.TodoID, types.TodoStatusReady)
		s.appendLogLocked(sessionID, newExec.ExecutionID, item.TodoID, "info", "scheduler", "runtime", "execution.retry_scheduled", "execution retry scheduled", map[string]any{"previous_execution_id": prev.ExecutionID, "reason": req.Reason})
		results = append(results, types.RetryResultItem{TodoID: item.TodoID, PreviousExecutionID: prev.ExecutionID, NewExecutionID: newExec.ExecutionID, PreviousAttempt: prev.Attempt, NewAttempt: newExec.Attempt, Accepted: true})
	}
	sess.Status = types.SessionStatusExecuting
	sess.CurrentStage = string(types.SessionStatusExecuting)
	sess.UpdatedAt = time.Now()
	return results, cloneRuntimeSession(sess), nil
}

func (s *RuntimeMemoryStore) ListLogs(sessionID, executionID, todoID, level string, limit, offset int) ([]types.RuntimeLogEntry, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := s.logs[sessionID]
	filtered := make([]types.RuntimeLogEntry, 0, len(items))
	for _, item := range items {
		if executionID != "" && item.ExecutionID != executionID {
			continue
		}
		if todoID != "" && item.TodoID != todoID {
			continue
		}
		if level != "" && !strings.EqualFold(item.Level, level) {
			continue
		}
		filtered = append(filtered, item)
	}
	total := len(filtered)
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
	return append([]types.RuntimeLogEntry{}, filtered[offset:end]...), total, nil
}

func (s *RuntimeMemoryStore) GetPreview(sessionID string) (map[string]any, *types.RuntimeSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[sessionID]
	if !ok {
		return nil, nil, fmt.Errorf("session not found: %s", sessionID)
	}
	preview := map[string]any{}
	for k, v := range sess.Preview {
		preview[k] = v
	}
	return preview, cloneRuntimeSession(sess), nil
}

func (s *RuntimeMemoryStore) mustSessionLocked(sessionID string) (*types.RuntimeSession, error) {
	sess, ok := s.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	return sess, nil
}

func (s *RuntimeMemoryStore) appendLogLocked(sessionID, executionID, todoID, level, sourceType, sourceCode, eventType, message string, data map[string]any) {
	s.logCounter++
	s.logs[sessionID] = append(s.logs[sessionID], types.RuntimeLogEntry{LogID: s.logCounter, Time: time.Now(), Level: level, SourceType: sourceType, SourceCode: sourceCode, EventType: eventType, Message: message, SessionID: sessionID, ExecutionID: executionID, TodoID: todoID, Data: data})
	sort.Slice(s.logs[sessionID], func(i, j int) bool { return s.logs[sessionID][i].Time.Before(s.logs[sessionID][j].Time) })
}

func (s *RuntimeMemoryStore) updateTodoStatusLocked(sess *types.RuntimeSession, todoID string, status types.TodoStatus) {
	if sess.Todo == nil {
		return
	}
	for i := range sess.Todo.Items {
		if sess.Todo.Items[i].ID == todoID {
			sess.Todo.Items[i].Status = status
			return
		}
	}
}

func latestExecutionForTodo(items []types.RuntimeExecution, todoID string) (types.RuntimeExecution, int) {
	var found types.RuntimeExecution
	idx := -1
	for i, item := range items {
		if item.TodoID == todoID {
			if idx < 0 || item.Attempt >= found.Attempt {
				found = item
				idx = i
			}
		}
	}
	return found, idx
}

func cloneRuntimeSession(in *types.RuntimeSession) *types.RuntimeSession {
	if in == nil {
		return nil
	}
	cp := *in
	if in.PRD != nil {
		prd := *in.PRD
		cp.PRD = &prd
	}
	if in.Todo != nil {
		cp.Todo = cloneRuntimeTodo(in.Todo)
	}
	if in.Preview != nil {
		cp.Preview = map[string]any{}
		for k, v := range in.Preview {
			cp.Preview[k] = v
		}
	}
	cp.NextActions = append([]string{}, in.NextActions...)
	return &cp
}

func cloneRuntimeTodo(in *types.RuntimeTodoArtifact) *types.RuntimeTodoArtifact {
	if in == nil {
		return nil
	}
	items := append([]types.RuntimeTodoItem{}, in.Items...)
	return &types.RuntimeTodoArtifact{Version: in.Version, Items: items}
}

func summaryForExecution(status types.ExecutionStatus) string {
	switch status {
	case types.ExecutionStatusFailed:
		return "执行失败，等待 retry 或人工介入"
	case types.ExecutionStatusDone:
		return "执行完成"
	default:
		return "已入队，等待调度"
	}
}
