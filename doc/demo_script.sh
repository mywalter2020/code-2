#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"

echo "[1] healthz"
curl -s "$BASE_URL/healthz" | jq . || curl -s "$BASE_URL/healthz"
echo

echo "[2] create task"
CREATE=$(curl -s -X POST "$BASE_URL/execute" \
  -H 'Content-Type: application/json' \
  -d '{"scene":"product","input":"演示脚本商品上架任务","payload":{"platform":"alibaba"}}')
echo "$CREATE" | jq . || echo "$CREATE"
TASK_ID=$(printf '%s' "$CREATE" | sed -n 's/.*"task_id":"\([^"]*\)".*/\1/p')
echo

echo "[3] preview"
curl -s "$BASE_URL/tasks/$TASK_ID/preview" | jq . || curl -s "$BASE_URL/tasks/$TASK_ID/preview"
echo

echo "[4] logs"
curl -s "$BASE_URL/tasks/$TASK_ID/logs" | jq . || curl -s "$BASE_URL/tasks/$TASK_ID/logs"
echo

echo "[5] confirm"
curl -s -X POST "$BASE_URL/tasks/$TASK_ID/confirm" \
  -H 'Content-Type: application/json' \
  -d '{"approved":true,"comment":"demo script confirm"}' | jq . || true
echo

echo "[6] list tasks"
curl -s "$BASE_URL/tasks?scene=product&limit=5" | jq . || curl -s "$BASE_URL/tasks?scene=product&limit=5"
