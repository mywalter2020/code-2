package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"juyu-ai-platform/internal/types"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	s := &SQLiteStore{db: db}
	if err := s.init(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *SQLiteStore) init() error {
	_, err := s.db.Exec(`
	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		master_agent TEXT,
		scene TEXT,
		status TEXT,
		payload TEXT,
		created_at TEXT,
		updated_at TEXT
	);`)
	return err
}

func (s *SQLiteStore) Save(task *types.Task) error {
	b, err := json.Marshal(task)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`
	INSERT INTO tasks(id, master_agent, scene, status, payload, created_at, updated_at)
	VALUES(?,?,?,?,?,?,?)
	ON CONFLICT(id) DO UPDATE SET
	master_agent=excluded.master_agent,
	scene=excluded.scene,
	status=excluded.status,
	payload=excluded.payload,
	updated_at=excluded.updated_at
	`, task.ID, task.MasterAgent, task.Request.Scene, task.Status, string(b), task.CreatedAt.Format(time.RFC3339Nano), task.UpdatedAt.Format(time.RFC3339Nano))
	return err
}

func (s *SQLiteStore) Get(id string) (*types.Task, error) {
	var payload string
	row := s.db.QueryRow(`SELECT payload FROM tasks WHERE id = ?`, id)
	if err := row.Scan(&payload); err != nil {
		return nil, err
	}
	var task types.Task
	if err := json.Unmarshal([]byte(payload), &task); err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *SQLiteStore) List(filter types.TaskFilter) ([]types.Task, error) {
	query := `SELECT payload FROM tasks WHERE 1=1`
	args := make([]any, 0)
	if filter.Status != "" {
		query += ` AND status = ?`
		args = append(args, filter.Status)
	}
	if filter.Scene != "" {
		query += ` AND scene = ?`
		args = append(args, filter.Scene)
	}
	query += ` ORDER BY datetime(updated_at) DESC`
	if filter.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, filter.Limit)
		if filter.Offset > 0 {
			query += ` OFFSET ?`
			args = append(args, filter.Offset)
		}
	}
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []types.Task{}
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var task types.Task
		if err := json.Unmarshal([]byte(payload), &task); err != nil {
			return nil, err
		}
		items = append(items, task)
	}
	return items, nil
}

func ParseDSN(path string) string {
	if strings.TrimSpace(path) == "" {
		return "file:juyu.db?_journal=WAL"
	}
	return fmt.Sprintf("file:%s?_journal=WAL", path)
}
