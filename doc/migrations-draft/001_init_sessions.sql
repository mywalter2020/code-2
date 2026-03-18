-- 001_init_sessions.sql
-- Agent Runtime V1 draft migration: sessions + session_artifacts

begin;

create table if not exists sessions (
  id text primary key,
  input_type text not null,
  status text not null,
  current_stage text not null,
  user_input jsonb not null default '{}'::jsonb,
  prd_version integer not null default 0,
  todo_version integer not null default 0,
  preview_version integer not null default 0,
  latest_error_code text,
  latest_error_message text,
  created_by text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  canceled_at timestamptz,
  completed_at timestamptz,
  constraint chk_sessions_status check (
    status in (
      'input_received',
      'analysis_done',
      'waiting_prd_confirm',
      'todo_done',
      'waiting_todo_confirm',
      'executing',
      'done',
      'failed',
      'canceled'
    )
  )
);

create index if not exists idx_sessions_status
  on sessions(status);

create index if not exists idx_sessions_input_type
  on sessions(input_type);

create index if not exists idx_sessions_created_at
  on sessions(created_at desc);

create index if not exists idx_sessions_updated_at
  on sessions(updated_at desc);

create table if not exists session_artifacts (
  id bigserial primary key,
  session_id text not null references sessions(id) on delete cascade,
  artifact_type text not null,
  version integer not null,
  content_format text not null default 'json',
  content jsonb not null default '{}'::jsonb,
  markdown_content text,
  summary text,
  created_by_agent text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  is_current boolean not null default true,
  constraint uq_session_artifacts unique (session_id, artifact_type, version),
  constraint chk_session_artifacts_type check (
    artifact_type in (
      'prd',
      'todo',
      'preview',
      'product_summary',
      'final_result'
    )
  ),
  constraint chk_session_artifacts_format check (
    content_format in ('json', 'markdown', 'mixed')
  )
);

create index if not exists idx_session_artifacts_session_id
  on session_artifacts(session_id);

create index if not exists idx_session_artifacts_type_current
  on session_artifacts(session_id, artifact_type, is_current);

create index if not exists idx_session_artifacts_created_at
  on session_artifacts(created_at desc);

commit;
