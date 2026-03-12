# API Examples

## 1. 健康检查

```bash
curl http://localhost:8080/healthz
```

## 2. 查看能力列表

```bash
curl http://localhost:8080/abilities
```

## 3. 商品运营场景执行

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

## 4. 比赛项目场景执行

```bash
curl -X POST http://localhost:8080/execute \
  -H 'Content-Type: application/json' \
  -d '{
    "scene": "competition",
    "input": "为阿里AI大赛生成一套项目方案和演示材料",
    "payload": {
      "contest": "alibaba-ai"
    }
  }'
```
