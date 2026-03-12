package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"juyu-ai-platform/internal/types"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	s := &PostgresStore{db: db}
	if err := s.init(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *PostgresStore) init() error {
	_, err := s.db.Exec(`
	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		master_agent TEXT,
		scene TEXT,
		status TEXT,
		payload JSONB NOT NULL,
		created_at TIMESTAMPTZ NOT NULL,
		updated_at TIMESTAMPTZ NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
	CREATE INDEX IF NOT EXISTS idx_tasks_scene ON tasks(scene);
	CREATE INDEX IF NOT EXISTS idx_tasks_updated_at ON tasks(updated_at DESC);
	`)
	return err
}

func (s *PostgresStore) Save(task *types.Task) error {
	b, err := json.Marshal(task)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`
	INSERT INTO tasks(id, master_agent, scene, status, payload, created_at, updated_at)
	VALUES($1,$2,$3,$4,$5::jsonb,$6,$7)
	ON CONFLICT(id) DO UPDATE SET
	master_agent=EXCLUDED.master_agent,
	scene=EXCLUDED.scene,
	status=EXCLUDED.status,
	payload=EXCLUDED.payload,
	updated_at=EXCLUDED.updated_at
	`, task.ID, task.MasterAgent, task.Request.Scene, task.Status, string(b), task.CreatedAt, task.UpdatedAt)
	return err
}

func (s *PostgresStore) Get(id string) (*types.Task, error) {
	var payload []byte
	row := s.db.QueryRow(`SELECT payload FROM tasks WHERE id = $1`, id)
	if err := row.Scan(&payload); err != nil {
		return nil, err
	}
	var task types.Task
	if err := json.Unmarshal(payload, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *PostgresStore) List(filter types.TaskFilter) ([]types.Task, error) {
	query := `SELECT payload FROM tasks WHERE 1=1`
	args := make([]any, 0)
	idx := 1
	if filter.Status != "" {
		query += fmt.Sprintf(` AND status = $%d`, idx)
		args = append(args, filter.Status)
		idx++
	}
	if filter.Scene != "" {
		query += fmt.Sprintf(` AND scene = $%d`, idx)
		args = append(args, filter.Scene)
		idx++
	}
	query += ` ORDER BY updated_at DESC`
	if filter.Limit > 0 {
		query += fmt.Sprintf(` LIMIT $%d`, idx)
		args = append(args, filter.Limit)
		idx++
		if filter.Offset > 0 {
			query += fmt.Sprintf(` OFFSET $%d`, idx)
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
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var task types.Task
		if err := json.Unmarshal(payload, &task); err != nil {
			return nil, err
		}
		items = append(items, task)
	}
	return items, nil
}

func DefaultPostgresDSN() string {
	return strings.Join([]string{
		"host=127.0.0.1",
		"port=5432",
		"user=postgres",
		"password=postgres",
		"dbname=juyu",
		"sslmode=disable",
		"timezone=Asia/Shanghai",
	}, " ")
}

func _unused(t time.Time) time.Time { return t }
