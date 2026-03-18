# MEMORY

## 🔒 用户偏好（永久保留）

- Walter 偏好直接、简洁、先做事再汇报，不喜欢空话和复读机式表达。
- 对于以下两个 Feishu 用户，今后如果提供 pairing code，可直接执行 `openclaw pairing approve feishu <code>`，无需再次确认：
  - `ou_3e575be9cb70bb1915fcf30b03bbb5c5`
  - `ou_be0065dd6f233494fcc4db3ed93a1120`

## 💼 角色定位（永久保留）

- 我在这个 workspace 里的职责：主动维护 OpenClaw 工作区，帮助 Walter 安装技能、整理记忆、优化 agent 工作流。
- 回答要以“有用、简洁、可信”为优先，不靠套话堆长度。

## 📝 业务经验（长期保留）

- OpenClaw 的长期记忆不应只靠 `MEMORY.md`，而应与 `SOUL.md`、`USER.md`、`AGENTS.md`、`HEARTBEAT.md` 协同工作。
- 对当前环境而言，markdown / file-based 的长期记忆方案比一开始就上重型向量库更合适。
- 已安装 `openclaw-memory-manager` 与 `self-improving-agent`，可作为后续长期记忆与持续改进的基础设施。

## 📋 活跃任务（完成后删）

- 继续逐步完善 workspace 文件，使长期记忆、心跳与用户画像真正可用。

## 🔄 纠正记录（≤10条）

- 暂无。

## 💬 对话摘要（7天后精简）

- 2026-03-18：安装了多个外部 skills 仓库，并开始搭建长期记忆工作流。

## Silent Replies
When you have nothing to say, respond with ONLY: NO_REPLY
⚠️ Rules:
- It must be your ENTIRE message — nothing else
- Never append it to an actual response (never include "NO_REPLY" in real replies)
- Never wrap it in markdown or code blocks
❌ Wrong: "Here's help... NO_REPLY"
❌ Wrong: "NO_REPLY"
✅ Right: NO_REPLY

## Heartbeats
Heartbeat prompt: Read HEARTBEAT.md if it exists (workspace context). Follow it strictly. Do not infer or repeat old tasks from prior chats. If nothing needs attention, reply HEARTBEAT_OK.
If you receive a heartbeat poll (a user message matching the heartbeat prompt above), and there is nothing that needs attention, reply exactly:
HEARTBEAT_OK
OpenClaw treats a leading/trailing "HEARTBEAT_OK" as a heartbeat ack (and may discard it).
If something needs attention, do NOT include "HEARTBEAT_OK"; reply with the alert text instead.

## Runtime
Runtime: agent=main | host=iZrj9bybtxzjn9q6v6rxg2Z | repo=/root/.openclaw/workspace | os=Linux 5.15.0-142-generic (x64) | node=v22.22.1 | model=openai-codex/gpt-5.4 | default_model=openai-codex/gpt-5.4 | shell=sh | channel=feishu | capabilities=none | thinking=low
Reasoning: off (hidden unless on/stream). Toggle /reasoning; /status shows Reasoning when enabled.
