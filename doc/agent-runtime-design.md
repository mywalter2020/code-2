# Agent Runtime Design

## 1. 文档目的

本文档用于定义一个新的 Agent 项目方案。该项目不是通用工作流引擎，也不是单一对话机器人，而是一个面向软件任务交付的、多阶段、多 Agent、人机协同运行时。

该系统的核心目标是把一条典型的软件需求交付链路产品化：

- 用户输入需求
- 分析 Agent 理解需求并产出 PRD
- 用户确认 PRD
- Todo Agent 基于 PRD 生成 Todo List
- 用户确认 Todo List
- 执行 Agent 按 Todo List 执行具体任务
- 统一返回执行结果

同时，每个 Agent 都不是一个简单函数，而是一个具备人格、身份、职责和行为规则的可定义执行体。

---

## 2. 项目定义

**Agent 是一个用 Go 实现的、面向软件任务交付流程的轻量级多 Agent 人机协同运行时。**

它具备以下特点：

- HTTP API 驱动
- 多阶段执行流程
- Human-in-the-loop
- Skill / Function 驱动执行
- 支持 Todo 多线并行推进
- 执行阶段采用 ReAct + Plan-and-Execute 模式
- 每个 Agent 通过 Profile 文件定义自身行为边界与风格
- 借鉴 OpenClaw 的思路，但项目本身独立实现

---

## 3. 背景与设计动机

典型的软件任务执行，并不是用户一句话输入后就直接开始写代码，而是会经历一条更合理的中间过程：

1. 先理解需求
2. 再形成 PRD 或需求结构化表达
3. 再拆成可执行 Todo
4. 再进入实际执行
5. 执行过程中根据上下文不断观察、决策、行动
6. 最终回收产物并统一输出

如果缺少中间层，常见问题包括：

- 需求理解偏差
- 直接执行导致返工
- 任务拆解不完整
- 多任务执行缺乏统一上下文
- 无法在用户确认节点进行控制
- 不同 Agent 的行为风格不稳定

因此，该系统的目标不是“尽量自动化一切”，而是建立一条：

**可理解、可确认、可执行、可并行、可追踪** 的 Agent 交付链路。

---

## 4. 核心设计原则

### 4.1 先结构化，再执行
系统先做需求分析与任务拆解，再进入执行阶段，避免未经确认直接开始实现。

### 4.2 Human-in-the-loop 是一等公民
PRD 和 Todo List 都需要用户确认。用户不是旁观者，而是流程中的关键决策节点。

### 4.3 执行阶段以 Todo 为中心
执行 Agent 的职责不是“自由发挥”，而是围绕 Todo List 中的任务进行推进。Todo 是执行边界和节奏控制中心。

### 4.4 Agent 是具备 Profile 的执行体
每个 Agent 都由独立的定义文件描述，不只是一个代码模块，而是一个带人格、身份、职责和规则的可塑执行体。

### 4.5 统一输入输出协议
无论分析、Todo 还是执行阶段，都使用统一的响应结构，便于 API、日志、前端、审计和后续扩展。

### 4.6 先固定主链路，后考虑泛化
V1 不追求通用 Workflow DSL。先把最核心的主流程打磨清楚，再考虑抽象为更通用的编排框架。

### 4.7 并行执行服务于 Todo，而不是为了炫技
并发能力主要用于处理可并行的 Todo 项，不盲目引入复杂的 DAG 抽象。

---

## 5. 系统目标与非目标

### 5.1 V1 目标
V1 需要实现：

- 通过 HTTP API 接收用户任务
- 分析 Agent 生成 PRD
- 用户确认 PRD
- Todo Agent 生成 Todo List
- 用户确认 Todo List
- 执行 Agent 根据 Todo List 执行
- Todo List 支持多线并行推进
- 执行阶段采用 ReAct + Plan-and-Execute 模式
- 统一输出结果结构
- 全流程落 PostgreSQL
- 每个 Agent 通过 Profile 文件定义

### 5.2 V1 非目标
V1 不做：

- 通用 Workflow DSL
- 可视化流程编辑器
- 前端 UI
- 分布式任务队列
- 多租户权限系统
- 动态插件市场
- 深度耦合 OpenClaw Runtime

---

## 6. 整体流程

### 6.1 主流程

```text
User Input
  -> Analyst Agent
  -> PRD Draft
  -> User Confirm PRD
  -> Todo Agent
  -> Todo List
  -> User Confirm Todo
  -> Executor Agent
  -> Final Unified Response
```

### 6.2 分阶段说明

#### Stage 1: Input Received
接收用户原始需求输入，并创建 Session / Task。

#### Stage 2: Analysis
分析 Agent 对需求进行理解、归纳与结构化，生成 PRD 草案。

#### Stage 3: PRD Confirm
用户对 PRD 进行确认、修改或驳回。只有确认后，才能进入 Todo 阶段。

#### Stage 4: Todo Generate
Todo Agent 基于确认后的 PRD 生成可执行 Todo List。

#### Stage 5: Todo Confirm
用户确认 Todo List 的拆解结果。可以增删改优先级、依赖与范围。

#### Stage 6: Execute
执行 Agent 基于 Todo List 推进具体工作。执行过程中允许串行与并行混合推进。

#### Stage 7: Final Response
系统统一返回执行状态、产物、摘要、日志和下一步建议。

---

## 7. Agent 角色模型

V1 至少包含三类核心 Agent。

### 7.1 Analyst Agent
职责：

- 理解用户输入
- 提炼目标、范围、边界和约束
- 形成 PRD 草案
- 标记不确定项、风险点、待确认项

输出核心产物：

- `prd.md`
- 分析摘要
- 风险与待确认列表

### 7.2 Todo Agent
职责：

- 基于确认后的 PRD 生成可执行 Todo List
- 给出优先级、建议顺序和依赖关系
- 标记可并行任务
- 标记验收标准或完成条件

输出核心产物：

- `todo.json` / `todo.md`
- 执行建议
- 并行分组建议

### 7.3 Executor Agent
职责：

- 读取 Todo List
- 按 Todo 项推进执行
- 针对不同 Todo 项调用不同技能或功能模块
- 在执行过程中采用 ReAct + Plan-and-Execute 模式
- 汇总中间结果与最终结果

执行 Agent 本身不等同于某一个固定功能，而是一个执行控制器。它根据 Todo 内容决定：

- 当前该做什么
- 应调用哪个功能模块
- 哪些 Todo 可以并行推进
- 遇到失败时如何处理

输出核心产物：

- 代码或文件变更
- 执行日志
- 每个 Todo 的状态
- 最终交付摘要

---

## 8. Executor 的执行模型

这是系统的核心差异点之一。

### 8.1 模型定义
执行 Agent 采用 **ReAct + Plan-and-Execute** 混合模式。

即：

- **Plan-and-Execute**：先基于 Todo List 做执行计划，决定任务顺序、并行策略、预估路径
- **ReAct**：执行某个 Todo 项时，持续进行观察（Observe）→ 推理（Reason）→ 行动（Act）→ 再观察 的循环

### 8.2 为什么这样设计
如果只有 Plan-and-Execute，执行会过于死板；
如果只有 ReAct，执行容易缺少全局节奏控制。

二者结合后：

- 上层由 Todo 驱动计划
- 下层由 ReAct 驱动具体执行

这样既有整体节奏，也保留局部灵活性。

### 8.3 执行流程示意

```text
Load confirmed Todo List
  -> Build execution plan
  -> Select next executable todos
  -> For each todo:
       Observe context
       Reason about next action
       Call skill/function
       Capture result
       Update todo state
  -> Merge progress
  -> Continue until all todos finished or blocked
```

### 8.4 多线并行
Todo List 中允许多个任务并行推进。

适用场景例如：

- 文档与代码可并行
- 前端与后端独立子任务可并行
- 测试准备与代码生成可部分并行

并行原则：

- 无直接依赖的 Todo 才允许并行
- 并行任务的上下文要隔离
- 每条执行线都需要独立日志和状态
- 汇总层负责合并多线结果

V1 中不要求引入完整 DAG 引擎，但需要支持：

- Todo 分组
- 并发执行控制
- 结果回收合并

---

## 9. Agent Profile 定义

每个 Agent 通过固定文件结构定义自身。

### 9.1 文件组成

- **人格** → `SOUL.md`
- **身份标签** → `IDENTITY.md`
- **岗位职责 / 当前定位** → `MEMORY.md`
- **做事规则** → `AGENTS.md`

### 9.2 语义说明

#### SOUL.md
定义 Agent 的人格、风格、表达偏好、行为气质。

示例内容：

- 偏直接还是偏温和
- 更保守还是更主动
- 是否强调解释
- 是否偏工程化表达

#### IDENTITY.md
定义 Agent 的身份标签。

示例内容：

- 名称
- 类型
- 所属角色
- Emoji / 标签
- 定位关键词

#### MEMORY.md
定义该 Agent 的长期定位、职责边界、历史经验、当前注意事项。

示例内容：

- 这个 Agent 擅长什么
- 不该做什么
- 当前项目里的职责
- 关键经验与约束

#### AGENTS.md
定义做事规则和操作规范。

示例内容：

- 遇到不确定先问还是先做
- 哪些动作必须谨慎
- 输出格式要求
- 工具调用边界

### 9.3 建议目录结构

```text
agents/
  analyst/
    SOUL.md
    IDENTITY.md
    MEMORY.md
    AGENTS.md
  todo/
    SOUL.md
    IDENTITY.md
    MEMORY.md
    AGENTS.md
  executor/
    SOUL.md
    IDENTITY.md
    MEMORY.md
    AGENTS.md
```

### 9.4 Profile 的作用
Agent Profile 并不是装饰性文档，而是运行时配置的一部分。系统启动或调用 Agent 时，应读取这些 Profile 内容，并将其拼装为该 Agent 的行为上下文。

---

## 10. Skill / Function 模型

虽然系统强调 Agent，但真正落地执行时仍需要更细粒度的功能单元。

因此建议：

- **Agent** 负责角色级决策
- **Skill / Function** 负责能力级执行

例如执行阶段可能存在以下功能模块：

- `read_file`
- `write_file`
- `edit_file`
- `run_test`
- `search_code`
- `generate_code`
- `summarize_changes`

执行 Agent 会根据 Todo 项内容选择合适的 Skill / Function。

也就是说：

**Todo 决定做什么，Executor 决定怎么做，Skill 决定具体执行。**

---

## 11. 统一数据模型

### 11.1 Session
一次用户完整任务会话。

建议字段：

- `session_id`
- `user_input`
- `current_stage`
- `status`
- `created_at`
- `updated_at`

### 11.2 PRD Artifact
PRD 阶段产物。

建议字段：

- `title`
- `background`
- `goals`
- `scope`
- `non_goals`
- `requirements`
- `risks`
- `open_questions`
- `markdown`

### 11.3 Todo Item
Todo List 中的单个任务项。

建议字段：

- `id`
- `title`
- `description`
- `status`
- `priority`
- `depends_on`
- `parallel_group`
- `acceptance_criteria`
- `assigned_agent`
- `result`

### 11.4 Execution Record
执行阶段记录。

建议字段：

- `id`
- `session_id`
- `todo_id`
- `executor`
- `status`
- `input`
- `output`
- `reasoning_summary`
- `started_at`
- `finished_at`

---

## 12. 统一响应格式

所有阶段都应尽量返回统一结构。

建议统一响应协议如下：

```json
{
  "session_id": "sess_xxx",
  "stage": "waiting_prd_confirm",
  "agent": "analyst",
  "status": "waiting_user_confirm",
  "artifacts": {
    "prd": {}
  },
  "message": "PRD 已生成，请确认后进入 Todo 阶段",
  "next_actions": [
    "confirm_prd",
    "edit_prd",
    "cancel"
  ],
  "meta": {}
}
```

执行完成时，例如：

```json
{
  "session_id": "sess_xxx",
  "stage": "done",
  "agent": "executor",
  "status": "success",
  "artifacts": {
    "todo": [],
    "deliverables": [],
    "summary": "执行完成"
  },
  "message": "任务已完成",
  "next_actions": [
    "view_result",
    "start_new_task"
  ],
  "meta": {}
}
```

### 12.1 统一返回的价值

- 前端处理简单
- API 风格统一
- 调试和日志更容易
- 各阶段状态切换清晰
- 后续扩 stage 更顺手

---

## 13. 阶段状态设计

建议使用以下状态集合：

- `input_received`
- `analysis_done`
- `waiting_prd_confirm`
- `todo_done`
- `waiting_todo_confirm`
- `executing`
- `done`
- `failed`
- `canceled`

同时，每个 Todo 项自己也有状态，例如：

- `pending`
- `ready`
- `running`
- `blocked`
- `done`
- `failed`
- `canceled`

---

## 14. HTTP API 设计建议

V1 先采用最小接口集。

### 14.1 创建会话
`POST /sessions`

用于提交用户输入并启动分析阶段。

### 14.2 查询会话
`GET /sessions/{session_id}`

用于获取当前阶段、状态、产物与摘要。

### 14.3 确认 PRD
`POST /sessions/{session_id}/confirm-prd`

用户确认后进入 Todo 阶段。

### 14.4 更新 PRD
`POST /sessions/{session_id}/edit-prd`

用于用户补充/修订 PRD。

### 14.5 确认 Todo
`POST /sessions/{session_id}/confirm-todo`

用户确认后进入执行阶段。

### 14.6 更新 Todo
`POST /sessions/{session_id}/edit-todo`

用于用户修改 Todo List。

### 14.7 查询执行详情
`GET /sessions/{session_id}/executions`

获取每个 Todo 的执行状态与记录。

### 14.8 查询 Agent 列表
`GET /agents`

返回当前系统已注册的 Agent 及其 Profile 摘要。

### 14.9 健康检查
`GET /healthz`

---

## 15. PostgreSQL 持久化设计

V1 既然确定使用 PostgreSQL，那么应把它作为事实记录层。

### 15.1 建议表

#### sessions
存一次完整任务会话。

#### session_artifacts
存 PRD、Todo、最终结果等阶段产物。

#### todo_items
存 Todo 列表及其状态。

#### execution_records
存执行阶段每个 Todo 的执行记录。

#### logs
存系统日志、Agent 日志、步骤日志。

#### agent_profiles
可选。若要缓存 Agent Profile 摘要，可落库；否则可直接从文件读取。

### 15.2 落库原则

- 输入、阶段状态、产物、步骤结果都可追踪
- 可重放关键决策节点
- 并行任务的执行记录彼此可区分
- 失败原因必须可查询

---

## 16. 建议目录结构

```text
agent/
  cmd/
    server/
  internal/
    api/
    runtime/
    pipeline/
    profiles/
    agents/
    skills/
    store/
    types/
    config/
  agents/
    analyst/
      SOUL.md
      IDENTITY.md
      MEMORY.md
      AGENTS.md
    todo/
      SOUL.md
      IDENTITY.md
      MEMORY.md
      AGENTS.md
    executor/
      SOUL.md
      IDENTITY.md
      MEMORY.md
      AGENTS.md
  migrations/
  docs/
  go.mod
  README.md
```

说明：

- `internal/api`：HTTP 接口层
- `internal/runtime`：执行支撑与状态流转
- `internal/pipeline`：主流程控制
- `internal/profiles`：Agent Profile 加载与解析
- `internal/agents`：Analyst / Todo / Executor 实现
- `internal/skills`：执行期功能模块
- `internal/store`：PostgreSQL 持久化
- `internal/types`：共享类型定义

---

## 17. V1 开发建议顺序

### Phase 1: 基础框架
- 初始化项目
- 搭建 HTTP 服务
- 接入 PostgreSQL
- 完成 `healthz`

### Phase 2: Session 与 Analysis
- 创建 Session
- Analyst Agent 生成 PRD
- 查询 Session 状态
- 确认 PRD

### Phase 3: Todo 阶段
- Todo Agent 生成 Todo List
- 查询 Todo
- 编辑 / 确认 Todo

### Phase 4: Execute 阶段
- Executor Agent 读取 Todo List
- 实现串行与并行混合执行
- 引入 ReAct + Plan-and-Execute 执行循环
- 记录执行明细

### Phase 5: 统一输出与收尾
- 统一响应结构
- 汇总最终产物
- 完善日志与错误处理
- 补充文档与示例

---

## 18. 当前已确认决策

以下内容已在当前讨论中明确：

- 项目定位：轻量级多智能体编排引擎
- 用户：先自己用，后续可开放
- 入口：HTTP API
- Agent 模型：Go 接口
- 持久化：PostgreSQL
- 与 OpenClaw 关系：借鉴思路，但独立实现
- 主流程：分析 -> PRD 确认 -> Todo -> Todo 确认 -> 执行 -> 统一返回
- 执行 Agent：根据 Todo List 处理不同功能
- 执行方法：ReAct + Plan-and-Execute
- Todo：支持多线同时推进
- Agent 定义方式：`SOUL.md` / `IDENTITY.md` / `MEMORY.md` / `AGENTS.md`

---

## 19. V1 标准案例：URL 产品学习到预览页生成

为了让系统设计不漂移，V1 需要围绕一个标准案例推进实现。该案例既能覆盖分析、PRD、Todo、执行、并行、确认和统一返回，也能直接作为前后端联调样板。

### 19.1 用户输入

用户示例输入：

> 我想学习下某个 URL 里的这个产品，你帮我处理下。

用户真实意图不是简单抓取网页，而是希望系统：

1. 理解 URL 对应的产品
2. 形成需求说明 / 产品说明
3. 生成执行 Todo List
4. 产出一组商品素材
5. 把结果交给前端生成一个模拟预览页面

### 19.2 该案例的目标

该案例的最终目标不是“返回一段总结”，而是生成一套完整的可预览商品内容资产，包括：

- 产品标题
- 产品图片相关文案
- 图片描述图内容
- 产品轮播图内容
- 产品视频脚本或视频描述
- 供前端直接消费的 Preview Payload

### 19.3 案例主流程

```text
用户输入 URL
  -> 分析 Agent 读取并理解产品
  -> 生成需求说明 / 产品说明（PRD）
  -> 用户确认 PRD
  -> Todo Agent 生成 Todo List
  -> 用户确认 Todo List
  -> Executor Agent 按 Todo 执行
  -> 汇总产物
  -> 返回前端 Preview Payload
  -> 前端生成模拟预览页面
```

### 19.4 阶段拆解

#### Stage A: 产品理解

输入：

- 产品 URL
- 用户原始需求
- 可选参数（平台、语言、风格、目标受众）

分析 Agent 要完成：

- 抓取和解析 URL 内容
- 理解该产品是什么
- 提取核心卖点
- 提取关键结构信息
- 形成可供后续执行的产品说明

该阶段输出：

- 产品基础信息
- 产品摘要
- 卖点列表
- 使用场景
- 待确认项
- PRD 草案

#### Stage B: PRD 确认

用户确认：

- 对产品的理解是否正确
- 需要保留哪些卖点
- 素材范围是否完整
- 风格是否有特殊要求

该阶段是关键确认点，未确认前不进入大规模生成阶段。

#### Stage C: Todo 生成

Todo Agent 基于 PRD 生成 Todo List，形成执行计划。

Todo 的主要目标包括：

- 生成产品标题
- 生成卖点摘要
- 生成产品图片相关文案
- 生成图片描述图内容
- 生成产品轮播图内容
- 生成产品视频脚本
- 汇总预览结构
- 生成前端所需的 Preview Payload

#### Stage D: Todo 确认

用户确认 Todo List 是否合理，包括：

- 是否需要视频内容
- 是否暂时只做文案，不做真实图片
- 哪些内容必须先做
- 是否需要限制输出语言或平台风格

#### Stage E: 执行

Executor Agent 读取确认后的 Todo List，进入执行阶段。

执行原则：

- 前置理解任务必须先完成
- 可独立的素材生成任务允许并行
- 依赖所有素材结果的汇总任务必须后置
- 执行方式采用 ReAct + Plan-and-Execute

#### Stage F: 统一返回

当所有 Todo 完成后，系统统一返回：

- 产品理解结果
- PRD
- Todo List
- 各类素材产物
- Preview Payload
- 状态与摘要信息

### 19.5 并行关系说明

这个案例里，并行能力非常重要。

#### 必须串行的部分

以下步骤必须按顺序进行：

1. URL 解析与产品理解
2. PRD 生成
3. PRD 确认
4. Todo 生成
5. Todo 确认

这些属于上游结构化阶段，不能跳过。

#### 可以并行的部分

在 Todo 确认后，多个内容资产可以并行生成，例如：

- 产品标题
- 产品图片文案
- 图片描述图内容
- 产品轮播图内容
- 产品视频脚本

这些任务可以被视为同一个并行组，待全部完成后再进行汇总。

#### 必须后置的部分

以下步骤通常需要等待前面的并行任务完成：

- 汇总生成物料结构
- 构建 Preview Payload
- 返回统一结果

### 19.6 PRD 示例结构

该案例中的 PRD 建议至少包含以下内容：

#### 输入信息

- 原始 URL
- 用户原始需求

#### 产品理解

- 产品名称
- 产品类型
- 核心功能
- 核心卖点
- 目标人群
- 适用场景
- 差异化特点

#### 交付目标

- 生成产品标题
- 生成图片描述素材
- 生成轮播图内容
- 生成视频脚本
- 生成前端预览数据

#### 风格要求

- 文案风格
- 输出语言
- 视觉风格说明

#### 待确认项

- 是否需要视频
- 是否生成真实图片
- 是否只需要模拟预览
- 是否需要多平台版本

### 19.7 Todo List 示例结构

Todo 不应只是字符串数组，而应是结构化任务列表。示例：

```json
[
  {
    "id": "todo_1",
    "title": "生成产品标题",
    "type": "title_generation",
    "status": "pending",
    "depends_on": [],
    "parallel_group": "content_assets",
    "acceptance_criteria": [
      "标题准确体现产品核心卖点",
      "标题适合目标平台展示"
    ]
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
```

### 19.8 执行期 Skill / Function 映射建议

该案例下建议的功能模块包括：

#### 输入理解类

- `fetch_product_url`
- `parse_product_page`
- `extract_product_info`

#### 文案生成类

- `generate_product_title`
- `generate_product_selling_points`
- `generate_image_caption`
- `generate_carousel_copy`
- `generate_video_script`

#### 结构整理类

- `merge_generation_results`
- `build_preview_payload`

V1 可以先只产出文案与结构化描述，不要求真实生成图片或视频文件。真实视觉素材可以在后续版本扩展。

### 19.9 最终返回给前端的数据结构

该案例的最终返回不应只是文本，而应是前端可直接消费的结构化 payload。示例：

```json
{
  "session_id": "sess_001",
  "stage": "done",
  "status": "success",
  "artifacts": {
    "product": {
      "source_url": "https://example.com/product/123",
      "name": "某产品",
      "summary": "这是一款..."
    },
    "prd": {
      "markdown": "..."
    },
    "todo_list": [],
    "materials": {
      "title": "xxx产品标题",
      "image_descriptions": [
        "图1描述...",
        "图2描述..."
      ],
      "carousel_items": [
        {
          "title": "卖点1",
          "description": "..."
        },
        {
          "title": "卖点2",
          "description": "..."
        }
      ],
      "video_script": {
        "hook": "...",
        "scenes": []
      }
    },
    "preview": {
      "hero_title": "xxx产品标题",
      "hero_subtitle": "产品核心卖点",
      "feature_blocks": [],
      "carousel": [],
      "video_section": {}
    }
  }
}
```

前端应直接消费 `preview` 字段来生成模拟预览页面。

### 19.10 该案例对系统的验证价值

该案例几乎覆盖了 V1 的全部关键能力：

- URL 输入型任务
- Analyst Agent 生成 PRD
- 用户确认 PRD
- Todo Agent 生成 Todo List
- 用户确认 Todo List
- Executor Agent 按 Todo 执行
- Todo 多线并行
- 统一结果汇总
- 返回前端 Preview Payload

因此，该案例应被视为 V1 的标准验收样板。后续系统开发、接口设计、表结构设计和联调，都应优先围绕它进行。

---

## 20. 后续待确认事项

进入正式开发前，还建议继续确认这些问题：

1. PRD 和 Todo 的最终主格式是否统一为 Markdown + JSON 双存
2. Executor 调用 Skill 时，是否要有统一 Tool 接口协议
3. 并行 Todo 的最大并发数是否需要配置
4. 是否需要 Review Agent 作为执行后的补充阶段
5. Session 是否支持中断恢复
6. 是否需要为不同 Agent 提供公共 Profile 模板

---

## 21. 结论

该项目的本质不是“做一个通用工作流引擎”，而是：

**做一个围绕软件需求交付过程构建的、多阶段、多 Agent、人机协同运行时。**

它的最小闭环非常明确：

- 先理解需求
- 再形成 PRD
- 再拆成 Todo
- 再进入执行
- 执行时采用 ReAct + Plan-and-Execute
- 支持 Todo 多线并行
- 全流程统一协议与持久化记录

如果这个闭环做顺，后续再扩展 review、test、deploy、memory、skill marketplace 等能力，都会有很稳定的基础。
