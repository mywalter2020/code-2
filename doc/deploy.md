# Deployment Guide

## Current release posture

当前版本已具备：

- 可编译、可测试、可通过 Docker 启动
- 任务编排 / 人工确认 / 任务查询闭环
- memory / sqlite / postgres 三种存储模式
- API Key 写操作鉴权
- provider 运行时切换

当前版本仍需注意：

- 平台 adapter 已具备 live HTTP 调用骨架，但默认仍建议先以 dry-run 验证
- Alibaba 已加入签名 envelope 与字段映射骨架（app_key / method / timestamp / v / sign / item）
- 若要真实执行发布，必须补齐平台 API 协议、鉴权签名与正式凭据

## Recommended production baseline

建议至少使用以下基线：

- `JUYU_STORE=postgres`
- 设置 `JUYU_API_KEY`
- 若未接真实平台，保持 `JUYU_ADAPTER_DRY_RUN=true`
- 通过反向代理或网关统一接入 TLS
- 挂载持久化 Postgres 数据卷

## Required environment variables

### Core

- `JUYU_CONFIG`：配置文件路径
- `JUYU_STORE`：`memory | sqlite | postgres`
- `JUYU_API_KEY`：写接口鉴权 key，生产环境强烈建议设置

### Store

当 `JUYU_STORE=sqlite` 时：

- `JUYU_SQLITE_PATH`

当 `JUYU_STORE=postgres` 时：

- `JUYU_PG_DSN`

### Adapter

- `JUYU_ADAPTER_DRY_RUN`
- `JUYU_ALIBABA_APP_KEY`
- `JUYU_ALIBABA_SECRET`
- `JUYU_ALIBABA_BASE_URL`
- `JUYU_ALIBABA_SIGN_METHOD`
- `JUYU_ALIBABA_VERSION`
- `JUYU_TAOBAO_APP_KEY`
- `JUYU_TAOBAO_SECRET`
- `JUYU_TAOBAO_BASE_URL`
- `JUYU_DOUYIN_CLIENT_ID`
- `JUYU_DOUYIN_CLIENT_SECRET`
- `JUYU_DOUYIN_BASE_URL`

### LLM Providers

NVIDIA:

- `NVIDIA_URL`
- `NVIDIA_KEY`
- `NVIDIA_MODEL`

OpenAI-compatible:

- `OPENAI_COMPAT_URL`
- `OPENAI_COMPAT_KEY`
- `OPENAI_COMPAT_MODEL`

## Runtime validation rules

应用启动前会校验：

- config 文件结构是否合法
- binding 是否引用了存在的 master / ability
- workflow step 是否重复或缺失
- `JUYU_STORE` 是否合法
- `sqlite / postgres` 所需参数是否齐全
- 当 `JUYU_ADAPTER_DRY_RUN=false` 时，`JUYU_API_KEY` 是否已设置
- adapter 在 live 模式下是否配置了平台 base URL

校验失败时，服务会直接启动失败，而不是带病运行。

## Docker Compose startup

```bash
cp .env.example .env
# 编辑 .env

docker compose up --build -d
curl http://127.0.0.1:8080/healthz
```

## Release checklist

上线前至少确认：

- [ ] `go test ./...` 通过
- [ ] `go vet ./...` 通过
- [ ] Docker 镜像可构建
- [ ] `healthz` 正常
- [ ] `JUYU_API_KEY` 已配置
- [ ] 存储不是 memory（正式环境）
- [ ] 若需要真实发布：平台凭据、base URL、鉴权签名与真实 adapter 已完成联调
- [ ] 若需要真实 LLM：provider 凭据已配置并验证
