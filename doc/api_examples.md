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

## 3. 创建商品运营任务（会进入待确认）

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

## 4. 查看任务列表

```bash
curl http://localhost:8080/tasks
```

## 5. 查询任务详情

```bash
curl http://localhost:8080/tasks/task-000001
```

## 6. 查看任务预览

```bash
curl http://localhost:8080/tasks/task-000001/preview
```

## 7. 查看任务日志

```bash
curl http://localhost:8080/tasks/task-000001/logs
```

## 8. 查看任务状态

```bash
curl http://localhost:8080/tasks/task-000001/status
```

## 9. 确认任务继续执行

```bash
curl -X POST http://localhost:8080/tasks/task-000001/confirm \
  -H 'Content-Type: application/json' \
  -d '{
    "approved": true,
    "comment": "页面内容确认无误，继续发布"
  }'
```

## 10. 拒绝任务

```bash
curl -X POST http://localhost:8080/tasks/task-000001/confirm \
  -H 'Content-Type: application/json' \
  -d '{
    "approved": false,
    "comment": "标题和卖点需要再调整"
  }'
```

## 11. 模拟发布失败

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

## 12. 取消任务

```bash
curl -X POST http://localhost:8080/tasks/task-000001/cancel \
  -H 'Content-Type: application/json' \
  -d '{
    "comment": "用户主动取消"
  }'
```

## 13. 重试任务

```bash
curl -X POST http://localhost:8080/tasks/task-000001/retry
```
