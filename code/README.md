# JuYu AI Platform Code

当前代码提供一个可继续开发的 Go 原型，包含：

- 多主控 Agent 路由
- 多能力 Agent 注册中心
- 平台 adapter 接口层（当前含 Alibaba / Taobao / Douyin stub 与协议骨架）
- 配置中心驱动绑定关系
- workflow 顺序执行 / 依赖编排 / 条件分支 / 子编排调用
- 同一依赖层的能力可并行执行
- 人工确认节点标记
- HTTP API 服务入口
- 任务流（task_id / 状态 / 查询 / 确认 / cancel / retry）
- 预览数据结构（preview）
- 商品 / 发布 / 确认三类业务对象模型
- 最小前端演示骨架（`/ui/`，含任务概览 / 预览卡片 / 确认操作 / 筛选）
- 可选 API Key 写操作鉴权（`JUYU_API_KEY`）
- 可选操作员身份鉴别（`JUYU_OPERATOR_TOKENS` + `X-Operator-ID` / `X-Operator-Token`）
- `content_gen` / `page_gen` 已抽象为通用 provider 接口（当前支持 `stub` / `nvidia` / `openai_compat`）
- 执行日志与失败状态
- memory / sqlite / postgres 三种任务存储模式
- 更严格的 JSON 请求解析（拒绝未知字段）
- 基础 HTTP 超时配置（header/read/write/idle）

## 目录

- `cmd/platform`：HTTP 服务入口
- `internal/config`：配置加载
- `internal/router`：场景到大 Agent 路由
- `internal/registry`：小 Agent 注册中心
- `internal/orchestrator`：编排执行与任务流
- `internal/store`：任务存储
- `internal/agents`：能力 Agent 实现
- `internal/adapters`：平台适配器接口与实现桩
- `internal/server`：HTTP 接口层
- `internal/types`：公共类型

## 接口

### `GET /healthz`
健康检查（包含 auth / runtime / adapters 运行态摘要）

### `GET /abilities`
查看当前注册的小 Agent 能力列表

### `GET /abilities/metadata`
查看能力元信息（描述、标签、版本等）

### `GET /agents/metadata`
查看主控 Agent 元信息

### `GET /bindings`
查看主控 Agent 与能力 / workflow 绑定关系

### `GET /adapters/health`
查看平台 adapter 健康状态（含 configured / dry_run / missing_fields / base_url_configured / live_ready），并返回聚合 summary（total / healthy / configured / live_ready / dry_run / misconfigured_live）

### `GET /adapters/descriptors`
查看平台 adapter 描述信息（支持动作、必填凭据、当前模式）

### `GET /adapters/credentials`
查看平台凭据配置模型（当前已加载的字段）

### `GET /`
跳转到最小前端演示页

### `GET /ui/`
最小前端演示骨架，可直接创建任务、查看预览、确认继续

### `GET /admin/providers`
查看当前 content/page provider 运行时配置，包含 requested_provider / effective_provider / live_ready / missing_requirements / fallback_active / fallback_reason

### `PUT /admin/providers`
更新当前进程内的 content/page provider 配置（需写权限；凭据仍从环境变量读取）

### `POST /execute`
按场景触发大 Agent 编排执行

### `GET /tasks`
查看任务列表，支持 `status / scene / operator / platform / q / limit / offset`

### `GET /tasks/{task_id}`
查询任务详情、日志、预览、状态

### `GET /tasks/{task_id}/preview`
查看任务预览数据

### `GET /tasks/{task_id}/logs`
查看任务执行日志

### `GET /tasks/{task_id}/status`
查看任务状态摘要

### `POST /tasks/{task_id}/confirm`
人工确认或拒绝任务继续执行

### `POST /tasks/{task_id}/cancel`
取消任务

### `POST /tasks/{task_id}/retry`
重试任务

更完整接口描述见：`../doc/openapi.yaml`

## 鉴权与操作员身份

### 1) 写接口 API Key（可选）

如果设置了 `JUYU_API_KEY`，所有写接口都需要：

- `X-API-Key: <key>`，或
- `Authorization: Bearer <key>`

### 2) 操作员身份（可选但推荐）

如果设置了 `JUYU_OPERATOR_TOKENS`，写接口还需要提供操作员身份头：

```bash
JUYU_OPERATOR_TOKENS='alice:token-a,bob:token-b'
```

请求头：

- `X-Operator-ID: alice`
- `X-Operator-Token: token-a`
- `X-Operator-Name: Alice Zhang`（可选，仅展示）

行为规则：

- `POST /execute` 中 body 的 `operator` 为空时，会自动使用 `X-Operator-ID`
- `POST /tasks/{id}/confirm` 中 body 的 `approver` 为空时，会自动使用 `X-Operator-ID`
- 如果 body 里的 `operator` / `approver` 与 header 身份不一致，请求会被拒绝

## 运行

```bash
export PATH="/root/.openclaw/workspace/.local/go/bin:$PATH"
JUYU_CONFIG=/root/.openclaw/workspace/configs/agents.yaml go run ./cmd/platform
```

使用 SQLite 持久化：

```bash
export PATH="/root/.openclaw/workspace/.local/go/bin:$PATH"
JUYU_STORE=sqlite JUYU_SQLITE_PATH=juyu.db JUYU_CONFIG=/root/.openclaw/workspace/configs/agents.yaml go run ./cmd/platform
```

使用 PostgreSQL：

```bash
export PATH="/root/.openclaw/workspace/.local/go/bin:$PATH"
JUYU_STORE=postgres \
JUYU_PG_DSN='host=127.0.0.1 port=5432 user=postgres password=postgres dbname=juyu sslmode=disable timezone=Asia/Shanghai' \
JUYU_CONFIG=/root/.openclaw/workspace/configs/agents.yaml \
 go run ./cmd/platform
```

使用 Docker Compose 一键运行：

```bash
cp ../.env.example ../.env
# 按需填写 .env 中的 JUYU_API_KEY / OPENAI_COMPAT_* / adapter 凭据

docker compose up --build -d
curl http://127.0.0.1:8080/healthz
```

密钥管理约定：

- `.env.example`：入库，作为模板
- `.env`：本地运行时真实配置，不入库
- prompt template 通过 `env_file: .env` 原样注入，避免 Docker Compose 对 `{{description}}` / `{{platform}}` 这类模板片段做变量插值污染

启用写操作 API Key（可选）：

```bash
JUYU_API_KEY=your-secret-key docker compose up --build -d
curl -H 'X-API-Key: your-secret-key' -H 'Content-Type: application/json' \
  -d '{"scene":"product","input":"demo","payload":{"platform":"alibaba"}}' \
  http://127.0.0.1:8080/execute
```

启用操作员身份校验：

```bash
JUYU_API_KEY=your-secret-key \
JUYU_OPERATOR_TOKENS='walter:operator-token' \
JUYU_CONFIG=/root/.openclaw/workspace/configs/agents.yaml \
 go run ./cmd/platform

curl -H 'X-API-Key: your-secret-key' \
  -H 'X-Operator-ID: walter' \
  -H 'X-Operator-Token: operator-token' \
  -H 'Content-Type: application/json' \
  -d '{"scene":"product","input":"demo","payload":{"platform":"alibaba"}}' \
  http://127.0.0.1:8080/execute
```

配置真实 adapter 凭据并控制 dry-run：

```bash
export JUYU_ADAPTER_DRY_RUN=false
export JUYU_ALIBABA_APP_KEY=xxx
export JUYU_ALIBABA_SECRET=xxx
export JUYU_TAOBAO_APP_KEY=xxx
export JUYU_TAOBAO_SECRET=xxx
export JUYU_DOUYIN_CLIENT_ID=xxx
export JUYU_DOUYIN_CLIENT_SECRET=xxx
```

配置 `content_gen` provider：

```bash
# 可选：stub | nvidia | openai_compat
export CONTENT_GEN_PROVIDER=nvidia
export CONTENT_GEN_MODEL=meta/llama-3.1-405b-instruct
export CONTENT_GEN_TEMPERATURE=0.4
export CONTENT_GEN_MAX_TOKENS=220
```

如果使用 NVIDIA：

```bash
export NVIDIA_URL=https://integrate.api.nvidia.com/v1/chat/completions
export NVIDIA_KEY=your-key
```

如果使用 OpenAI-compatible：

```bash
export OPENAI_COMPAT_URL=https://your-endpoint/v1/chat/completions
export OPENAI_COMPAT_KEY=your-key
export OPENAI_COMPAT_MODEL=gpt-4o-mini
```

如果要调 prompt，也可以直接改环境变量：

```bash
export CONTENT_GEN_SYSTEM_PROMPT='你擅长生成电商商品发布文案，输出准确、简洁、可直接使用。'
export CONTENT_GEN_PROMPT_TEMPLATE='你是电商运营文案助手。请为以下商品生成一段简洁但可直接用于发布页的中文商品文案，控制在120字内。输出纯文本，不要加标题。商品标题：{{title}}。商品描述：{{description}}。目标平台：{{platform}}。'
export PAGE_GEN_PROVIDER=nvidia
export PAGE_GEN_MODEL=meta/llama-3.1-405b-instruct
export PAGE_GEN_SYSTEM_PROMPT='你擅长生成电商商品页面结构，输出 JSON，字段必须稳定。'
export PAGE_GEN_PROMPT_TEMPLATE='请基于以下商品信息生成一个 JSON 页面草图。字段必须包含 title 和 sections，sections 为字符串数组。不要输出 markdown，不要输出解释。商品标题：{{title}}。商品描述：{{description}}。目标平台：{{platform}}。'
```

未配置完整凭据时：
- `GET /adapters/health` 会显示缺哪些字段
- dry-run=true 时仍可走通演示链路
- dry-run=false 时写动作会因缺凭据而失败

## 当前完成态（2026-03）

已完成：
- Go 服务主链路可运行，`go test ./...` 通过
- 8080 Docker 实例已对齐当前代码
- `content_gen` / `page_gen` 已切到 `openai_compat`
- DashScope coding endpoint (`qwen3-coder-plus`) 已实测跑通
- 写接口 API Key 已启用
- `/admin/providers` / `/execute` / `/tasks/{id}/confirm` 已验证可用
- `publish_exec` / `onshelf_exec` 在 `dry_run=true` 下已验证闭环
- `.env` / `.env.example` 已分离，Compose prompt 注入污染问题已修复

当前边界：
- 真实平台 adapter 仍未接入 live 凭据
- `JUYU_ADAPTER_DRY_RUN=true`，因此当前证明的是“系统链路通”，不是“真实平台已可生产发布”
- healthz/runtime 中个别 `enabled` 字段与实际 provider 可用性口径不完全一致，属于观测面问题，不阻塞执行链路

## 下一阶段待办

1. 单平台 live 接入（建议先 Alibaba 或 Taobao）
   - 补齐真实凭据
   - 校准 method/path/base_url
   - 切 `JUYU_ADAPTER_DRY_RUN=false`
   - 跑最小 publish / on_shelf live 验证
2. 密钥治理升级
   - 将生产密钥从 `.env` 迁到更稳妥的 secrets 管理方式
   - 视需要开启 `JUYU_OPERATOR_TOKENS`
3. 观测面修正
   - 校正 healthz/runtime 中 `enabled` 状态表达
   - 增加 provider / adapter live-ready 诊断信息

## 测试

```bash
export PATH="/root/.openclaw/workspace/.local/go/bin:$PATH"
cd /root/.openclaw/workspace/code
go test ./...
```

## 部署与上线准备

部署说明见：`../doc/deploy.md`

当前服务启动时会做配置与运行时校验，包括：

- 配置文件结构合法性
- workflow 引用完整性
- store 参数完整性
- live 模式下 API Key 要求

这可以避免配置缺失时服务“看起来启动了，实际不可用”。

## 当前说明

当前实现还是原型版：
- 小 Agent 仍以 stub/模拟能力为主
- 尚未接入真实模型或真实平台 API
- 已支持基础并行工作流 / 条件分支 / 子编排调用，但还未支持更完整的 DAG 可视化、消息队列、细粒度 RBAC、持久化 operator directory

但整体骨架已经适合作为后续继续开发的基础项目结构。
