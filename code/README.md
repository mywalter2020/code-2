# JuYu AI Platform Code

当前代码提供一个可运行的 Go 原型，包含：

- 多主控 Agent 路由
- 多能力 Agent 注册
- 配置中心驱动绑定关系
- 基础编排器执行流程

## 目录

- `cmd/platform`：程序入口
- `internal/config`：配置加载
- `internal/router`：场景到大 Agent 路由
- `internal/registry`：小 Agent 注册中心
- `internal/orchestrator`：编排执行
- `internal/agents`：能力 Agent 示例实现
- `internal/types`：公共类型

## 运行

在工作区根目录执行：

```bash
go run ./code/cmd/platform
```

配置文件位于：

```bash
configs/agents.yaml
```
