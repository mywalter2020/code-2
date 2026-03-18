# Agent Runtime API Draft

## 1. 文档目标

本文档定义 Agent Runtime 的 V1 HTTP API 草案。

API 设计围绕以下主链路展开：

1. 用户提交原始需求
2. 系统生成 PRD
3. 用户确认或修改 PRD
4. 系统生成 Todo List
5. 用户确认或修改 Todo List
6. 系统进入执行阶段
7. 系统返回统一结果与 Preview Payload

该草案的重点不是覆盖所有未来能力，而是先把 V1 样板流程约束清楚，让前后端、存储和执行层围绕同一套接口收敛。

---

## 2. API 设计原则

### 2.1 Session-first
所有接口围绕 `session` 展开。一次用户任务对应一个 session。

### 2.2 阶段驱动
API 需要显式体现阶段推进，而不是只暴露一个“run everything”的黑盒接口。

### 2.3 Human-in-the-loop
PRD 和 Todo 都是显式确认节点，必须有对应确认接口。

### 2.4 统一响应结构
尽量保证各接口返回结构统一，便于前端渲染和状态管理。

### 2.5 面向前端消费
最终返回中应包含可直接用于模拟预览页的数据结构。

---

## 3. 基础约定

### 3.1 Base URL

```text
/api/v1
```

### 3.2 Content-Type

```http
Content-Type: application/json
```

### 3.3 统一响应格式

成功响应：

```json
{
  "success": true,
  "data": {},
  "error": null,
  "meta": {}
}
```

失败响应：

```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "INVALID_STAGE",
    "message": "current session stage does not allow this operation"
  },
  "meta": {}
}
```

### 3.4 通用错误码建议

- `INVALID_REQUEST`
- `INVALID_STAGE`
- `SESSION_NOT_FOUND`
- `ARTIFACT_NOT_FOUND`
- `VALIDATION_FAILED`
- `EXECUTION_FAILED`
- `INTERNAL_ERROR`

---

## 4. 核心资源模型

V1 API 围绕以下资源展开：

- `sessions`
- `prd`
- `todo`
- `executions`
- `preview`
- `agents`

---

## 5. Session 状态模型

### 5.1 Session 状态

建议状态：

- `input_received`
- `analysis_done`
- `waiting_prd_confirm`
- `todo_done`
- `waiting_todo_confirm`
- `executing`
- `done`
- `failed`
- `canceled`

### 5.2 Todo 状态

- `pending`
- `ready`
- `running`
- `blocked`
- `done`
- `failed`
- `canceled`

---

## 6. API 列表概览

### Session
- `POST /sessions`
- `GET /sessions/{session_id}`
- `GET /sessions`
- `POST /sessions/{session_id}/cancel`

### PRD
- `GET /sessions/{session_id}/prd`
- `POST /sessions/{session_id}/confirm-prd`
- `POST /sessions/{session_id}/edit-prd`

### Todo
- `GET /sessions/{session_id}/todo`
- `POST /sessions/{session_id}/confirm-todo`
- `POST /sessions/{session_id}/edit-todo`

### Execution
- `GET /sessions/{session_id}/executions`
- `GET /sessions/{session_id}/executions/{execution_id}`
- `POST /sessions/{session_id}/executions/retry`

### Preview
- `GET /sessions/{session_id}/preview`

### Agent Registry
- `GET /agents`
- `GET /agents/{agent_code}`

### System
- `GET /healthz`

---

## 7. 详细接口设计

## 7.1 创建 Session

### `POST /sessions`

用于提交用户原始需求，并启动分析阶段。

#### Request

```json
{
  "input": {
    "type": "product_url_learning",
    "url": "https://example.com/product/123",
    "message": "我想学习下这个产品，你帮我处理下",
    "platform": "xiaohongshu",
    "language": "zh-CN",
    "style": "clean"
  }
}
```

#### Behavior

服务端行为建议：

1. 创建 session
2. 保存用户原始输入
3. 触发 Analyst Agent
4. 生成 PRD 草案
5. 将 session 推进到 `waiting_prd_confirm`

#### Response

```json
{
  "success": true,
  "data": {
    "session_id": "sess_001",
    "stage": "waiting_prd_confirm",
    "status": "waiting_prd_confirm",
    "message": "PRD 已生成，请确认",
    "artifacts": {
      "prd": {
        "title": "某产品学习与内容生成需求",
        "markdown": "# PRD ..."
      }
    },
    "next_actions": [
      "confirm_prd",
      "edit_prd",
      "cancel"
    ]
  },
  "error": null,
  "meta": {}
}
```

---

## 7.2 获取 Session 详情

### `GET /sessions/{session_id}`

用于获取当前 session 的总体状态。

#### Response

```json
{
  "success": true,
  "data": {
    "session_id": "sess_001",
    "status": "waiting_todo_confirm",
    "current_stage": "waiting_todo_confirm",
    "input": {
      "type": "product_url_learning",
      "url": "https://example.com/product/123"
    },
    "artifacts_summary": {
      "has_prd": true,
      "has_todo": true,
      "has_preview": false
    },
    "message": "Todo List 已生成，请确认后进入执行阶段",
    "next_actions": [
      "confirm_todo",
      "edit_todo",
      "cancel"
    ],
    "created_at": "2026-03-18T15:00:00+08:00",
    "updated_at": "2026-03-18T15:05:00+08:00"
  },
  "error": null,
  "meta": {}
}
```

---

## 7.3 获取 Session 列表

### `GET /sessions`

支持按状态、类型、时间范围过滤。

#### Query Params

- `status`
- `type`
- `limit`
- `offset`

#### Response

```json
{
  "success": true,
  "data": {
    "items": [
      {
        "session_id": "sess_001",
        "status": "executing",
        "type": "product_url_learning",
        "created_at": "2026-03-18T15:00:00+08:00"
      }
    ],
    "total": 1
  },
  "error": null,
  "meta": {}
}
```

---

## 7.4 获取 PRD

### `GET /sessions/{session_id}/prd`

返回当前 session 的 PRD 产物。

#### Response

```json
{
  "success": true,
  "data": {
    "session_id": "sess_001",
    "prd": {
      "title": "某产品学习与内容生成需求",
      "background": "...",
      "goals": ["..."],
      "requirements": ["..."],
      "open_questions": ["..."],
      "markdown": "# PRD ..."
    }
  },
  "error": null,
  "meta": {}
}
```

---

## 7.5 确认 PRD

### `POST /sessions/{session_id}/confirm-prd`

用户确认 PRD 后，服务端应触发 Todo Agent 生成 Todo List，并推进到 `waiting_todo_confirm`。

#### Request

```json
{
  "comment": "可以，继续生成 todo list"
}
```

#### Behavior

1. 校验当前 session 必须处于 `waiting_prd_confirm`
2. 记录用户确认意见
3. 触发 Todo Agent
4. 生成 Todo List
5. 更新 session 到 `waiting_todo_confirm`

#### Response

```json
{
  "success": true,
  "data": {
    "session_id": "sess_001",
    "stage": "waiting_todo_confirm",
    "status": "waiting_todo_confirm",
    "message": "Todo List 已生成，请确认",
    "artifacts": {
      "todo": {
        "items": []
      }
    },
    "next_actions": [
      "confirm_todo",
      "edit_todo",
      "cancel"
    ]
  },
  "error": null,
  "meta": {}
}
```

---

## 7.6 编辑 PRD

### `POST /sessions/{session_id}/edit-prd`

允许用户对 PRD 进行补充、修正或覆盖。

#### Request

```json
{
  "patch": {
    "goals": [
      "生成中文电商风格标题",
      "先不生成真实视频，只生成脚本"
    ],
    "open_questions": []
  },
  "comment": "视频先只要脚本"
}
```

#### Response

```json
{
  "success": true,
  "data": {
    "session_id": "sess_001",
    "stage": "waiting_prd_confirm",
    "status": "waiting_prd_confirm",
    "message": "PRD 已更新，请再次确认",
    "artifacts": {
      "prd": {
        "markdown": "# PRD updated ..."
      }
    },
    "next_actions": [
      "confirm_prd",
      "edit_prd",
      "cancel"
    ]
  },
  "error": null,
  "meta": {}
}
```

---

## 7.7 获取 Todo List

### `GET /sessions/{session_id}/todo`

返回当前 session 的 Todo List。

#### Response

```json
{
  "success": true,
  "data": {
    "session_id": "sess_001",
    "todo": {
      "items": [
        {
          "id": "todo_1",
          "title": "生成产品标题",
          "type": "title_generation",
          "status": "pending",
          "depends_on": [],
          "parallel_group": "content_assets"
        }
      ]
    }
  },
  "error": null,
  "meta": {}
}
```

---

## 7.8 确认 Todo List

### `POST /sessions/{session_id}/confirm-todo`

用户确认 Todo 后，服务端开始进入执行阶段。

#### Request

```json
{
  "comment": "可以开始执行"
}
```

#### Behavior

1. 校验当前 session 必须处于 `waiting_todo_confirm`
2. 记录确认意见
3. 启动 Executor Agent
4. 将 session 状态更新为 `executing`

#### Response

```json
{
  "success": true,
  "data": {
    "session_id": "sess_001",
    "stage": "executing",
    "status": "executing",
    "message": "已开始执行 Todo List",
    "next_actions": [
      "view_executions",
      "cancel"
    ]
  },
  "error": null,
  "meta": {}
}
```

---

## 7.9 编辑 Todo List

### `POST /sessions/{session_id}/edit-todo`

允许用户在执行前修改 Todo List。

#### Request

```json
{
  "items": [
    {
      "id": "todo_1",
      "title": "生成产品标题",
      "type": "title_generation",
      "status": "pending"
    },
    {
      "id": "todo_2",
      "title": "生成产品轮播图内容",
      "type": "carousel_generation",
      "status": "pending"
    }
  ],
  "comment": "去掉视频脚本，先保留标题和轮播图"
}
```

#### Response

```json
{
  "success": true,
  "data": {
    "session_id": "sess_001",
    "stage": "waiting_todo_confirm",
    "status": "waiting_todo_confirm",
    "message": "Todo List 已更新，请再次确认",
    "artifacts": {
      "todo": {
        "items": []
      }
    },
    "next_actions": [
      "confirm_todo",
      "edit_todo",
      "cancel"
    ]
  },
  "error": null,
  "meta": {}
}
```

---

## 7.10 获取执行列表

### `GET /sessions/{session_id}/executions`

用于查看每个 Todo 的执行情况。

#### Response

```json
{
  "success": true,
  "data": {
    "session_id": "sess_001",
    "status": "executing",
    "items": [
      {
        "execution_id": "exec_001",
        "todo_id": "todo_1",
        "title": "生成产品标题",
        "executor": "executor",
        "status": "done",
        "started_at": "2026-03-18T15:10:00+08:00",
        "finished_at": "2026-03-18T15:10:05+08:00"
      },
      {
        "execution_id": "exec_002",
        "todo_id": "todo_2",
        "title": "生成图片描述图内容",
        "executor": "executor",
        "status": "running",
        "started_at": "2026-03-18T15:10:01+08:00",
        "finished_at": null
      }
    ]
  },
  "error": null,
  "meta": {}
}
```

---

## 7.11 获取单个执行详情

### `GET /sessions/{session_id}/executions/{execution_id}`

返回某个 Todo 执行项的详细信息。

#### Response

```json
{
  "success": true,
  "data": {
    "execution_id": "exec_001",
    "todo_id": "todo_1",
    "status": "done",
    "input": {
      "title_style": "ecommerce"
    },
    "output": {
      "title": "某产品标题"
    },
    "reasoning_summary": "已根据产品卖点生成标题",
    "logs": [
      {
        "time": "2026-03-18T15:10:00+08:00",
        "level": "info",
        "message": "start title generation"
      }
    ]
  },
  "error": null,
  "meta": {}
}
```

---

## 7.12 重试执行

### `POST /sessions/{session_id}/executions/retry`

用于重试失败的执行项。

#### Request

```json
{
  "todo_ids": ["todo_3", "todo_4"]
}
```

#### Response

```json
{
  "success": true,
  "data": {
    "session_id": "sess_001",
    "status": "executing",
    "message": "已重新调度指定 Todo"
  },
  "error": null,
  "meta": {}
}
```

---

## 7.13 获取 Preview Payload

### `GET /sessions/{session_id}/preview`

返回当前 session 的预览数据结构。

#### Behavior

- 若执行尚未完成，可返回部分 preview
- 若全部完成，返回完整 preview payload

#### Response

```json
{
  "success": true,
  "data": {
    "session_id": "sess_001",
    "status": "done",
    "preview": {
      "hero_title": "某产品标题",
      "hero_subtitle": "产品核心卖点",
      "feature_blocks": [
        {
          "title": "卖点一",
          "description": "..."
        }
      ],
      "carousel": [
        {
          "title": "轮播 1",
          "description": "..."
        }
      ],
      "video_section": {
        "hook": "开场吸引语",
        "scenes": []
      }
    }
  },
  "error": null,
  "meta": {}
}
```

---

## 7.14 取消 Session

### `POST /sessions/{session_id}/cancel`

用于取消当前 session。

#### Request

```json
{
  "reason": "用户主动取消"
}
```

#### Response

```json
{
  "success": true,
  "data": {
    "session_id": "sess_001",
    "status": "canceled",
    "message": "session 已取消"
  },
  "error": null,
  "meta": {}
}
```

---

## 7.15 获取 Agent 列表

### `GET /agents`

用于返回当前系统注册的 agent 信息。

#### Response

```json
{
  "success": true,
  "data": {
    "items": [
      {
        "code": "analyst",
        "name": "Analyst Agent",
        "role": "analysis",
        "profile": {
          "identity": "...",
          "soul": "..."
        }
      },
      {
        "code": "todo",
        "name": "Todo Agent",
        "role": "planning"
      },
      {
        "code": "executor",
        "name": "Executor Agent",
        "role": "execution"
      }
    ]
  },
  "error": null,
  "meta": {}
}
```

---

## 7.16 获取单个 Agent 详情

### `GET /agents/{agent_code}`

返回某个 agent 的详细 profile 摘要。

#### Response

```json
{
  "success": true,
  "data": {
    "code": "executor",
    "name": "Executor Agent",
    "role": "execution",
    "profile": {
      "identity": {
        "name": "Executor"
      },
      "soul_summary": "偏执行导向",
      "memory_summary": "围绕 todo 推进任务",
      "rules_summary": [
        "按 todo list 推进",
        "支持并行",
        "失败时保留日志"
      ]
    }
  },
  "error": null,
  "meta": {}
}
```

---

## 7.17 健康检查

### `GET /healthz`

#### Response

```json
{
  "success": true,
  "data": {
    "status": "ok"
  },
  "error": null,
  "meta": {}
}
```

---

## 8. API 与阶段推进关系

建议阶段推进关系如下：

```text
POST /sessions
  -> waiting_prd_confirm

POST /sessions/{id}/confirm-prd
  -> waiting_todo_confirm

POST /sessions/{id}/confirm-todo
  -> executing

GET /sessions/{id}/preview
  -> done 后返回最终 preview
```

补充说明：

- `edit-prd` 不推进阶段，只刷新 PRD 并停留在 `waiting_prd_confirm`
- `edit-todo` 不推进阶段，只刷新 Todo 并停留在 `waiting_todo_confirm`
- `cancel` 可在未完成前任何阶段触发

---

## 9. 与 V1 标准案例的映射

对于“用户给一个产品 URL，希望系统理解产品、生成素材、返回前端预览页”的案例，API 流程如下：

1. 前端调用 `POST /sessions` 提交 URL
2. 系统生成 PRD，前端展示给用户确认
3. 用户调用 `POST /sessions/{id}/confirm-prd`
4. 系统生成 Todo List，前端展示给用户确认
5. 用户调用 `POST /sessions/{id}/confirm-todo`
6. 前端轮询 `GET /sessions/{id}` 或 `GET /sessions/{id}/executions`
7. 完成后前端调用 `GET /sessions/{id}/preview`
8. 前端据此渲染模拟预览页面

---

## 10. 当前待确认问题

API 草案完成后，建议继续确认以下事项：

1. 是否需要 Webhook / SSE / WebSocket 用于执行进度推送
2. `edit-prd` 和 `edit-todo` 是 patch 语义还是 full replace 语义
3. Preview 是否允许部分返回
4. 是否需要单独的 `GET /sessions/{id}/artifacts` 接口
5. 是否需要对执行并发数提供 API 可见字段
6. 是否需要操作人字段与审计字段

---

## 11. 结论

V1 API 应围绕 Session 状态机和确认节点来设计，而不是直接暴露一个黑盒执行入口。

这样做的价值在于：

- 保证需求理解、Todo 拆解和执行之间边界清晰
- 支持 Human-in-the-loop
- 便于前端逐阶段展示
- 便于后端持久化状态
- 能自然承载并行执行与最终 Preview Payload 输出
