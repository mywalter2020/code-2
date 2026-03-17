package store

import (
	"path/filepath"
	"testing"
	"time"

	"juyu-ai-platform/internal/types"
)

func sampleTask(id, status, scene, operator, platform, input string, createdAt time.Time) *types.Task {
	return &types.Task{
		ID:          id,
		MasterAgent: "product_ops",
		Status:      status,
		Operator:    operator,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
		Request: types.Request{
			Scene:    scene,
			Input:    input,
			Operator: operator,
			Payload:  map[string]any{"platform": platform},
		},
	}
}

func TestMemoryStoreListFiltersAndPagination(t *testing.T) {
	s := NewMemoryStore()
	now := time.Now()

	for _, task := range []*types.Task{
		sampleTask("task-1", types.TaskStatusSuccess, "product", "alice", "alibaba", "alpha chair", now.Add(-3*time.Minute)),
		sampleTask("task-2", types.TaskStatusPendingConfirm, "product", "bob", "taobao", "beta desk", now.Add(-2*time.Minute)),
		sampleTask("task-3", types.TaskStatusFailed, "publish", "alice", "alibaba", "gamma lamp", now.Add(-1*time.Minute)),
	} {
		if err := s.Save(task); err != nil {
			t.Fatalf("save task: %v", err)
		}
	}

	items, err := s.List(types.TaskFilter{Operator: "alice", Platform: "alibaba", Limit: 1})
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ID != "task-3" {
		t.Fatalf("expected newest matching task first, got %s", items[0].ID)
	}

	items, err = s.List(types.TaskFilter{Query: "desk"})
	if err != nil {
		t.Fatalf("list by query: %v", err)
	}
	if len(items) != 1 || items[0].ID != "task-2" {
		t.Fatalf("expected task-2 query match, got %+v", items)
	}
}

func TestMemoryStoreReturnsCopies(t *testing.T) {
	s := NewMemoryStore()
	task := sampleTask("task-1", types.TaskStatusSuccess, "product", "alice", "alibaba", "alpha chair", time.Now())
	if err := s.Save(task); err != nil {
		t.Fatalf("save task: %v", err)
	}

	got, err := s.Get("task-1")
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	got.Status = types.TaskStatusFailed
	got.Request.Payload["platform"] = "tampered"

	reloaded, err := s.Get("task-1")
	if err != nil {
		t.Fatalf("reload task: %v", err)
	}
	if reloaded.Status != types.TaskStatusSuccess {
		t.Fatalf("expected stored task status unchanged, got %s", reloaded.Status)
	}
	if reloaded.Request.Payload["platform"] != "alibaba" {
		t.Fatalf("expected stored payload unchanged, got %v", reloaded.Request.Payload["platform"])
	}
}

func TestSQLiteStoreRoundTripAndFilters(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "tasks.db")
	s, err := NewSQLiteStore(ParseDSN(dbPath))
	if err != nil {
		t.Fatalf("new sqlite store: %v", err)
	}

	now := time.Now()
	if err := s.Save(sampleTask("task-1", types.TaskStatusSuccess, "product", "alice", "alibaba", "alpha chair", now)); err != nil {
		t.Fatalf("save task-1: %v", err)
	}
	if err := s.Save(sampleTask("task-2", types.TaskStatusPendingConfirm, "product", "bob", "taobao", "beta desk", now.Add(time.Minute))); err != nil {
		t.Fatalf("save task-2: %v", err)
	}

	got, err := s.Get("task-1")
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if got.Request.Payload["platform"] != "alibaba" {
		t.Fatalf("expected platform alibaba, got %v", got.Request.Payload["platform"])
	}

	items, err := s.List(types.TaskFilter{Status: types.TaskStatusPendingConfirm, Platform: "taobao"})
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(items) != 1 || items[0].ID != "task-2" {
		t.Fatalf("expected only task-2, got %+v", items)
	}
}
