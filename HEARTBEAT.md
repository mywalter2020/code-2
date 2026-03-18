# HEARTBEAT.md

## Default behavior

如果没有明确需要检查的事项，直接回复：`HEARTBEAT_OK`

## Quiet hours

- 23:00-08:00 除非紧急，否则不要主动打扰。

## Periodic checks

按需轮询，不要每次 heartbeat 都全量执行：

1. 检查最近是否有新的重要消息或待继续的任务。
2. 检查 workspace 是否有明显未整理状态：
   - 大量未提交改动
   - 新增 memory 值得提炼到 `MEMORY.md`
   - 新 learnings 值得晋升到 `AGENTS.md` / `TOOLS.md`
3. 如果 8 小时以上没有主动输出，且有明确进展或提醒价值，可以主动发一条简短更新。

## Reach out only when

- 有新的重要进展
- 有明确阻塞需要 Walter 决策
- 发现高价值但低风险的可整理项
- 未来 2 小时内有需要提醒的事项

## Avoid

- 重复提醒同一件事
- 没有新信息时硬发消息
- 为了完成 heartbeat 而制造噪音
