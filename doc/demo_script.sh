#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"

pretty_print() {
  if command -v jq >/dev/null 2>&1; then
    jq .
  else
    cat
  fi
}

echo "[1] healthz"
curl -s "$BASE_URL/healthz" | pretty_print
echo

echo "[2] create task"
CREATE=$(curl -s -X POST "$BASE_URL/execute" \
  -H 'Content-Type: application/json' \
  -d '{"scene":"product","input":"演示脚本商品上架任务","payload":{"platform":"alibaba"}}')
printf '%s' "$CREATE" | pretty_print
TASK_ID=$(printf '%s' "$CREATE" | sed -n 's/.*"task_id":"\([^"]*\)".*/\1/p')
echo

echo "[3] preview"
curl -s "$BASE_URL/tasks/$TASK_ID/preview" | pretty_print
echo

echo "[4] logs"
curl -s "$BASE_URL/tasks/$TASK_ID/logs" | pretty_print
echo

echo "[5] status"
curl -s "$BASE_URL/tasks/$TASK_ID/status" | pretty_print
echo

echo "[6] confirm"
curl -s -X POST "$BASE_URL/tasks/$TASK_ID/confirm" \
  -H 'Content-Type: application/json' \
  -d '{"approved":true,"comment":"demo script confirm","approver":"Walter"}' | pretty_print
echo

echo "[7] list tasks"
curl -s "$BASE_URL/tasks?scene=product&limit=5" | pretty_print
