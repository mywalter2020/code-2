# JuYu AI Platform Code

当前代码提供一个可继续开发的 Go 原型，包含：

- 多主控 Agent 路由
- 多能力 Agent 注册中心
- 平台 adapter 接口层（当前含 Alibaba stub）
- 配置中心驱动绑定关系
- workflow 顺序执行
- 人工确认节点标记
- HTTP API 服务入口
- 任务流（task_id / 状态 / 查询 / 确认 / cancel / retry）
- 预览数据结构（preview）
- 执行日志与失败状态
- memory / sqlite / postgres 三种任务存储模式

## 目录

- `cmd/platform`：HTTP 服务入口
- `internal/config`：配置加载
- `internal/router`：场景到大 Agent 路由
- `internal/registry`：小 Agent 注册中心
- `internal/orchestrator`：编排执行与任务流
- `internal/store`：任务存储（当前为内存版）
- `internal/agents`：能力 Agent 实现
- `internal/adapters`：平台适配器接口与实现桩
- `internal/server`：HTTP 接口层
- `internal/types`：公共类型

## 接口

### `GET /healthz`
健康检查

### `GET /abilities`
查看当前注册的小 Agent 能力列表

### `GET /abilities/metadata`
查看能力元信息（描述、标签、版本等）

### `POST /execute`
按场景触发大 Agent 编排执行

### `GET /tasks`
查看任务列表

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
docker compose up --build -d
curl http://127.0.0.1:8080/healthz
```

## 当前说明

当前实现还是原型版：
- 任务存储仍为内存版
- 小 Agent 仍以 stub/模拟能力为主
- 尚未接入真实模型或真实平台 API
- 尚未支持鉴权、数据库、消息队列、并行工作流

但整体骨架已经适合作为后续继续开发的基础项目结构。
