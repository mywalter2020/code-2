-- 002_init_todos.sql
-- Agent Runtime V1 draft migration: todo_items + execution_records

begin;

create table if not exists todo_items (
  id text not null,
  session_id text not null references sessions(id) on delete cascade,
  version integer not null,
  title text not null,
  task_type text not null,
  description text,
  status text not null,
  priority integer not null default 100,
  parallel_group text,
  depends_on jsonb not null default '[]'::jsonb,
  acceptance_criteria jsonb not null default '[]'::jsonb,
  assigned_agent text,
  result_snapshot jsonb not null default '{}'::jsonb,
  sort_order integer not null default 0,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  finished_at timestamptz,
  primary key (id, session_id, version),
  constraint chk_todo_items_status check (
    status in (
      'pending',
      'ready',
      'running',
      'blocked',
      'done',
      'failed',
      'canceled'
    )
  ),
  constraint chk_todo_items_priority check (priority >= 0)
);

create index if not exists idx_todo_items_session_version
  on todo_items(session_id, version);

create index if not exists idx_todo_items_status
  on todo_items(session_id, status);

create index if not exists idx_todo_items_parallel_group
  on todo_items(session_id, parallel_group);

create index if not exists idx_todo_items_sort_order
  on todo_items(session_id, version, sort_order);

create table if not exists execution_records (
  id text primary key,
  session_id text not null references sessions(id) on delete cascade,
  todo_id text not null,
  todo_version integer not null,
  executor_agent text not null,
  skill_code text,
  status text not null,
  attempt integer not null default 1,
  input_payload jsonb not null default '{}'::jsonb,
  output_payload jsonb not null default '{}'::jsonb,
  reasoning_summary text,
  error_code text,
  error_message text,
  started_at timestamptz,
  finished_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint fk_execution_records_todo
    foreign key (todo_id, session_id, todo_version)
    references todo_items(id, session_id, version)
    on delete cascade,
  constraint chk_execution_records_status check (
    status in (
      'queued',
      'running',
      'done',
      'failed',
      'canceled'
    )
  ),
  constraint chk_execution_records_attempt check (attempt >= 1)
);

create index if not exists idx_execution_records_session
  on execution_records(session_id);

create index if not exists idx_execution_records_todo
  on execution_records(session_id, todo_id, todo_version);

create index if not exists idx_execution_records_status
  on execution_records(session_id, status);

create index if not exists idx_execution_records_started_at
  on execution_records(started_at desc);

commit;
