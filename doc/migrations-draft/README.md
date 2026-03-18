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
  - includes retry-chain / attempt / queue-reason fields

- `003_init_logs.sql`
  - `logs`
  - supports session / execution / todo level retrieval

- `004_init_agents.sql`
  - `agents`
  - `agent_profiles`

## Notes

- These are **draft** migrations for design convergence, not yet production migrations.
- IDs are currently `text` to stay aligned with API-level ids like `sess_001`, `todo_1`, `exec_001`.
- Status fields use `text + check constraint` instead of PostgreSQL enum types for flexibility in early iterations.
- `depends_on` is stored as `jsonb` in V1 for simplicity.
- `session_artifacts` is a shared artifact table for PRD / Todo / Preview / Final Result.
- `execution_records` preserves retry history instead of mutating failed rows back to queued.
- `logs` keeps `event_type` for product/debug views instead of only raw text messages.

## Recommended next step

After schema review:

1. move these files into the real repo migration directory
2. add rollback/down migrations if needed
3. align generated SQL with the actual migration tool choice
4. add seed/demo rows for a full session -> execution -> retry -> logs walkthrough
