# Engineering Notes

## Current project status

The repository now includes:

- multi-master-agent routing
- ability-agent registry
- config-driven workflows
- task flow with confirmation
- preview + logs + failure states
- 商品 / 发布 / 确认业务对象模型
- 最小前端演示骨架（/ui/，已支持任务状态徽标、业务预览卡片、确认操作）
- HTTP service endpoints
- local Go toolchain support
- Makefile / Dockerfile / docker-compose / .gitignore
- memory / sqlite / postgres storage modes
- docker-compose 联调已跑通（app + postgres）
- app healthcheck 已改为 curl 并在镜像内补齐依赖
- 新增可选写操作 API Key 鉴权（JUYU_API_KEY）
- 任务列表新增 status / scene / operator / platform / q 多条件筛选
- adapter 层升级为统一基类，支持 descriptor / configured / dry_run / missing_fields / base_url_configured / live_ready
- adapter 凭据改为从环境变量加载，便于后续替换成密钥中心
- content_gen 新增可选 NVIDIA Chat Completions 集成，未配置时自动回退 stub
- content_gen 的 model / temperature / max_tokens / system_prompt / prompt_template 已参数化
- content_gen 已抽象为通用 provider 接口，当前支持 stub / nvidia，后续可继续接 OpenAI / 百炼 / 自建模型
- page_gen 也已迁到同一套 provider 架构，支持 stub / nvidia，并支持 JSON 页面草图生成
- 新增第二个真实 provider 骨架 `openai_compat`，可接 OpenAI / 硅基流动 / 本地兼容网关
- 新增 provider 管理接口 `GET/PUT /admin/providers` 与前端配置面板，可在运行时切换 content_gen / page_gen 的 provider 与核心参数

## Suggested next steps

1. Add authentication and operator identity
2. Add real adapters for platform APIs
3. Add unit tests for server/store layers
4. Add OpenAPI/Swagger docs
5. Complete platform-specific auth/signature protocol for at least one live adapter
6. Add deployment manifests and CI
