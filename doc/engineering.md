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

## Suggested next steps

1. Add authentication and operator identity
2. Add real adapters for platform APIs
3. Add unit tests for server/store layers
4. Add OpenAPI/Swagger docs
5. Add deployment manifests and CI
