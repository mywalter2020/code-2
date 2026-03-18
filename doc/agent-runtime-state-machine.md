# Agent Runtime State Machine

## 1. 文档目标

本文档定义 Agent Runtime 的状态机设计，用于明确：

- Session 在各阶段之间如何流转
- PRD / Todo 的确认节点如何生效
- Execution 的运行、失败、重试、取消如何处理
- 哪些状态是终态，哪些状态可以继续推进

该文档是设计文档、API 草案、PG Schema 和后续实现之间的状态契约。

---

## 2. 设计原则

### 2.1 Session 是主状态机
整个系统的核心状态机对象是 `session`。PRD、Todo 和 Execution 都围绕 Session 推进。

### 2.2 确认节点是显式状态
PRD 和 Todo 确认不能隐式跳过，必须由状态明确表达：

- `waiting_prd_confirm`
- `waiting_todo_confirm`

### 2.3 执行态与执行结果分离
`executing` 表示 session 正在执行；Todo 和 Execution 自身有子状态，用于表达细粒度进度。

### 2.4 终态必须稳定
一旦 session 进入终态：

- `done`
- `failed`
- `canceled`

系统应避免再进入普通处理中间态，除非显式定义“重开/复制/新建 session”语义。

### 2.5 重试是执行层行为，不是重置整个世界
默认情况下，retry 优先作用在执行项，而不是直接重置整个 session 流程。

---

## 3. 状态机层级

系统至少有三层状态：

1. **Session 状态**：全局主流程状态
2. **Todo 状态**：任务拆解后的计划项状态
3. **Execution 状态**：具体运行实例状态

这三层有从属关系：

```text
Session
  └── Todo Items
        └── Execution Records
```

---

## 4. Session 状态定义

### 4.1 状态集合

建议 Session 状态如下：

- `input_received`
- `analysis_done`
- `waiting_prd_confirm`
- `todo_done`
- `waiting_todo_confirm`
- `executing`
- `done`
- `failed`
- `canceled`

### 4.2 状态语义

#### input_received
刚收到用户输入，session 已创建，但分析结果尚未持久化完成。

#### analysis_done
分析阶段已完成，PRD 已生成。该状态通常是瞬时状态，实际对外更常表现为 `waiting_prd_confirm`。

#### waiting_prd_confirm
PRD 已生成，等待用户确认或修改。

#### todo_done
Todo 已生成。该状态通常也是瞬时状态，实际对外更常表现为 `waiting_todo_confirm`。

#### waiting_todo_confirm
Todo List 已生成，等待用户确认或修改。

#### executing
Executor 已开始按 Todo 推进执行。

#### done
整个 session 已完成，最终产物已生成。

#### failed
整个 session 已失败，且当前没有继续自动推进的路径。

#### canceled
session 被用户或系统主动取消。

---

## 5. Session 主流程状态图

```text
[input_received]
    |
    v
[analysis_done]
    |
    v
[waiting_prd_confirm] -- edit-prd --> [waiting_prd_confirm]
    |
    | confirm-prd
    v
[todo_done]
    |
    v
[waiting_todo_confirm] -- edit-todo --> [waiting_todo_confirm]
    |
    | confirm-todo
    v
[executing] -- all success --> [done]
    |   \
    |    \-- unrecoverable failure --> [failed]
    |
    \-- cancel --> [canceled]

[waiting_prd_confirm] -- cancel --> [canceled]
[waiting_todo_confirm] -- cancel --> [canceled]
```

### 5.1 说明

- `analysis_done` 和 `todo_done` 可以是内部瞬时状态，不一定需要长期暴露给前端。
- 对前端来说，最重要的是：
  - `waiting_prd_confirm`
  - `waiting_todo_confirm`
  - `executing`
  - `done`
  - `failed`
  - `canceled`

---

## 6. Session 状态流转规则

### 6.1 创建 Session

触发：`POST /sessions`

流转：

```text
none -> input_received -> analysis_done -> waiting_prd_confirm
```

约束：

- 创建成功后必须有 PRD 或至少有 PRD 生成任务
- 若分析阶段失败，可直接进入 `failed`

### 6.2 PRD 编辑

触发：`POST /sessions/{id}/edit-prd`

流转：

```text
waiting_prd_confirm -> waiting_prd_confirm
```

约束：

- 只能在 `waiting_prd_confirm` 状态下执行
- 编辑会提升 PRD 版本号
- 不推进阶段，只刷新 PRD

### 6.3 PRD 确认

触发：`POST /sessions/{id}/confirm-prd`

流转：

```text
waiting_prd_confirm -> todo_done -> waiting_todo_confirm
```

约束：

- 只能在 `waiting_prd_confirm` 状态下执行
- 确认后系统应触发 Todo 生成
- 如果 Todo 生成失败，可进入 `failed`

### 6.4 Todo 编辑

触发：`POST /sessions/{id}/edit-todo`

流转：

```text
waiting_todo_confirm -> waiting_todo_confirm
```

约束：

- 只能在 `waiting_todo_confirm` 状态下执行
- 编辑会提升 Todo 版本号
- 不推进阶段，只刷新 Todo

### 6.5 Todo 确认

触发：`POST /sessions/{id}/confirm-todo`

流转：

```text
waiting_todo_confirm -> executing
```

约束：

- 只能在 `waiting_todo_confirm` 状态下执行
- 确认后立即启动 Executor
- 执行开始后不允许再编辑当前 Todo 版本

### 6.6 执行完成

触发：全部 Todo 执行完成且汇总成功

流转：

```text
executing -> done
```

约束：

- 所有必须执行的 Todo 都为成功完成态
- 最终 Preview / Final Result 已生成

### 6.7 执行失败

触发：关键执行失败且无法自动恢复

流转：

```text
executing -> failed
```

约束：

- 至少有一个关键 Todo 或汇总步骤失败
- 当前策略下已无继续自动推进路径

### 6.8 取消 Session

触发：`POST /sessions/{id}/cancel`

可取消的来源状态：

- `waiting_prd_confirm`
- `waiting_todo_confirm`
- `executing`

流转：

```text
waiting_prd_confirm -> canceled
waiting_todo_confirm -> canceled
executing -> canceled
```

约束：

- 取消后应停止后续自动推进
- 若执行中已有在跑任务，应尝试标记为取消或停止调度新任务

---

## 7. Todo 状态定义

### 7.1 状态集合

- `pending`
- `ready`
- `running`
- `blocked`
- `done`
- `failed`
- `canceled`

### 7.2 状态语义

#### pending
已存在于 Todo List 中，但还未满足执行条件。

#### ready
已满足执行条件，可进入执行队列。

#### running
当前正在执行。

#### blocked
由于依赖失败、缺少输入或人工介入导致当前无法继续。

#### done
该 Todo 已成功完成。

#### failed
该 Todo 当前执行失败。

#### canceled
该 Todo 被取消。

---

## 8. Todo 状态流转规则

### 8.1 初始化

当 Todo List 刚生成并被确认后：

- 无依赖的 Todo：`pending -> ready`
- 有依赖的 Todo：保持 `pending`

### 8.2 开始执行

```text
ready -> running
```

### 8.3 执行成功

```text
running -> done
```

### 8.4 执行失败

```text
running -> failed
```

### 8.5 依赖阻塞

```text
pending -> blocked
```

适用场景：

- 依赖项失败
- 缺少必要输入
- 需要等待人工处理

### 8.6 取消

```text
pending -> canceled
ready -> canceled
running -> canceled
blocked -> canceled
```

### 8.7 重试后恢复

若失败 Todo 允许重试：

```text
failed -> ready
```

若是 blocked Todo 且条件恢复：

```text
blocked -> ready
```

---

## 9. Execution 状态定义

### 9.1 状态集合

- `queued`
- `running`
- `done`
- `failed`
- `canceled`

### 9.2 状态语义

#### queued
执行记录已创建，等待资源调度。

#### running
正在运行。

#### done
本次执行成功完成。

#### failed
本次执行失败。

#### canceled
本次执行被取消。

---

## 10. Execution 状态流转规则

### 10.1 创建执行记录

```text
none -> queued
```

### 10.2 开始执行

```text
queued -> running
```

### 10.3 执行成功

```text
running -> done
```

### 10.4 执行失败

```text
running -> failed
```

### 10.5 执行取消

```text
queued -> canceled
running -> canceled
```

### 10.6 重试

Execution 的 retry 不建议把原记录状态改回 queued，而是：

- 保留原 execution record
- 新建一条新的 execution record
- `attempt = previous_attempt + 1`

也就是说：

```text
failed execution (historical record kept)
    -> create new queued execution
    -> running
    -> done/failed
```

这样更利于审计与调试。

---

## 11. 并行执行下的状态规则

系统支持 Todo 多线并行，因此需要明确 Session 与 Todo 的关系。

### 11.1 Session 在 executing 时不等待单个 Todo
只要存在至少一个 Todo 在运行或待调度，Session 仍保持 `executing`。

### 11.2 Session 进入 done 的条件
Session 只有在以下条件同时满足时才进入 `done`：

1. 所有必需 Todo 都完成
2. 汇总类 Todo 完成
3. Preview Payload 已生成
4. 没有正在运行中的 Execution

### 11.3 Session 进入 failed 的条件
以下情况可以使 Session 进入 `failed`：

1. 某关键 Todo 失败且不可恢复
2. 汇总类 Todo 失败
3. Preview 构建失败且无 fallback
4. 当前策略判定无法继续推进

### 11.4 局部失败不一定导致 Session 立刻 failed
如果失败的是可选 Todo，则可以有两种策略：

- 继续执行，最终带 warning 完成
- 进入 failed

V1 建议先简单化：

- 关键 Todo 失败 -> session failed
- 非关键 Todo 是否允许 optional，可后续扩展

---

## 12. Retry 规则

### 12.1 Session 层面 retry
V1 建议优先只支持 Execution / Todo 层面的 retry，不直接支持完整 Session 回滚重跑。

### 12.2 Execution retry
触发：`POST /sessions/{id}/executions/retry`

行为：

1. 选择失败 Todo
2. 将其对应 Todo 状态改为 `ready`
3. 新建 execution record，状态为 `queued`
4. Session 若之前为 `failed`，则恢复为 `executing`

流转示意：

```text
session: failed -> executing
failed todo -> ready
new execution -> queued -> running
```

### 12.3 Retry 约束

- 只能重试失败或 blocked 的 Todo
- 已 done 的 Todo 不应重复重试，除非显式支持“强制重跑”
- retry 必须保留历史 execution 记录

---

## 13. Cancel 规则

### 13.1 Session cancel
一旦 session 被取消：

- session -> `canceled`
- 未执行 Todo -> `canceled`
- queued / running execution -> 尽量标记为 `canceled`
- 不再调度新任务

### 13.2 执行中的取消
如果某执行项无法立即中断，V1 可以采用：

- 先将 session 标记为 `canceled`
- 等运行中的任务自然结束
- 忽略其后续结果，不再触发新步骤

这是可接受的 V1 简化策略。

---

## 14. API 与状态机映射

### 14.1 `POST /sessions`

- 创建 session
- 最终进入 `waiting_prd_confirm` 或 `failed`

### 14.2 `POST /sessions/{id}/edit-prd`

- 仅在 `waiting_prd_confirm` 可用
- 状态不前进

### 14.3 `POST /sessions/{id}/confirm-prd`

- 从 `waiting_prd_confirm` 进入 `waiting_todo_confirm`

### 14.4 `POST /sessions/{id}/edit-todo`

- 仅在 `waiting_todo_confirm` 可用
- 状态不前进

### 14.5 `POST /sessions/{id}/confirm-todo`

- 从 `waiting_todo_confirm` 进入 `executing`

### 14.6 `POST /sessions/{id}/executions/retry`

- 仅在存在可重试 Todo 时有效
- 若 session 为 `failed`，可回到 `executing`

### 14.7 `POST /sessions/{id}/cancel`

- 在允许阶段进入 `canceled`

---

## 15. 建议的实现判断条件

### 15.1 Session 进入 done
建议由统一聚合器判断，而不是零散在多个 handler 中判断。

伪逻辑：

```text
if all required todos are done
  and no running executions
  and preview exists
then session = done
```

### 15.2 Session 进入 failed

```text
if critical todo failed and no retry path
  or final aggregation failed
then session = failed
```

### 15.3 Todo 进入 ready

```text
if all depends_on todos are done
  and todo not canceled
then todo = ready
```

---

## 16. 标准案例下的状态演进示例

以“用户提交产品 URL，生成商品素材与预览页”为例：

### Step 1
用户提交 URL：

```text
none -> input_received -> waiting_prd_confirm
```

### Step 2
用户确认 PRD：

```text
waiting_prd_confirm -> waiting_todo_confirm
```

### Step 3
用户确认 Todo：

```text
waiting_todo_confirm -> executing
```

### Step 4
并行执行：

- `生成标题` todo: ready -> running -> done
- `生成图片描述图` todo: ready -> running -> done
- `生成轮播图内容` todo: ready -> running -> done
- `生成视频脚本` todo: ready -> running -> done

### Step 5
汇总 preview：

- `build preview payload`: pending -> ready -> running -> done

### Step 6
整体完成：

```text
executing -> done
```

---

## 17. 结论

Agent Runtime 的状态机设计应坚持以下原则：

- Session 作为主状态机
- PRD / Todo 确认节点显式化
- Todo 与 Execution 子状态分离
- Retry 保留历史记录，不覆盖旧执行
- Cancel 优先停止调度，而不是追求绝对即时中断
- done / failed / canceled 作为稳定终态

这套状态机已经足够支撑 V1 的主流程与标准案例，并可为后续真正实现 API、Store、Executor、Scheduler 提供清晰边界。
