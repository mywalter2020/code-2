-- 003_init_logs.sql
-- Agent Runtime V1 draft migration: logs

begin;

create table if not exists logs (
  id bigserial primary key,
  session_id text not null references sessions(id) on delete cascade,
  execution_id text references execution_records(id) on delete cascade,
  todo_id text,
  level text not null,
  source_type text not null,
  source_code text not null,
  event_type text not null,
  message text not null,
  data jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now(),
  constraint chk_logs_level check (
    level in ('debug', 'info', 'warn', 'error')
  ),
  constraint chk_logs_source_type check (
    source_type in ('system', 'agent', 'skill', 'scheduler')
  )
);

create index if not exists idx_logs_session_created_at
  on logs(session_id, created_at);

create index if not exists idx_logs_execution_created_at
  on logs(execution_id, created_at);

create index if not exists idx_logs_todo_created_at
  on logs(session_id, todo_id, created_at);

create index if not exists idx_logs_level
  on logs(level);

create index if not exists idx_logs_event_type
  on logs(event_type);

commit;
