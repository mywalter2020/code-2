# CHANGELOG

## 2026-03 JuYu runtime hardening and validation

### 1. 服务监听地址支持配置化
**问题**
- 服务监听地址写死为 `:8080`
- 现网实例占用 8080 时，无法起第二套实例做隔离联调
- 难以区分“代码问题”和“旧容器/旧配置问题”

**修复**
- 提交：`d4bd8d7 feat(platform): allow configurable listen address`
- 将监听地址改为 `JUYU_ADDR`，默认仍为 `:8080`

**结果**
- 可在 `:8081` 起隔离实例
- 成功完成 DashScope 真链路验证
- 明确定位 8080 落后根因来自旧镜像，而非代码不可运行

---

### 2. 8080 Docker 实例与当前代码脱节
**问题**
- 8080 上运行的是旧镜像，缺少当前代码已具备的接口能力（如 `/admin/providers`）
- Docker 容器长时间未重建，运行态落后于仓库代码

**修复**
- 重建并重启：`docker compose up --build -d`
- 核验容器镜像构建时间、路由能力与当前 workspace 代码一致性

**结果**
- 8080 Docker 实例已对齐当前代码
- `/admin/providers` 恢复可用
- 运行态从“旧版本镜像”切换到“当前分支版本”

---

### 3. 写接口鉴权在重建后丢失
**问题**
- 容器重建后 `JUYU_API_KEY` 未稳定注入
- 导致 `api_key_enabled: false`

**修复**
- 建立并整理 `.env`
- 将 `JUYU_API_KEY` 固化到 `.env`
- 以 `.env` 作为 Compose 的稳定配置源

**结果**
- 写接口鉴权恢复
- 当前写接口统一通过 `X-API-Key` 校验

---

### 4. 默认 provider 不适合当前联调
**问题**
- 既有 NVIDIA provider 在当前环境下存在超时与联调不稳定问题
- 无法稳定支撑当前验证节奏

**修复**
- 将 `content_gen` / `page_gen` 切换为 `openai_compat`
- 接入 DashScope coding endpoint
- 使用模型：`qwen3-coder-plus`

**结果**
- `content_gen` 真生成成功
- `page_gen` 真生成成功
- 8080 / 8081 两套验证链路均可稳定跑通到人工确认节点

---

### 5. Docker Compose 污染 prompt template
**问题**
- 容器内 prompt template 被污染，出现重复片段和脏尾巴
- 根因是 Compose 对模板里的占位片段发生变量插值误判

**修复**
- 提交：`f104511 fix(compose): preserve prompt templates from env file`
- 为 `app` 添加 `env_file: .env`
- 将 `CONTENT_GEN_PROMPT_TEMPLATE` / `PAGE_GEN_PROMPT_TEMPLATE` 从 `environment:` 显式映射移除

**结果**
- prompt template 原样进入容器
- `/healthz` 与 `/admin/providers` 中模板值恢复干净
- DashScope 生成链不再携带污染 prompt

---

### 6. 配置与密钥管理缺乏正式治理
**问题**
- `.env` / 模板边界不清
- 密钥容易误入仓库
- README 缺少阶段性完成态和后续路线

**修复**
- 提交：`29e5d64 docs(config): formalize env template and secret handling`
- `.gitignore` 增加 `.env` / `.env.*`，保留 `.env.example`
- `.env.example` 重构为正式模板
- `README.md` 增加密钥管理约定、当前完成态、下一阶段待办

**结果**
- `.env.example` 入库，`.env` 不入库
- 配置模板分层清晰
- 当前状态可交接、可复盘、可持续维护

---

### 7. 主链路需要硬验证而非口头判断
**问题**
- 仅靠代码阅读或健康检查不能证明系统真的可用

**修复**
- 实际完成多轮验证：
  - `go test ./...`
  - 8081 隔离实例验证
  - 8080 Docker 实例验证
  - `/execute`
  - `/tasks/{id}/confirm`

**结果**
- 已验证闭环：
  - `content_gen`
  - `page_gen`
  - `review_check`
  - `publish_exec`（dry_run）
  - `onshelf_exec`（dry_run）
- 任务最终状态可达 `success`

---

## 当前阶段结论

JuYu 当前已达到：
- 可运行
- 可验证
- 可演示
- 可继续迭代

但尚未达到：
- 真实平台 adapter fully live
- 生产级 secrets 管理完善
- 完整观测面口径一致
