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

## 1. 健康检查

```bash
curl http://localhost:8080/healthz
```

## 2. 查看能力列表

```bash
curl http://localhost:8080/abilities
```

## 3. 查看能力元信息

```bash
curl http://localhost:8080/abilities/metadata
```

## 4. 查看主控 Agent 元信息

```bash
curl http://localhost:8080/agents/metadata
```

## 5. 查看绑定关系

```bash
curl http://localhost:8080/bindings
```

## 5.1 查看 adapter 健康

```bash
curl http://localhost:8080/adapters/health
```

## 5.2 查看 adapter 描述信息

```bash
curl http://localhost:8080/adapters/descriptors
```

## 5.3 查看 adapter 凭据加载情况

```bash
curl http://localhost:8080/adapters/credentials
```

## 6. 创建商品运营任务（会进入待确认）

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

## 7. 查看任务列表

```bash
curl http://localhost:8080/tasks
```

## 8. 查看任务汇总

```bash
curl http://localhost:8080/tasks/summary
```

## 9. 查询任务详情

```bash
curl http://localhost:8080/tasks/task-000001
```

## 10. 查看任务预览

```bash
curl http://localhost:8080/tasks/task-000001/preview
```

## 11. 查看任务日志

```bash
curl http://localhost:8080/tasks/task-000001/logs
```

## 12. 查看任务状态

```bash
curl http://localhost:8080/tasks/task-000001/status
```

## 13. 确认任务继续执行

```bash
curl -X POST http://localhost:8080/tasks/task-000001/confirm \
  -H 'Content-Type: application/json' \
  -d '{
    "approved": true,
    "comment": "页面内容确认无误，继续发布"
  }'
```

## 14. 拒绝任务

```bash
curl -X POST http://localhost:8080/tasks/task-000001/confirm \
  -H 'Content-Type: application/json' \
  -d '{
    "approved": false,
    "comment": "标题和卖点需要再调整"
  }'
```

## 15. 模拟发布失败

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

## 16. 取消任务

```bash
curl -X POST http://localhost:8080/tasks/task-000001/cancel \
  -H 'Content-Type: application/json' \
  -d '{
    "comment": "用户主动取消"
  }'
```

## 17. 重试任务

```bash
curl -X POST http://localhost:8080/tasks/task-000001/retry
```
