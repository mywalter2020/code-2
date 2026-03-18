-- 003_init_logs_agents.sql
-- Agent Runtime V1 draft migration: logs + agents + agent_profiles

begin;

create table if not exists agents (
  code text primary key,
  name text not null,
  role text not null,
  enabled boolean not null default true,
  description text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists idx_agents_role
  on agents(role);

create index if not exists idx_agents_enabled
  on agents(enabled);

create table if not exists agent_profiles (
  id bigserial primary key,
  agent_code text not null references agents(code) on delete cascade,
  version integer not null,
  soul_content text,
  identity_content text,
  memory_content text,
  rules_content text,
  summary jsonb not null default '{}'::jsonb,
  is_current boolean not null default true,
  loaded_at timestamptz not null default now(),
  constraint uq_agent_profiles unique (agent_code, version)
);

create index if not exists idx_agent_profiles_current
  on agent_profiles(agent_code, is_current);

create table if not exists logs (
  id bigserial primary key,
  session_id text not null references sessions(id) on delete cascade,
  execution_id text references execution_records(id) on delete cascade,
  todo_id text,
  level text not null,
  source_type text not null,
  source_code text not null,
  message text not null,
  data jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now(),
  constraint chk_logs_level check (
    level in ('debug', 'info', 'warn', 'error')
  ),
  constraint chk_logs_source_type check (
    source_type in ('system', 'agent', 'skill')
  )
);

create index if not exists idx_logs_session_created_at
  on logs(session_id, created_at);

create index if not exists idx_logs_execution_created_at
  on logs(execution_id, created_at);

create index if not exists idx_logs_level
  on logs(level);

commit;
