# Due v2.5.8 API 参考

## 应用启动流程

1. `stack.NewApplication(options...)` —— 配置节点名称、定位器、注册中心、传输器、模块
2. `app.Run()` —— 创建 `due.Container`，构建 `node.Node`，对每个模块调用 `Init()`，调用 `container.Serve()`（阻塞直到 SIGTERM）

## 路由注册

```go
// RouteHandler = func(ctx Context)
proxy.AddRouteHandler(route int32, handler RouteHandler, opts ...RouteOptions)
```

关键选项：

| 选项 | 说明 |
|------|------|
| `node.AuthorizedRoute` | 需要认证（请求必须携带合法 UID） |
| `node.StatefulRoute` | 有状态路由（同一玩家请求路由到固定节点） |
| `node.InternalRoute` | 内部路由（不对外暴露） |

`stack.StatefulAuthorizedRoute` 组合了 `Stateful + Authorized`，适用于大多数已登录路由。

## 事件注册

```go
// EventHandler = func(ctx Context)
proxy.AddEventHandler(cluster.Connect, handler)
proxy.AddEventHandler(cluster.Disconnect, handler)
proxy.AddEventHandler(cluster.Reconnect, handler)
```

注意：事件处理器中**不能**调用 `ctx.Response()`，会返回错误。

## Context 接口 (`node.Context`)

| 方法 | 说明 |
|------|------|
| `Parse(v any) error` | 反序列化请求体 |
| `Response(message any) error` | 向客户端发送响应（仅路由可用） |
| `UID() int64` | 获取用户 ID |
| `CID() int64` | 获取连接 ID |
| `GID() string` | 获取网关 ID |
| `BindGate(uid ...int64) error` | 将会话绑定到用户（标记为已认证） |
| `Disconnect(force ...bool) error` | 强制断开连接 |

## 节点间通信 (Proxy)

Node 的 `Proxy` 提供了多种跨节点通信方式：

### 消息投递

| 方法 | 用途 |
|------|------|
| `proxy.Deliver(ctx, &DeliverArgs{NID, UID, Message})` | 投递消息到指定节点或用户所在节点 |
| `proxy.Broadcast(ctx, &BroadcastArgs{Kind, Message})` | 广播消息到所有网关用户 |
| `proxy.Multicast(ctx, &MulticastArgs{Kind, Targets, Message})` | 组播消息到指定用户列表 |
| `proxy.Push(ctx, &PushArgs{Kind, Target, Message})` | 推送消息到单个用户 |

`Message` 结构体包含 `Route`（路由 ID）和 `Data`（消息体），接收方通过路由处理器处理。

### gRPC 服务调用 (Service Provider / Mesh Client)

模块在 `Init` 中通过 `module/<name>/rpc/server` 注册 gRPC 服务，其他节点通过 `module/<name>/rpc/client` 创建 Mesh Client 调用。规范示例：`module/player/rpc/server/server.go` 和 `module/player/rpc/client/client.go`。

**服务端：**

```go
// 在 Init 中调用
rpcserver.Register(name, proxy, repo)
```

**调用方 — 两种模式：**

```go
import (
    playerrpc "github.com/skeletongo/game-stack/module/player/rpc/client"
    playerpb "github.com/skeletongo/game-stack/module/player/rpc/grpc"
)

// 无状态调用（服务发现，随机节点）
client, _ := playerrpc.New(proxy)
resp, _ := client.GetPlayer(ctx, &playerpb.GetPlayerReq{Uid: uid})

// 面向玩家调用（定位到玩家所在节点，访问内存数据）
client, _ := playerrpc.NewWithUID(ctx, proxy, proxy.GetName(), uid)
resp, _ := client.GetPlayer(ctx, &playerpb.GetPlayerReq{Uid: uid})
```

NewMeshClient 支持三种寻址模式：

| 模式 | 示例 | 说明 |
|------|------|------|
| 直连（ID） | `direct://<node-id>` | 通过实例 ID 直连 |
| 直连（地址） | `direct://127.0.0.1:8011` | 通过地址直连 |
| 服务发现 | `discovery://service_name` | 通过 etcd 注册中心发现 |

### 内部路由 (InternalRoute)

仅允许 Gate 间流转或 Node 间 RPC 触发的路由，客户端不可访问：

```go
proxy.AddRouteHandler(route, handler, node.InternalRoute)
```

配合 `proxy.Deliver` 投递消息到目标节点，实现节点间协作操作。

## 框架依赖

核心模块：`github.com/dobyte/due/v2`（v2.5.8，所有 cluster/ 包在此模块下）

子模块（独立版本管理，使用 `@main`）：

| 包 | 用途 |
|----|------|
| `due/locate/redis/v2` | Redis 定位服务 |
| `due/network/ws/v2` | WebSocket 网络层 |
| `due/registry/etcd/v2` | etcd 服务注册 |
| `due/transport/grpc/v2` | gRPC 节点间通信 |
| `due/eventbus/redis/v2` | Redis 事件总线 |
| `due/component/http/v2` | HTTP 组件 |

注意：v2.5.8 中不存在 `due/v2/cluster/master`（v2.2.3 之后已移除）。无状态服务请使用 `cluster/mesh`。
