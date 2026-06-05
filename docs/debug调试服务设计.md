# Debug 调试服务设计

## 概述

Debug 服务是一个内嵌 HTTP 调试端点，用于运行时查询和修改游戏数据。面向开发者，不走客户端协议、不走 proto。

**核心原则：通用、零侵入。** 模块只需一行 `debug.Register`，命令列表和查询能力全部自动发现。

## 架构

```
┌────────────────────────────────────┐
│            Debug HTTP              │  端口可配（默认关闭）
│  GET  /debug/modules               │
│  GET  /debug/module/:name          │
│  POST /debug/query                 │
│  POST /debug/command               │
│  POST /debug/patch                 │
│  GET  /debug/swagger.json          │  ← API 文档
│  GET  /debug/swagger/              │  ← Swagger UI
└──────────┬─────────────────────────┘
           │ 模块注册时注入
           ▼
┌──────────────────────────────────────┐
│  Module{Load, Save, CmdBus}          │
│  Load/Save 闭合泛型 Repository[T]    │
└──────────┬───────────────────────────┘
           │
           ▼
┌──────────────────────────────────────┐
│  ddd.Snapshot(agg)     → 反射读      │
│  ddd.ApplyPatch(agg)   → unsafe 写   │
│  ddd.Scalar            → VO 展开     │
│  ddd.CommandBus        → 命令调度     │
└──────────────────────────────────────┘
```

## 文件清单

| 文件 | 职责 |
|------|------|
| `stack/debug/registry.go` | 泛型 `Register[T]`，存 Load/Save 闭合函数 |
| `stack/debug/server.go` | HTTP server，路由注册 |
| `stack/debug/discover.go` | `GET /debug/modules` + `/debug/module/:name` |
| `stack/debug/query.go` | `POST /debug/query` |
| `stack/debug/command.go` | `POST /debug/command` |
| `stack/debug/patch.go` | `POST /debug/patch` |
| `stack/debug/swagger.go` | `//go:embed docs/swagger.json` + 内联 Swagger UI |
| `stack/debug/doc.go` | godoc + swag `@title` `@version` 注释 |
| `stack/debug/docs.sh` | `swag init --outputTypes json,yaml` 生成文档 |
| `ddd/snapshot.go` | `Snapshot()` + `ApplyPatch()` |
| `ddd/value_object.go` | `Scalar` 接口 |
| `ddd/command.go` | `Register` 泛型化，`NewCommand()` / `CommandNames()` |

---

## 端点

### 服务发现

`GET /debug/modules`

```
["player", "auth"]
```

`GET /debug/module/player`

```json
{
  "queries": ["get"],
  "commands": ["player.create", "player.get_info", "player.set_avatar", "player.add_exp",
               "player.add_gold", "player.deduct_gold", "player.add_diamond",
               "player.deduct_diamond", "player.delete"]
}
```

`commands` 直接从 `cmdBus.CommandNames()` 获取。若模块传 `nil` 给 cmdBus（如不希望暴露命令），响应中不含 `commands` 字段。

### 查询聚合

`POST /debug/query`

```json
→ {"module": "player", "id": 12345}

← {
  "id": 12345, "nickname": "test", "level": 5,
  "exp": 450, "gold": 9999, "diamond": 100,
  "avatar": "", "createdAt": 1717200000, "updatedAt": 1717200000
}
```

实现：`Module.Load(id)` → `ddd.Snapshot(agg)` → JSON。`Scalar` 类型自动展开（`Gold(999)` → `999`）。

### 执行命令

`POST /debug/command`

```json
→ {"module": "player", "cmd": "player.add_gold", "params": {"player_id": 12345, "amount": 999}}
← 999    （命令返回值：add_gold 返回 int32，JSON 序列化为数字）
```

无返回值命令（如 `player.set_avatar`）返回 `{}`（`ddd.NoResult{}` 序列化结果）。

实现：`cmdBus.NewCommand(name)` 获取零值 → `json.Unmarshal(params, cmd)` → `cmdBus.Dispatch(ctx, cmd)`。命令通过已有的 Handler 处理器执行并返回结果，自动发布领域事件。

### 直接修改内存

`POST /debug/patch`

```json
→ {"module": "player", "id": 12345, "fields": {"gold": 8888}}
← {"ok": true, "patched": ["gold"], "snapshot": { ... }}
```

实现：`Module.Load(id)` → `ddd.ApplyPatch(agg, fields)` → `Module.Save(agg)`。用 unsafe 绕过导出限制直接写非导出字段，不经过构造校验。

### API 文档

`GET /debug/swagger.json` — 由 `docs.sh` 生成的 OpenAPI 规范。

`GET /debug/swagger/` — Swagger UI 页面，CDN 加载，指向 `/debug/swagger.json`。

---

## 模块注册

```go
// module/player/module.go
import "github.com/skeletongo/game-stack/stack/debug"

func (m *playerModule) Init(proxy *node.Proxy) error {
    // ...
    debug.Register[*domain.Player]("player", repo, cmdBus)
}
```

`Register` 是泛型函数，`T` 从 repo 参数自动推导。内部闭合 `repo.Load` / `repo.Save` 为 `func(int64) (Aggregate, error)` 擦除泛型，存入注册表。

```go
// module/auth/module.go
debug.Register[*domain.Account]("auth", repo, cmdBus)
```

传 `nil` 给 cmdBus 时模块只开放 query 和 patch，不开放 command（如不希望 debug 端点执行敏感操作）。

### 注册表结构

```go
type Module struct {
    Name   string
    Load   func(ctx context.Context, id int64) (ddd.Aggregate, error)
    Save   func(ctx context.Context, agg ddd.Aggregate) error
    CmdBus *ddd.CommandBus
}
```

---

## DDD 层

### Scalar 接口

定义在 `ddd/value_object.go`。值对象实现后，`Snapshot` 自动展开为 JSON 原始类型。

```go
type Scalar interface {
    Scalar() any
}
```

当前实现：player 7 个 VO（`PlayerID` `Nickname` `Level` `Gold` `Diamond` `Exp` `Avatar`）+ auth 5 个 VO（`UserID` `Username` `PasswordHash` `Nickname` `Token`），每个一行。

### Snapshot — 反射读

```go
func Snapshot(agg Aggregate) map[string]any
```

遍历聚合所有非导出字段：`Scalar` 展开、`time.Time` 转 Unix、其他类型递归或取字符串。导出字段（protobuf 嵌入的 `state`/`sizeCache`）跳过。

### ApplyPatch — unsafe 写

```go
func ApplyPatch(agg Aggregate, fields map[string]any) error
```

`Snapshot` 的逆操作。`reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr()))` 绕过导出限制写入。

### CommandBus 工厂

```go
func Register[C Command, T any](bus *CommandBus, name string, handler CommandHandler[C, T])

func (b *CommandBus) NewCommand(name string) (any, bool)

func (b *CommandBus) CommandNames() []string

func Dispatch[T any](ctx context.Context, bus *CommandBus, cmd Command) (T, error)
```

`Register` 存零值工厂 `func() any { return new(C) }` + 调用闭包（注册时类型断言，Dispatch 时执行）。`NewCommand` 返回 `any`（`*Cmd` 零值指针）供 debug 端点反序列化参数。`Dispatch` 返回 `(any, error)`，类型安全的 `Dispatch[T any]` 辅助函数提供编译时类型推导。

---

## 启用

```go
// cmd/node/main.go
app := stack.NewApplication(
    stack.WithDebug("127.0.0.1:6060"),
    // ...
)
```

不传 `WithDebug` 不启动，零开销。默认只监听 `127.0.0.1`，无认证（开发工具靠端口绑定保证安全）。

## Swagger 生成

端点的 swag 注释写在 handler 函数上（`@Summary` `@Router` 等），API 全局信息写在 `doc.go`（`@title` `@version`）。运行：

```bash
bash stack/debug/docs.sh
```

`swag init` 扫描注释 → 生成 `docs/swagger.json` + `docs/swagger.yaml` → `swagger.go` 通过 `//go:embed` 编译进二进制。
