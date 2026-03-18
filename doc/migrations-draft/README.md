# Agent Runtime Migration Drafts

These SQL files are draft migrations derived from:

- `doc/agent-runtime-pg-schema-draft.md`
- `doc/agent-runtime-api-draft.md`
- `doc/openapi-agent-runtime.yaml`

## Files

- `001_init_sessions.sql`
  - `sessions`
  - `session_artifacts`

- `002_init_todos.sql`
  - `todo_items`
  - `execution_records`

- `003_init_logs_agents.sql`
  - `agents`
  - `agent_profiles`
  - `logs`

## Notes

- These are **draft** migrations for design convergence, not yet production migrations.
- IDs are currently `text` to stay aligned with API-level ids like `sess_001`, `todo_1`, `exec_001`.
- Status fields use `text + check constraint` instead of PostgreSQL enum types for flexibility in early iterations.
- `depends_on` is stored as `jsonb` in V1 for simplicity.
- `session_artifacts` is a shared artifact table for PRD / Todo / Preview / Final Result.

## Recommended next step

After schema review:

1. move these files into the real repo migration directory
2. add rollback/down migrations if needed
3. align generated SQL with the actual migration tool choice
