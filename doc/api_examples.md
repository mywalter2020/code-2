# API Examples

先启动服务：

```bash
export PATH="/root/.openclaw/workspace/.local/go/bin:$PATH"
JUYU_CONFIG=/root/.openclaw/workspace/configs/agents.yaml go run ./cmd/platform
```

如果用 PostgreSQL：

```bash
export PATH="/root/.openclaw/workspace/.local/go/bin:$PATH"
JUYU_STORE=postgres \
JUYU_PG_DSN='host=127.0.0.1 port=5432 user=postgres password=postgres dbname=juyu sslmode=disable timezone=Asia/Shanghai' \
JUYU_CONFIG=/root/.openclaw/workspace/configs/agents.yaml \
go run ./cmd/platform
```

## 当前平台 API 示例

### 1. 健康检查

```bash
curl http://localhost:8080/healthz
```

### 2. 查看能力列表

```bash
curl http://localhost:8080/abilities
```

### 3. 查看能力元信息

```bash
curl http://localhost:8080/abilities/metadata
```

### 4. 查看主控 Agent 元信息

```bash
curl http://localhost:8080/agents/metadata
```

### 5. 查看绑定关系

```bash
curl http://localhost:8080/bindings
```

### 5.1 查看 adapter 健康

```bash
curl http://localhost:8080/adapters/health
```

### 5.2 查看 adapter 描述信息

```bash
curl http://localhost:8080/adapters/descriptors
```

### 5.3 查看 adapter 凭据加载情况

```bash
curl http://localhost:8080/adapters/credentials
```

### 6. 创建商品运营任务（会进入待确认）

```bash
curl -X POST http://localhost:8080/execute \
  -H 'Content-Type: application/json' \
  -d '{
    "scene": "product",
    "input": "在阿里平台生成商品页面并准备上架",
    "payload": {
      "platform": "alibaba"
    }
  }'
```

### 7. 查看任务列表

```bash
curl http://localhost:8080/tasks
```

### 8. 查看任务汇总

```bash
curl http://localhost:8080/tasks/summary
```

### 9. 查询任务详情

```bash
curl http://localhost:8080/tasks/task-000001
```

### 10. 查看任务预览

```bash
curl http://localhost:8080/tasks/task-000001/preview
```

### 11. 查看任务日志

```bash
curl http://localhost:8080/tasks/task-000001/logs
```

### 12. 查看任务状态

```bash
curl http://localhost:8080/tasks/task-000001/status
```

### 13. 确认任务继续执行

```bash
curl -X POST http://localhost:8080/tasks/task-000001/confirm \
  -H 'Content-Type: application/json' \
  -d '{
    "approved": true,
    "comment": "页面内容确认无误，继续发布"
  }'
```

### 14. 拒绝任务

```bash
curl -X POST http://localhost:8080/tasks/task-000001/confirm \
  -H 'Content-Type: application/json' \
  -d '{
    "approved": false,
    "comment": "标题和卖点需要再调整"
  }'
```

### 15. 模拟发布失败

```bash
curl -X POST http://localhost:8080/execute \
  -H 'Content-Type: application/json' \
  -d '{
    "scene": "product",
    "input": "在阿里平台生成商品页面并准备上架",
    "payload": {
      "platform": "alibaba",
      "simulate_error": true
    }
  }'
```

### 16. 取消任务

```bash
curl -X POST http://localhost:8080/tasks/task-000001/cancel \
  -H 'Content-Type: application/json' \
  -d '{
    "comment": "用户主动取消"
  }'
```

### 17. 重试任务

```bash
curl -X POST http://localhost:8080/tasks/task-000001/retry
```

---

# Agent Runtime Draft API Examples

下面这组示例对应 `doc/openapi-agent-runtime.yaml` 的草稿接口，目标是把以下闭环走通：

- create session
- confirm / edit PRD
- confirm / edit Todo
- list executions
- get execution detail
- retry failed execution
- query session / execution logs
- get preview

默认 base URL：

```bash
BASE=http://localhost:8080/api/v1
```

## 1. 创建 session

```bash
curl -X POST "$BASE/sessions" \
  -H 'Content-Type: application/json' \
  -d '{
    "input": {
      "type": "product_url_learning",
      "url": "https://example.com/product/123",
      "message": "我想学习下这个产品，并生成一版商品预览内容",
      "platform": "xiaohongshu",
      "language": "zh-CN",
      "style": "clean"
    }
  }'
```

示例响应：

```json
{
  "success": true,
  "code": "",
  "message": "ok",
  "error": "",
  "data": {
    "session_id": "sess_001",
    "stage": "waiting_prd_confirm",
    "status": "waiting_prd_confirm",
    "message": "PRD 已生成，请确认后进入 Todo 阶段",
    "artifacts": {
      "prd": {
        "version": 1,
        "title": "某产品学习与预览内容生成",
        "background": "用户希望基于商品 URL 理解产品并生成预览素材",
        "goals": [
          "理解产品定位与卖点",
          "生成预览页所需商品内容资产"
        ],
        "requirements": [
          "输出标题、图片描述图内容、轮播图内容、视频脚本",
          "最终返回前端可消费的 preview payload"
        ],
        "open_questions": [
          "是否需要视频脚本",
          "是否只做模拟预览"
        ],
        "markdown": "# PRD\n..."
      }
    },
    "next_actions": [
      "confirm_prd",
      "edit_prd",
      "cancel"
    ]
  }
}
```

## 2. 查看 session 详情

```bash
curl "$BASE/sessions/sess_001"
```

示例响应（注意这里直接返回 session 对象，而不是 `artifacts_summary`）：

```json
{
  "success": true,
  "code": "",
  "message": "ok",
  "error": "",
  "data": {
    "session_id": "sess_001",
    "status": "waiting_prd_confirm",
    "current_stage": "waiting_prd_confirm",
    "input": {
      "type": "product_url_learning",
      "url": "https://example.com/product/123",
      "message": "我想学习下这个产品，并生成一版商品预览内容",
      "platform": "xiaohongshu",
      "language": "zh-CN",
      "style": "clean"
    },
    "prd": {
      "version": 1,
      "title": "产品学习与预览内容生成"
    },
    "message": "PRD 已生成，请确认后进入 Todo 阶段",
    "next_actions": ["confirm_prd", "edit_prd", "cancel"]
  }
}
```

## 3. 编辑 PRD

示例：edit 后返回的 `artifacts.prd.version` 会递增。

```bash
curl -X POST "$BASE/sessions/sess_001/edit-prd" \
  -H 'Content-Type: application/json' \
  -d '{
    "patch": {
      "open_questions": [
        "只做模拟预览，不生成真实图片",
        "视频脚本保留简版"
      ]
    },
    "comment": "把范围收窄一点，先聚焦预览内容"
  }'
```

## 4. 确认 PRD，触发 Todo 生成

```bash
curl -X POST "$BASE/sessions/sess_001/confirm-prd" \
  -H 'Content-Type: application/json' \
  -d '{
    "comment": "PRD OK，进入 Todo 阶段"
  }'
```

示例响应：

```json
{
  "success": true,
  "code": "",
  "message": "ok",
  "error": "",
  "data": {
    "session_id": "sess_001",
    "stage": "waiting_todo_confirm",
    "status": "waiting_todo_confirm",
    "message": "Todo 已生成，请确认后开始执行",
    "artifacts": {
      "todo": {
        "version": 1,
        "items": [
          {
            "id": "todo_1",
            "title": "生成产品标题",
            "type": "title_generation",
            "status": "pending",
            "depends_on": [],
            "parallel_group": "content_assets"
          },
          {
            "id": "todo_2",
            "title": "生成图片描述图内容",
            "type": "feature_image_copy",
            "status": "pending",
            "depends_on": [],
            "parallel_group": "content_assets"
          },
          {
            "id": "todo_3",
            "title": "生成产品轮播图内容",
            "type": "carousel_generation",
            "status": "pending",
            "depends_on": [],
            "parallel_group": "content_assets"
          },
          {
            "id": "todo_4",
            "title": "生成产品视频脚本",
            "type": "video_script_generation",
            "status": "pending",
            "depends_on": [],
            "parallel_group": "content_assets"
          },
          {
            "id": "todo_5",
            "title": "汇总预览页面数据",
            "type": "preview_payload_build",
            "status": "pending",
            "depends_on": ["todo_1", "todo_2", "todo_3", "todo_4"]
          }
        ]
      }
    },
    "next_actions": [
      "confirm_todo",
      "edit_todo",
      "cancel"
    ]
  }
}
```

## 5. 查看当前 Todo

```bash
curl "$BASE/sessions/sess_001/todo"
```

## 6. 编辑 Todo

示例：edit 后返回的 `artifacts.todo.version` 会递增。

```bash
curl -X POST "$BASE/sessions/sess_001/edit-todo" \
  -H 'Content-Type: application/json' \
  -d '{
    "items": [
      {
        "id": "todo_1",
        "title": "生成产品标题",
        "type": "title_generation",
        "status": "pending",
        "depends_on": [],
        "parallel_group": "content_assets"
      },
      {
        "id": "todo_2",
        "title": "生成图片描述图内容",
        "type": "feature_image_copy",
        "status": "pending",
        "depends_on": [],
        "parallel_group": "content_assets"
      },
      {
        "id": "todo_3",
        "title": "生成产品轮播图内容",
        "type": "carousel_generation",
        "status": "pending",
        "depends_on": [],
        "parallel_group": "content_assets"
      },
      {
        "id": "todo_5",
        "title": "汇总预览页面数据",
        "type": "preview_payload_build",
        "status": "pending",
        "depends_on": ["todo_1", "todo_2", "todo_3"]
      }
    ],
    "comment": "先去掉视频脚本，聚焦图文预览"
  }'
```

## 7. 确认 Todo，开始执行

```bash
curl -X POST "$BASE/sessions/sess_001/confirm-todo" \
  -H 'Content-Type: application/json' \
  -d '{
    "comment": "Todo OK，开始执行"
  }'
```

## 8. 查看 execution 列表

```bash
curl "$BASE/sessions/sess_001/executions"
```

示例响应：

```json
{
  "success": true,
  "code": "",
  "message": "ok",
  "error": "",
  "data": {
    "session_id": "sess_001",
    "status": "executing",
    "items": [
      {
        "execution_id": "exec_001",
        "parent_execution_id": null,
        "todo_id": "todo_1",
        "title": "生成产品标题",
        "executor": "executor",
        "skill_code": "title_generation",
        "status": "done",
        "attempt": 1,
        "retry_of_execution_id": null,
        "started_at": "2026-03-19T07:20:00+08:00",
        "finished_at": "2026-03-19T07:20:03+08:00"
      },
      {
        "execution_id": "exec_002",
        "parent_execution_id": null,
        "todo_id": "todo_2",
        "title": "生成图片描述图内容",
        "executor": "executor",
        "skill_code": "feature_image_copy",
        "status": "failed",
        "attempt": 1,
        "retry_of_execution_id": null,
        "started_at": "2026-03-19T07:20:00+08:00",
        "finished_at": "2026-03-19T07:20:08+08:00"
      },
      {
        "execution_id": "exec_003",
        "parent_execution_id": null,
        "todo_id": "todo_3",
        "title": "生成产品轮播图内容",
        "executor": "executor",
        "skill_code": "carousel_generation",
        "status": "queued",
        "attempt": 1,
        "retry_of_execution_id": null,
        "started_at": "2026-03-19T07:20:01+08:00",
        "finished_at": null
      }
    ]
  }
}
```

## 9. 查看单个 execution detail

```bash
curl "$BASE/sessions/sess_001/executions/exec_002"
```

示例响应：

```json
{
  "success": true,
  "code": "",
  "message": "ok",
  "error": "",
  "data": {
    "execution_id": "exec_002",
    "parent_execution_id": null,
    "retry_of_execution_id": null,
    "session_id": "sess_001",
    "todo_id": "todo_2",
    "todo_version": 2,
    "executor": "executor",
    "skill_code": "generate_image_caption",
    "status": "failed",
    "attempt": 1,
    "queue_reason": "todo_confirmed",
    "input": {
      "product_summary": "这是一款便携产品..."
    },
    "output": {},
    "reasoning_summary": "执行时调用内容生成模型超时，未拿到完整结果",
    "error_code": "MODEL_TIMEOUT",
    "error_message": "upstream model timeout after 12s",
    "created_at": "2026-03-19T07:19:59+08:00",
    "started_at": "2026-03-19T07:20:00+08:00",
    "finished_at": "2026-03-19T07:20:08+08:00",
    "metrics": {
      "latency_ms": 8004,
      "input_tokens": 1620,
      "output_tokens": 0
    },
    "logs": [
      {
        "log_id": 101,
        "time": "2026-03-19T07:20:00+08:00",
        "level": "info",
        "source_type": "scheduler",
        "source_code": "runtime",
        "event_type": "execution.queued",
        "message": "execution queued",
        "session_id": "sess_001",
        "execution_id": "exec_002",
        "todo_id": "todo_2",
        "data": {
          "attempt": 1
        }
      },
      {
        "log_id": 102,
        "time": "2026-03-19T07:20:08+08:00",
        "level": "error",
        "source_type": "skill",
        "source_code": "feature_image_copy",
        "event_type": "execution.failed",
        "message": "execution failed",
        "session_id": "sess_001",
        "execution_id": "exec_002",
        "todo_id": "todo_2",
        "data": {
          "error_code": "MODEL_TIMEOUT"
        }
      }
    ]
  }
}
```

## 10. 查看 session 级 logs

```bash
curl "$BASE/sessions/sess_001/logs?limit=100&offset=0"
```

按 level 过滤：

```bash
curl "$BASE/sessions/sess_001/logs?level=error"
```

按 todo 过滤：

```bash
curl "$BASE/sessions/sess_001/logs?todo_id=todo_2"
```

## 11. 查看 execution 级 logs

```bash
curl "$BASE/sessions/sess_001/executions/exec_002/logs?limit=100&offset=0"
```

示例响应：

```json
{
  "success": true,
  "code": "",
  "message": "ok",
  "error": "",
  "data": {
    "session_id": "sess_001",
    "execution_id": "exec_002",
    "items": [
      {
        "log_id": 101,
        "time": "2026-03-19T07:20:00+08:00",
        "level": "info",
        "source_type": "scheduler",
        "source_code": "runtime",
        "event_type": "execution.queued",
        "message": "execution queued",
        "data": {
          "attempt": 1
        }
      },
      {
        "log_id": 102,
        "time": "2026-03-19T07:20:08+08:00",
        "level": "error",
        "source_type": "skill",
        "source_code": "feature_image_copy",
        "event_type": "execution.failed",
        "message": "execution failed",
        "data": {
          "error_code": "MODEL_TIMEOUT"
        }
      }
    ],
    "total": 2
  }
}
```

## 12. 对失败 Todo 发起 retry

```bash
curl -X POST "$BASE/sessions/sess_001/executions/retry" \
  -H 'Content-Type: application/json' \
  -d '{
    "items": [
      {
        "todo_id": "todo_2",
        "latest_failed_execution_id": "exec_002",
        "force": false
      }
    ],
    "reason": "上游模型超时，重试一次"
  }'
```

示例响应：

```json
{
  "success": true,
  "code": "",
  "message": "ok",
  "error": "",
  "data": {
    "session_id": "sess_001",
    "status": "executing",
    "message": "1 execution retry scheduled",
    "items": [
      {
        "todo_id": "todo_2",
        "previous_execution_id": "exec_002",
        "new_execution_id": "exec_004",
        "previous_attempt": 1,
        "new_attempt": 2,
        "accepted": true,
        "reason": null
      }
    ]
  }
}
```

## 13. 再次查看 execution 列表，确认 retry 链出现

```bash
curl "$BASE/sessions/sess_001/executions"
```

你应该能看到新的 execution：

- `exec_004`
- `todo_id = todo_2`
- `attempt = 2`
- `retry_of_execution_id = exec_002`
- `status = queued / running / done`

## 14. 查看 preview payload

```bash
curl "$BASE/sessions/sess_001/preview"
```

示例响应：

```json
{
  "success": true,
  "code": "",
  "message": "ok",
  "error": "",
  "data": {
    "session_id": "sess_001",
    "status": "done",
    "preview": {
      "hero_title": "便携轻巧，随手可用的某产品",
      "hero_subtitle": "围绕核心卖点组织成一版适合预览页展示的商品内容",
      "feature_blocks": [
        {
          "title": "卖点 1",
          "description": "轻便易携带，适合日常通勤与短途出行"
        }
      ],
      "carousel": [
        {
          "title": "场景展示",
          "description": "图文轮播内容示例"
        }
      ],
      "video_section": {
        "hook": "开场一句话抓住注意力",
        "scenes": []
      }
    }
  }
}
```

## 15. 取消 session

```bash
curl -X POST "$BASE/sessions/sess_001/cancel" \
  -H 'Content-Type: application/json' \
  -d '{
    "reason": "用户决定终止这次任务"
  }'
```
