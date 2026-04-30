# CLAUDE.md

本文件为 Claude Code（claude.ai/code）在此仓库中工作时提供指导。

## 项目概述

`game-stack` 是一个分布式游戏服务器框架 SDK，基于 **due v2.5.5**（`github.com/dobyte/due/v2`）构建。它将 due 的基础组件（gateway、node、registry、transport、eventbus）封装为可插拔的模块系统，提供统一的路由/错误/中间件管理，并内置 14 个游戏模块。

模块：`github.com/skeletongo/game-stack` | Go 1.25 | 目标组件：WebSocket gateway + etcd registry + Redis locate/eventbus + gRPC transport

## 构建与验证

```bash
# 获取/重置依赖
bash update_due.sh

# 构建所有包
go build ./...

# 检查所有包
go vet ./...

# 启动开发基础设施（etcd + Redis）
docker-compose -f docker/docker-compose.yaml up -d
```

目前尚无测试。`go build ./...` 和 `go vet ./...` 是主要的验证命令。

## 架构

### 分层

```
cmd/         → 入口点（gate、node）—— 组装组件与模块
stack/       → 核心 SDK —— 应用启动、Module 接口、路由、错误、中间件、服务注册
module/      → 14 个可插拔游戏模块（auth、player、chat、match、room、inventory、quest、combat、guild、mail、shop、leaderboard、activity、social）
protocol/    → 消息结构体定义（带 json/msgpack 标签的 Go 结构体，无需 protoc）
```

### 模块模式（规范示例：`module/auth/`）

每个模块都是一个 6 文件的包，遵循以下固定模式：

| 文件 | 用途 |
|------|------|
| `store.go` | 数据类型 + `Store` 接口（抽象持久化） |
| `store_memory.go` | 默认的内存 `Store` 实现（`map` + `sync.RWMutex`） |
| `option.go` | 函数式选项（`WithStore(s Store)`） |
| `service.go` | 跨模块调用的 `Service` 接口 |
| `impl.go` | 路由处理器实现（`handleXxx(ctx node.Context)`） |
| `module.go` | 返回 `stack.Module` 的 `Module()` 构造函数，在 `Init()` 中注册路由和事件 |

### `stack.Module` 接口

```go
type Module interface {
    Name() string
    Init(proxy *node.Proxy) error
}
```

`Init()` 在应用启动时调用。模块通过 `proxy.AddRouteHandler()` 注册路由，通过 `proxy.AddEventHandler()` 注册事件，通过 `stack.RegisterService()` 暴露服务。

### 应用启动流程

1. `stack.NewApplication(options...)` —— 配置节点名称、定位器、注册中心、传输器、模块
2. `app.Run()` —— 创建 `due.Container`，构建 `node.Node`，对每个模块调用 `Init()`，调用 `container.Serve()`（阻塞直到 SIGTERM）

### Due v2.5.5 API 模式

**路由注册：**
```go
// RouteHandler = func(ctx Context)
proxy.AddRouteHandler(route int32, handler RouteHandler, opts ...RouteOptions)
// 关键选项：node.AuthorizedRoute（需要认证）、node.StatefulRoute、node.InternalRoute
```

**事件注册：**
```go
// EventHandler = func(ctx Context)
proxy.AddEventHandler(cluster.Connect, handler)  // cluster.Disconnect、cluster.Reconnect
// 重要：事件处理器不能调用 ctx.Response() —— 会返回错误
```

**Context 接口（`node.Context`）：**
- `Parse(v any) error` —— 反序列化请求体
- `Response(message any) error` —— 向客户端发送响应（仅路由可用）
- `UID() int64`、`CID() int64`、`GID() string` —— 会话标识符
- `BindGate(uid ...int64) error` —— 将会话绑定到用户（标记为已认证）
- `Disconnect(force ...bool) error` —— 强制断开连接

### 模块间通信

- **同节点**：在 `Init()` 中调用 `stack.RegisterService(name, svc)` → 通过类型断言使用 `stack.GetService(name)`
- **跨节点**：通过 `stack/event.go` 中的事件主题常量使用 Redis EventBus（例如 `EventPlayerLevelUp = "player:level_up"`）

### 路由编号

所有路由 ID 集中定义在 `stack/route.go` 中。每个模块分配一个 100 号的区间：

| 模块 | 中文名称 | 描述 | 区间 |
|------|----------|------|------|
| auth | 认证 | 用户登录、鉴权、会话管理 | 1–99 |
| player | 玩家 | 玩家基础信息、属性管理 | 101–199 |
| chat | 聊天 | 消息发送、频道管理 | 201–299 |
| match | 匹配 | 对局匹配队列 | 301–399 |
| room | 房间 | 房间创建、加入、离开 | 401–499 |
| inventory | 背包 | 物品存储、使用、装备 | 501–599 |
| quest | 任务 | 任务接取、进度推进、完成 | 601–699 |
| combat | 战斗 | 战斗逻辑与伤害计算 | 701–799 |
| guild | 公会 | 公会创建、管理、成员维护 | 801–899 |
| mail | 邮件 | 邮件收发、附件领取 | 901–999 |
| shop | 商店 | 商品浏览与购买 | 1001–1099 |
| leaderboard | 排行榜 | 排名数据管理与查询 | 1101–1199 |
| activity | 活动 | 限时活动参与与奖励 | 1201–1299 |
| social | 社交 | 好友与关系链管理 | 1301–1399 |

### 错误码

系统错误（0–999）定义在 `stack/errcode.go` 中，业务错误按模块定义（1000+）。所有响应通过 `stack.Respond()` 辅助函数使用信封格式 `{code, message, data}`。

### 添加新模块

1. 创建 `protocol/<name>/message.go` —— 请求/响应结构体
2. 按照 6 文件模式创建 `module/<name>/`
3. 在 `stack/route.go` 中添加路由常量（在其对应区间内）
4. 在 `stack/errcode.go` 中添加错误码（如有需要）
5. 导入并添加到 `cmd/node/main.go` 中

### Store 模式

每个模块通过 `WithStore(s Store)` 接受自定义的 `Store` 实现。默认是适合开发环境的内存存储。生产环境存储（Redis、MySQL）实现相同的接口。

### Due 框架依赖

核心模块：`github.com/dobyte/due/v2`（v2.5.5，所有 cluster/ 包在此模块下）
子模块（独立版本管理，使用 `@main`）：`due/locate/redis/v2`、`due/network/ws/v2`、`due/registry/etcd/v2`、`due/transport/grpc/v2`、`due/eventbus/redis/v2`、`due/component/http/v2`

注意：v2.5.5 中不存在 `due/v2/cluster/master`（v2.2.3 之后已移除）。无状态服务请使用 `cluster/mesh`。
