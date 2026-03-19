package store

import (
	"os"
	"testing"

	"juyu-ai-platform/internal/types"
)

func TestRuntimePostgresStoreRoundTripIfConfigured(t *testing.T) {
	dsn := os.Getenv("JUYU_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("JUYU_TEST_PG_DSN not set")
	}

	s, err := NewRuntimePostgresStore(dsn)
	if err != nil {
		t.Fatalf("new runtime postgres store: %v", err)
	}

	sess, err := s.CreateSession(types.CreateSessionInput{
		Type:    "product_url_learning",
		URL:     "https://example.com/product/pg-test",
		Message: "postgres round trip",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	sess, err = s.EditPrd(sess.SessionID, map[string]any{"title": "PG title updated"}, "edit prd")
	if err != nil {
		t.Fatalf("edit prd: %v", err)
	}
	if sess.PRD == nil || sess.PRD.Version < 2 {
		t.Fatalf("expected prd version bumped, got %+v", sess.PRD)
	}

	sess, err = s.ConfirmPrd(sess.SessionID, "confirm prd")
	if err != nil {
		t.Fatalf("confirm prd: %v", err)
	}
	if sess.Todo == nil || sess.Todo.Version != 1 {
		t.Fatalf("expected todo version 1, got %+v", sess.Todo)
	}

	editedItems := append([]types.RuntimeTodoItem{}, sess.Todo.Items...)
	editedItems = editedItems[:len(editedItems)-1]
	sess, err = s.EditTodo(sess.SessionID, editedItems, "drop final item for version bump test")
	if err != nil {
		t.Fatalf("edit todo: %v", err)
	}
	if sess.Todo == nil || sess.Todo.Version < 2 {
		t.Fatalf("expected todo version bumped, got %+v", sess.Todo)
	}

	sess, err = s.ConfirmTodo(sess.SessionID, "run")
	if err != nil {
		t.Fatalf("confirm todo: %v", err)
	}

	reloaded, err := s.GetSession(sess.SessionID)
	if err != nil {
		t.Fatalf("reload session: %v", err)
	}
	if reloaded.PRD == nil || reloaded.PRD.Version < 2 {
		t.Fatalf("expected reloaded prd version >=2, got %+v", reloaded.PRD)
	}
	if reloaded.Todo == nil || reloaded.Todo.Version < 2 {
		t.Fatalf("expected reloaded todo version >=2, got %+v", reloaded.Todo)
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
		t.Fatalf("expected failed execution")
	}
	if failed.TodoVersion != reloaded.Todo.Version {
		t.Fatalf("expected execution todo version %d, got %d", reloaded.Todo.Version, failed.TodoVersion)
	}

	retryItems, _, err := s.RetryExecutions(sess.SessionID, types.RetryExecutionsRequest{
		Items:  []types.RetryTarget{{TodoID: failed.TodoID, LatestFailedExecutionID: failed.ExecutionID}},
		Reason: "postgres retry",
	})
	if err != nil {
		t.Fatalf("retry executions: %v", err)
	}
	if len(retryItems) != 1 || !retryItems[0].Accepted {
		t.Fatalf("expected retry accepted, got %+v", retryItems)
	}

	logs, total, err := s.ListLogs(sess.SessionID, failed.ExecutionID, "", "", 100, 0)
	if err != nil {
		t.Fatalf("list logs: %v", err)
	}
	if total == 0 || len(logs) == 0 {
		t.Fatalf("expected logs")
	}
}
