# game-stack

基于 [due](https://github.com/dobyte/due) (v2.5.5) 封装的分布式游戏服务端框架 SDK，开箱即用。

## 快速开始

```bash
# 获取依赖
bash update_due.sh

# 启动基础设施（etcd + Redis）
docker compose -f docker/docker-compose.yaml up -d

# 启动网关（WebSocket）
go run cmd/gate/main.go

# 启动逻辑服（14 个游戏模块）
go run cmd/node/main.go
```

## 架构分层

```
cmd/        入口点（gate、node）
stack/      核心 SDK：应用启动、Module 接口、路由、错误码、中间件、服务注册、延迟清理器
module/     14 个可插拔游戏模块
protocol/   消息结构体定义（Go 结构体 + json/msgpack 标签）
```

## Actor 模型（重点）

基于 due 的 Actor 系统，实现玩家状态的串行化管理和断线数据安全保障。

### 为什么需要 Actor

分布式游戏服务器面临两个核心问题：

1. **并发竞争**：同一玩家的多个请求可能并发到达，同时修改金币/背包/战斗状态
2. **断线幽灵消息**：玩家断线→旧节点消息队列中仍有待处理操作→新节点已创建新状态→状态变更丢失

### 解决方案

每个在线玩家拥有一个 `PlayerActor`（`kind="player"`）。所有状态修改操作必须通过 Actor 执行。

**生命周期**：

```
登录 → SpawnPlayerActor → BindActor → 处理消息
断线 → KillPlayerActor → mailbox 关闭 → 待处理消息全部丢弃
重连 → 全新 Actor 创建，完全独立的状态
```

### 两种使用模式

#### 模式1：Invoke（fire-and-forget）

适用于不需要同步返回结果的操作。

```go
// 在 Actor 中串行执行状态修改
actor.Invoke(ctx.Proxy(), uid, func() {
    playerSvc.DeductGold(uid, 100)
    inventorySvc.AddItem(uid, itemID, count)
})
// Invoke 立即返回，函数会排队在 Actor 中执行
```

**流程**：`actor.Invoke(fn)` → `Actor.fnChan` → `dispatch goroutine` 串行执行

**断线时**：`KillPlayerActor` → `fnChan` 关闭 → 所有未执行的 `Invoke` 被丢弃。

#### 模式2：RouteToActor（同步请求-响应）

适用于需要同步返回结果的操作（购买、交易、提交任务等）。

**模块 Init 时做两件事**：

```go
func (m *shopModule) Init(proxy *node.Proxy) error {
    impl := newImpl(o.store)

    // ① Node 路由用 RouteToActor 做转发
    proxy.AddRouteHandler(stack.RouteShopBuy,
        actor.RouteToActor(actor.KindPlayer),
        stack.StatefulAuthorizedRoute,
    )

    // ② 注册 Actor 路由初始化器（每个玩家 Actor Spawn 时自动调用）
    actor.RegisterRouteInitializer(func(act *node.Actor) {
        act.AddRouteHandler(stack.RouteShopBuy, impl.handleBuyActor)
    })
}
```

**Actor 中的处理器**：

```go
func (i *impl) handleBuyActor(ctx node.Context) {
    // 运行在 Actor 的 dispatch goroutine 中，天然线程安全
    req := &pshop.BuyRequest{}
    ctx.Parse(req)

    // 扣金币 → 扣库存 → 加道具，全部在同一 goroutine 串行
    playerSvc.DeductGold(ctx.UID(), price)
    shopStore.BuyItem(ctx, ctx.UID(), itemID, count)

    ctx.Response(OK)  // 同步响应
}
```

**完整消息流**：

```
客户端 → Gate → Node(StatefulRoute) → RouteToActor → Actor.mailbox → handleBuyActor → ctx.Response()
```

### API 一览

| 函数 | 用途 | 调用时机 |
|------|------|---------|
| `actor.SpawnPlayer(proxy, uid)` | 创建并绑定玩家 Actor | 登录成功 |
| `actor.KillPlayer(proxy, uid)` | 杀死 Actor，关闭 mailbox | 断线 |
| `actor.Invoke(proxy, uid, fn)` | 投递函数到 Actor 串行执行 | 状态修改 |
| `actor.RouteToActor(kind)` | 返回 Node 路由处理器，转发到 Actor | 路由注册时 |
| `actor.RegisterRouteInitializer(fn)` | 注册 Actor 路由初始化器 | 模块 Init 时 |
| `actor.Get(proxy, uid)` | 获取 PlayerActor（nil=不在线） | 任意 |

## Module 接口

```go
type Module interface {
    Name() string
    Init(proxy *node.Proxy) error
}
```

`Init()` 在应用启动时调用。模块通过 `proxy.AddRouteHandler()` 注册路由，通过 `proxy.AddEventHandler()` 注册事件，通过 `stack.RegisterService()` 暴露跨模块服务。

## 模块间通信

| 场景 | 方式 | 示例 |
|------|------|------|
| **同节点调用** | `stack.RegisterService(name, svc)` + `stack.GetService(name)` | 商城扣金币 |
| **跨节点事件** | EventBus（Redis pub/sub） | 玩家升级通知 |
| **客户端消息** | WebSocket + StatefulRoute | 所有玩家操作 |

## 路由编号

所有路由 ID 集中定义在 `stack/route.go`，每模块预留 100 个号段：

| 模块 | 号段 | 模块 | 号段 |
|------|------|------|------|
| auth | 1–99 | combat | 701–799 |
| player | 101–199 | guild | 801–899 |
| chat | 201–299 | mail | 901–999 |
| match | 301–399 | shop | 1001–1099 |
| room | 401–499 | leaderboard | 1101–1199 |
| inventory | 501–599 | activity | 1201–1299 |
| quest | 601–699 | social | 1301–1399 |

## 路由选项

| 选项 | 含义 | 使用场景 |
|------|------|---------|
| 无选项 | 无状态、无需认证 | 登录、注册 |
| `node.AuthorizedRoute` | 需认证（UID 不为空） | 全局只读数据 |
| `stack.StatefulAuthorizedRoute` | 有状态 + 需认证 | 所有需要玩家数据的操作 |

## 错误码

系统错误 (0–999) 定义在 `stack/errcode.go`。业务错误按模块分配 1000+，每模块预留 100 个号段。

## 断线清理

使用 `PlayerDoneCleaner`（`stack/cleaner.go`）实现 30 秒 Grace Period：

```
断线 → 立即 UnbindNode + DeleteToken + KillActor
     → 30 秒后：proxy.AskNode() 确认玩家不在本节点 → 清理内存
重连 → SpawnPlayer + Cleaner.OnLogin() → 取消旧清理定时器
```

## 模块文件规范

每个模块遵循 6 文件模式：

| 文件 | 用途 |
|------|------|
| `store.go` | 数据类型 + `Store` 接口（抽象持久化） |
| `store_memory.go` | 默认内存 `Store` 实现（`map` + `sync.RWMutex`） |
| `option.go` | 函数式选项（`WithStore(s Store)` 等） |
| `service.go` | `Service` 接口，供其他模块调用 |
| `impl.go` | 路由处理器实现 |
| `module.go` | `Module()` 构造函数 + `Init()` 注册路由和事件 |

## 添加新模块步骤

1. 创建 `protocol/<name>/message.go` — 请求/响应结构体
2. 创建 `module/<name>/` 按 6 文件模式
3. 在 `stack/route.go` 添加路由常量
4. 在 `stack/errcode.go` 添加错误码（如需要）
5. 导入并添加到 `cmd/node/main.go`

## 依赖

- 核心：`github.com/dobyte/due/v2` (v2.5.5)
- 子模块：`due/locate/redis/v2`, `due/network/ws/v2`, `due/registry/etcd/v2`, `due/transport/grpc/v2`, `due/eventbus/redis/v2`
