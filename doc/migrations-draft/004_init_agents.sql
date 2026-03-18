-- 004_init_agents.sql
-- Agent Runtime V1 draft migration: agents + agent_profiles

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

commit;
