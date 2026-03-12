# API Examples

先启动服务：

```bash
export PATH="/root/.openclaw/workspace/.local/go/bin:$PATH"
JUYU_CONFIG=/root/.openclaw/workspace/configs/agents.yaml go run ./cmd/platform
```

## 1. 健康检查

```bash
curl http://localhost:8080/healthz
```

## 2. 查看能力列表

```bash
curl http://localhost:8080/abilities
```

## 3. 创建商品运营任务

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

## 4. 查询任务

```bash
curl http://localhost:8080/tasks/task-000001
```

## 5. 确认任务继续执行

```bash
curl -X POST http://localhost:8080/tasks/task-000001/confirm \
  -H 'Content-Type: application/json' \
  -d '{
    "approved": true
  }'
```

## 6. 拒绝任务

```bash
curl -X POST http://localhost:8080/tasks/task-000001/confirm \
  -H 'Content-Type: application/json' \
  -d '{
    "approved": false
  }'
```
