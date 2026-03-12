# JuYu AI Platform Code

当前代码已提供一个更完整的 Go 原型，包含：

- 多主控 Agent 路由
- 多能力 Agent 注册中心
- 配置中心驱动绑定关系
- workflow 顺序执行
- 人工确认节点标记
- HTTP API 服务入口
- 基础健康检查与能力列表接口

## 目录

- `cmd/platform`：HTTP 服务入口
- `internal/config`：配置加载
- `internal/router`：场景到大 Agent 路由
- `internal/registry`：小 Agent 注册中心
- `internal/orchestrator`：编排执行
- `internal/agents`：能力 Agent 示例实现
- `internal/server`：HTTP 接口层
- `internal/types`：公共类型

## 接口

### `GET /healthz`
健康检查

### `GET /abilities`
查看当前注册的小 Agent 能力列表

### `POST /execute`
按场景触发大 Agent 编排执行

示例请求：

```json
{
  "scene": "product",
  "input": "在阿里平台生成商品页面并准备上架",
  "payload": {
    "platform": "alibaba"
  }
}
```

## 运行

在工作区根目录执行：

```bash
go run ./code/cmd/platform
```

配置文件默认自动查找：

```bash
configs/agents.yaml
../configs/agents.yaml
```

也可以手动指定：

```bash
JUYU_CONFIG=configs/agents.yaml go run ./code/cmd/platform
```

## 当前说明

当前实现还是原型版：
- 小 Agent 先用 stub 返回结果
- 尚未接入真实模型或真实平台 API
- 尚未支持持久化、鉴权、并行执行、任务队列

但骨架已经适合作为后续继续开发的基础项目结构。
