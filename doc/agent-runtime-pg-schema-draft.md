# Agent Runtime PostgreSQL Schema Draft

## 1. 文档目标

本文档定义 Agent Runtime 的 V1 PostgreSQL 表结构草案。

表结构设计需要服务于以下核心能力：

- Session 驱动的多阶段流程
- PRD / Todo / Preview 等中间产物存储
- Todo 多线并行执行
- Execution 级别的状态追踪
- Agent Profile 元信息记录
- 全流程日志与审计

本草案以当前已确认的 API 和标准案例为基础，优先保证：

- 结构清晰
- 状态可追踪
- 对 V1 足够实用
- 不为未来过度设计

---

## 2. 设计原则

### 2.1 Session-first
所有业务数据围绕 `session` 聚合。

### 2.2 阶段产物显式存储
PRD、Todo、Preview、最终结果都应作为可独立查询的 Artifact 存储，而不是只塞在一张大表里。

### 2.3 Todo 与 Execution 分离
Todo 表示计划与任务项，Execution 表示实际运行记录。二者语义不同，不应混合。

### 2.4 日志单独存储
日志应独立建表，便于按 session / execution / level 检索。

### 2.5 JSONB + 结构化字段混合
核心状态与索引用结构化字段保存，大块可变内容用 JSONB 保存。

### 2.6 优先支持查询与调试
V1 最重要的是能清楚回答：

- 当前 session 到哪一步了
- PRD 是什么
- Todo 是什么
- 哪些执行成功了
- 哪些失败了
- Preview 数据是什么

---

## 3. 枚举建议

V1 不一定强制使用 PostgreSQL enum type，也可以先用 `text + check constraint`。为了灵活起见，建议先用 `text`。

### 3.1 session.status

- `input_received`
- `analysis_done`
- `waiting_prd_confirm`
- `todo_done`
- `waiting_todo_confirm`
- `executing`
- `done`
- `failed`
- `canceled`

### 3.2 todo_items.status

- `pending`
- `ready`
- `running`
- `blocked`
- `done`
- `failed`
- `canceled`

### 3.3 execution_records.status

- `queued`
- `running`
- `done`
- `failed`
- `canceled`

### 3.4 logs.level

- `debug`
- `info`
- `warn`
- `error`

---

## 4. 核心表总览

V1 建议至少包含以下表：

1. `sessions`
2. `session_artifacts`
3. `todo_items`
4. `execution_records`
5. `logs`
6. `agents`
7. `agent_profiles`（可选但建议保留）

如果要更轻，也可以把 `agents` 和 `agent_profiles` 延后，但我建议先预留。

---

## 5. 表结构设计

## 5.1 sessions

### 作用
存储一次完整用户任务会话，是整个系统的根对象。

### 建议字段

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uuid / text | session 主键 |
| input_type | text | 输入类型，如 `product_url_learning` |
| status | text | session 当前状态 |
| current_stage | text | 当前阶段，通常与 status 接近但可单独表达 |
| user_input | jsonb | 原始用户输入 |
| prd_version | integer | 当前 PRD 版本号 |
| todo_version | integer | 当前 Todo 版本号 |
| preview_version | integer | 当前 Preview 版本号 |
| latest_error_code | text | 最近一次错误码 |
| latest_error_message | text | 最近一次错误信息 |
| created_by | text | 发起人，可为空 |
| created_at | timestamptz | 创建时间 |
| updated_at | timestamptz | 更新时间 |
| canceled_at | timestamptz | 取消时间 |
| completed_at | timestamptz | 完成时间 |

### 说明

- `user_input` 用 JSONB 保存原始输入，避免早期频繁改字段。
- `prd_version` / `todo_version` / `preview_version` 用于支持后续编辑与版本化。
- `status` 与 `current_stage` 可以先保持相同，后续如果需要区分可独立使用。

### 索引建议

- `pk_sessions(id)`
- `idx_sessions_status(status)`
- `idx_sessions_input_type(input_type)`
- `idx_sessions_created_at(created_at desc)`

### 示例 SQL

```sql
create table sessions (
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
  completed_at timestamptz
);
```

---

## 5.2 session_artifacts

### 作用
存储 PRD、Todo、Preview、最终汇总等阶段产物。

### 设计思路
不要把所有产物都塞回 `sessions.user_input` 或某个大 JSON 中。单独建 artifact 表更利于：

- 版本管理
- 历史追踪
- 独立查询
- 不同产物类型并存

### 建议字段

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigserial | 主键 |
| session_id | text | 关联 session |
| artifact_type | text | 如 `prd` / `todo` / `preview` / `final_result` |
| version | integer | 版本号 |
| content_format | text | 如 `json` / `markdown` / `mixed` |
| content | jsonb | 结构化内容 |
| markdown_content | text | markdown 正文，可为空 |
| summary | text | 产物摘要 |
| created_by_agent | text | 产出该 artifact 的 agent |
| created_at | timestamptz | 创建时间 |
| updated_at | timestamptz | 更新时间 |
| is_current | boolean | 是否当前生效版本 |

### 唯一约束建议

- `unique(session_id, artifact_type, version)`

### 索引建议

- `idx_session_artifacts_session_id(session_id)`
- `idx_session_artifacts_type(session_id, artifact_type, is_current)`

### 示例 SQL

```sql
create table session_artifacts (
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
  unique (session_id, artifact_type, version)
);
```

### artifact_type 建议

- `prd`
- `todo`
- `preview`
- `product_summary`
- `final_result`

---

## 5.3 todo_items

### 作用
存储某个 session 的 Todo 列表明细。

### 建议字段

| 字段 | 类型 | 说明 |
|---|---|---|
| id | text | todo 主键，如 `todo_1` |
| session_id | text | 所属 session |
| version | integer | 对应 todo 版本 |
| title | text | 任务标题 |
| task_type | text | 如 `title_generation` |
| description | text | 任务描述 |
| status | text | 当前状态 |
| priority | integer | 优先级 |
| parallel_group | text | 并行组名，可为空 |
| depends_on | jsonb | 依赖 todo id 列表 |
| acceptance_criteria | jsonb | 验收标准 |
| assigned_agent | text | 指定 agent，可为空 |
| result_snapshot | jsonb | 当前任务结果快照 |
| sort_order | integer | 用于前端展示顺序 |
| created_at | timestamptz | 创建时间 |
| updated_at | timestamptz | 更新时间 |
| finished_at | timestamptz | 完成时间 |

### 说明

- `version` 用于区分 Todo 被编辑后的版本。
- `depends_on` 先用 JSONB 数组保存，V1 不急着拆关系表。
- `result_snapshot` 用来保留最新任务产物的摘要，便于快速读。

### 索引建议

- `pk_todo_items(id, session_id, version)` 或单独 surrogate key
- `idx_todo_items_session_version(session_id, version)`
- `idx_todo_items_status(session_id, status)`
- `idx_todo_items_parallel_group(session_id, parallel_group)`

### 示例 SQL

```sql
create table todo_items (
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
  primary key (id, session_id, version)
);
```

---

## 5.4 execution_records

### 作用
存储 Todo 的实际执行记录。

### 为什么单独建表
Todo 表示计划，Execution 表示运行实例。一个 Todo 理论上可以：

- 执行一次成功
- 失败后重试多次
- 被不同执行策略处理

所以必须把“任务定义”和“运行记录”分开。

### 建议字段

| 字段 | 类型 | 说明 |
|---|---|---|
| id | text | execution id |
| session_id | text | 所属 session |
| todo_id | text | 对应 todo |
| todo_version | integer | 对应 todo 版本 |
| executor_agent | text | 执行 agent，如 `executor` |
| skill_code | text | 本次主要调用 skill，可为空 |
| status | text | 执行状态 |
| attempt | integer | 第几次尝试 |
| input_payload | jsonb | 执行输入 |
| output_payload | jsonb | 执行输出 |
| reasoning_summary | text | 执行摘要 |
| error_code | text | 错误码 |
| error_message | text | 错误信息 |
| started_at | timestamptz | 开始时间 |
| finished_at | timestamptz | 结束时间 |
| created_at | timestamptz | 创建时间 |
| updated_at | timestamptz | 更新时间 |

### 索引建议

- `pk_execution_records(id)`
- `idx_execution_records_session(session_id)`
- `idx_execution_records_todo(session_id, todo_id, todo_version)`
- `idx_execution_records_status(session_id, status)`
- `idx_execution_records_started_at(started_at desc)`

### 示例 SQL

```sql
create table execution_records (
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
  foreign key (todo_id, session_id, todo_version)
    references todo_items(id, session_id, version)
    on delete cascade
);
```

---

## 5.5 logs

### 作用
记录全流程日志，可关联 session，也可关联 execution。

### 建议字段

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigserial | 主键 |
| session_id | text | 所属 session |
| execution_id | text | 所属 execution，可为空 |
| todo_id | text | 所属 todo，可为空 |
| level | text | 日志级别 |
| source_type | text | `system` / `agent` / `skill` |
| source_code | text | 如 `analyst` / `executor` / `generate_title` |
| message | text | 日志消息 |
| data | jsonb | 扩展数据 |
| created_at | timestamptz | 创建时间 |

### 索引建议

- `idx_logs_session(session_id, created_at)`
- `idx_logs_execution(execution_id, created_at)`
- `idx_logs_level(level)`

### 示例 SQL

```sql
create table logs (
  id bigserial primary key,
  session_id text not null references sessions(id) on delete cascade,
  execution_id text references execution_records(id) on delete cascade,
  todo_id text,
  level text not null,
  source_type text not null,
  source_code text not null,
  message text not null,
  data jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now()
);
```

---

## 5.6 agents

### 作用
记录系统中注册的 Agent 基本信息。

### 建议字段

| 字段 | 类型 | 说明 |
|---|---|---|
| code | text | agent code，主键 |
| name | text | agent 名称 |
| role | text | `analysis` / `planning` / `execution` |
| enabled | boolean | 是否启用 |
| description | text | 描述 |
| created_at | timestamptz | 创建时间 |
| updated_at | timestamptz | 更新时间 |

### 示例 SQL

```sql
create table agents (
  code text primary key,
  name text not null,
  role text not null,
  enabled boolean not null default true,
  description text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);
```

---

## 5.7 agent_profiles

### 作用
缓存或记录 Agent Profile 内容，方便 API 读取与调试。

如果运行时每次都从文件直接加载，也可以不把全文落库。但建议至少保留一张摘要表，后续会更方便。

### 建议字段

| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigserial | 主键 |
| agent_code | text | 关联 agent |
| version | integer | profile 版本 |
| soul_content | text | SOUL.md 内容 |
| identity_content | text | IDENTITY.md 内容 |
| memory_content | text | MEMORY.md 内容 |
| rules_content | text | AGENTS.md 内容 |
| summary | jsonb | 摘要 |
| is_current | boolean | 当前版本 |
| loaded_at | timestamptz | 加载时间 |

### 示例 SQL

```sql
create table agent_profiles (
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
  unique (agent_code, version)
);
```

---

## 6. 关系说明

### 6.1 sessions -> session_artifacts
一个 session 会有多个 artifact：

- PRD
- Todo
- Preview
- Final Result

是一对多关系。

### 6.2 sessions -> todo_items
一个 session 有多个 Todo 项，也是一对多关系。

### 6.3 todo_items -> execution_records
一个 Todo 项可能对应多次执行记录，是一对多关系。

### 6.4 sessions -> logs
一个 session 会产生大量日志，是一对多关系。

### 6.5 agents -> agent_profiles
一个 agent 可有多个 profile 版本，是一对多关系。

---

## 7. 针对标准案例的数据流映射

以“用户输入产品 URL，希望系统理解产品并生成前端预览页”为例，数据流如下：

### 7.1 创建 session
- 写入 `sessions`
- `status = waiting_prd_confirm`

### 7.2 生成 PRD
- 写入 `session_artifacts`
- `artifact_type = prd`
- 更新 `sessions.prd_version`

### 7.3 确认 PRD，生成 Todo
- 写入 `session_artifacts`
- `artifact_type = todo`
- 写入多条 `todo_items`
- 更新 `sessions.todo_version`
- `status = waiting_todo_confirm`

### 7.4 确认 Todo，进入执行
- 更新 `sessions.status = executing`
- 为每个 Todo 创建 `execution_records`
- 执行期间持续写 `logs`

### 7.5 汇总结果，构建 preview
- 写入 `session_artifacts`
- `artifact_type = preview`
- 更新 `sessions.preview_version`

### 7.6 全部完成
- 更新 `sessions.status = done`
- 写入 `final_result` artifact（可选）

---

## 8. 推荐查询场景

表结构设计需要保证以下查询容易实现：

### 8.1 获取当前 session 总览
查询：
- `sessions`
- 当前 `prd` / `todo` / `preview` artifact

### 8.2 获取当前 Todo List
查询：
- `todo_items where session_id = ? and version = current_version`

### 8.3 获取执行进度
查询：
- `execution_records where session_id = ?`
- 按 `status` 聚合

### 8.4 获取失败原因
查询：
- `execution_records.error_code/error_message`
- `logs` 中 `level = error`

### 8.5 获取 Preview Payload
查询：
- `session_artifacts where artifact_type = 'preview' and is_current = true`

---

## 9. V1 可接受的简化

为了尽快进入实现，V1 可以接受以下简化：

### 9.1 depends_on 不拆中间表
先存在 `todo_items.depends_on` 的 JSONB 数组中。

### 9.2 artifact 统一表
PRD / Todo / Preview 先共用 `session_artifacts`，不拆专表。

### 9.3 agent_profiles 可以先做摘要
如果不想一开始就全文落库，可以只存 summary。

### 9.4 logs 不强制关联 todo_id
如果某些日志只跟 session 相关，可只挂 session_id。

---

## 10. 后续可扩展方向

后续若系统继续演进，可考虑：

1. 增加 `todo_dependencies` 关系表
2. 增加 `execution_steps` 表记录更细粒度 ReAct 过程
3. 增加 `session_events` 表记录状态机事件
4. 增加 `attachments` 表存文件、图片、视频引用
5. 增加 `review_records` 表支持执行后 review 阶段
6. 增加 `operator_actions` 表支持更强审计

---

## 11. 建议的初始 migration 划分

建议不要把所有表塞进一个 migration，初始可分成：

### 001_init_sessions.sql
- `sessions`
- `session_artifacts`

### 002_init_todos.sql
- `todo_items`
- `execution_records`

### 003_init_logs_agents.sql
- `logs`
- `agents`
- `agent_profiles`

这样后面调整成本更低。

---

## 12. 结论

V1 的 PG 表结构重点不是“尽量抽象”，而是：

- 围绕 session 管主流程
- 围绕 artifact 管中间产物
- 围绕 todo 管任务计划
- 围绕 execution 管实际执行
- 围绕 logs 管追踪与调试

这套结构已经足够承载当前已确认的主流程：

- 用户输入
- 生成 PRD
- 确认 PRD
- 生成 Todo
- 确认 Todo
- 执行 Todo（支持并行）
- 汇总结果
- 返回 Preview Payload

如果后续再扩 review、test、deploy 等阶段，也仍然能在这套结构上自然演进。
