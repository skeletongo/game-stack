# CLAUDE.md

本文件为 Claude Code（claude.ai/code）在此仓库中工作时提供指导。

## 项目概述

`game-stack` 是一个分布式游戏服务器框架，基于 **due v2.5.5**（`github.com/dobyte/due/v2`）构建。它将 due 的基础组件封装为可插拔的模块系统，提供统一的路由/错误/中间件管理，并内置游戏模块。

## 构建与验证

```bash
bash update_due.sh          # 获取/重置依赖
bash gen_proto.sh           # 生成所有模块的 proto 代码
go build ./...              # 构建所有包
go vet ./...                # 检查所有包
docker-compose -f docker/docker-compose.yaml up -d  # 启动开发基础设施（etcd + Redis）
```

目前尚无测试。`go build ./...` 和 `go vet ./...` 是主要的验证命令。

## 架构

```
docs/        → 详细设计文档
cmd/         → 入口点（gate、node）—— 组装组件与模块
stack/       → 核心框架 —— 应用启动、Module 接口、路由、错误、中间件、服务注册
module/      → 可插拔游戏模块（actor / auth / player）
proto/       → 客户端通信协议（protobuf 定义，客户端与服务端共用）
```

## 路由编号

公式：**模块号 × 1000 + 子协议号**。0–999 为系统内部预留。常量定义在 `stack/route.go`。详见 `docs/客户端服务端通信协议设计.md`。

## 错误码

系统错误（0–999）复用 HTTP 语义，业务错误采用 `模块号 × 1000 + 子码` 与路由统一。定义在 `stack/errcode.go`。详见 `docs/客户端服务端通信协议设计.md`。

## 模块开发

每个模块遵循 6 文件模式：`store.go`、`store_memory.go`、`option.go`、`service.go`、`impl.go`、`module.go`。规范示例：`module/auth/`。详见 `docs/模块开发规范.md`。

## 参考文档

- `docs/due-api.md` —— Due v2.5.5 API 参考（应用启动、路由/事件注册、Context 接口、框架依赖）
- `docs/客户端服务端通信协议设计.md` —— 路由编号方案、模块号分配、错误码体系、消息信封格式
- `docs/store设计.md` —— Store 接口抽象模式、内存实现、函数式选项注入、Actor 分工、删除幂等/查询重加载约束
- `docs/模块开发规范.md` —— 模块 6 文件模式、Service 模式、响应辅助函数、模块间通信、添加新模块流程
- `docs/用户绑定节点设计.md` —— 玩家节点绑定生命周期、有状态/无状态路由、Redis Locator 机制、Lua 脚本解绑安全性、断线重连节点切换
- `docs/用户延迟登出设计.md` —— 断线两阶段清理（立即安全清理 + Grace Period 延迟释放）、PlayerDoneCleaner 机制、Actor 竞态处理
- `docs/用户数据并发修改安全设计.md` —— Actor 串行化模型、RouteToActor/Invoke 两种模式、归属权校验、Service 调用约束
- `docs/go语言规范.md` —— Go 编码规范（注释、错误处理、工具库、命名、代码组织、并发）
