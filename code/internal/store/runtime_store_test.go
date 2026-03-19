package store

import (
	"testing"

	"juyu-ai-platform/internal/types"
)

func TestRuntimeMemoryStoreMainFlow(t *testing.T) {
	s := NewRuntimeMemoryStore()

	sess, err := s.CreateSession(types.CreateSessionInput{
		Type:    "product_url_learning",
		URL:     "https://example.com/p/1",
		Message: "learn this product",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if sess.Status != types.SessionStatusWaitingPrdConfirm {
		t.Fatalf("unexpected initial status: %s", sess.Status)
	}

	sess, err = s.EditPrd(sess.SessionID, map[string]any{"title": "updated title"}, "edit")
	if err != nil {
		t.Fatalf("edit prd: %v", err)
	}
	if sess.PRD == nil || sess.PRD.Version != 2 {
		t.Fatalf("expected prd version 2 after edit, got %+v", sess.PRD)
	}

	sess, err = s.ConfirmPrd(sess.SessionID, "ok")
	if err != nil {
		t.Fatalf("confirm prd: %v", err)
	}
	if sess.Status != types.SessionStatusWaitingTodo {
		t.Fatalf("unexpected status after prd confirm: %s", sess.Status)
	}
	if sess.Todo == nil || len(sess.Todo.Items) == 0 {
		t.Fatalf("expected todo items")
	}
	if sess.Todo.Version != 1 {
		t.Fatalf("expected todo version 1, got %+v", sess.Todo)
	}

	editedItems := append([]types.RuntimeTodoItem{}, sess.Todo.Items...)
	editedItems = editedItems[:len(editedItems)-1]
	sess, err = s.EditTodo(sess.SessionID, editedItems, "edit todo")
	if err != nil {
		t.Fatalf("edit todo: %v", err)
	}
	if sess.Todo == nil || sess.Todo.Version != 2 {
		t.Fatalf("expected todo version 2 after edit, got %+v", sess.Todo)
	}

	sess, err = s.ConfirmTodo(sess.SessionID, "go")
	if err != nil {
		t.Fatalf("confirm todo: %v", err)
	}
	if sess.Status != types.SessionStatusExecuting {
		t.Fatalf("unexpected status after todo confirm: %s", sess.Status)
	}

	execs, _, err := s.ListExecutions(sess.SessionID, "", "", 100, 0)
	if err != nil {
		t.Fatalf("list executions: %v", err)
	}
	if len(execs) == 0 {
		t.Fatalf("expected executions")
	}

	var failed *types.RuntimeExecution
	for i := range execs {
		if execs[i].Status == types.ExecutionStatusFailed {
			failed = &execs[i]
			break
		}
	}
	if failed == nil {
		t.Fatalf("expected one failed execution to retry")
	}

	items, updated, err := s.RetryExecutions(sess.SessionID, types.RetryExecutionsRequest{
		Items:  []types.RetryTarget{{TodoID: failed.TodoID, LatestFailedExecutionID: failed.ExecutionID}},
		Reason: "retry timeout",
	})
	if err != nil {
		t.Fatalf("retry executions: %v", err)
	}
	if len(items) != 1 || !items[0].Accepted {
		t.Fatalf("expected retry accepted, got %+v", items)
	}
	if updated.Status != types.SessionStatusExecuting {
		t.Fatalf("unexpected status after retry: %s", updated.Status)
	}

	logs, total, err := s.ListLogs(sess.SessionID, failed.ExecutionID, "", "", 100, 0)
	if err != nil {
		t.Fatalf("list logs: %v", err)
	}
	if total == 0 || len(logs) == 0 {
		t.Fatalf("expected logs for failed execution")
	}
}
