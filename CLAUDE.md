# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`game-stack` is a distributed game server framework SDK built atop **due v2.5.5** (`github.com/dobyte/due/v2`). It wraps due's building blocks (gateway, node, registry, transport, eventbus) with a pluggable module system, unified route/error/middleware management, and 14 pre-built game modules.

Module: `github.com/skeletongo/game-stack` | Go 1.25 | Target components: WebSocket gateway + etcd registry + Redis locate/eventbus + gRPC transport

## Build & Verify

```bash
# Fetch/reset dependencies
bash update_due.sh

# Build all packages
go build ./...

# Vet all packages
go vet ./...

# Start dev infrastructure (etcd + Redis)
docker-compose -f docker/docker-compose.yaml up -d
```

There are no tests yet. `go build ./...` and `go vet ./...` are the primary verification commands.

## Architecture

### Layers

```
cmd/         → Entry points (gate, node) — assemble components and modules
stack/       → Core SDK — App bootstrap, Module interface, routes, errors, middleware, service registry
module/      → 14 pluggable game modules (auth, player, chat, match, room, inventory, quest, combat, guild, mail, shop, leaderboard, activity, social)
protocol/    → Message struct definitions (Go structs with json/msgpack tags, no protoc required)
```

### Module Pattern (canonical: `module/auth/`)

Every module is a 6-file package that follows this exact pattern:

| File | Purpose |
|------|---------|
| `store.go` | Data types + `Store` interface (abstracts persistence) |
| `store_memory.go` | Default in-memory `Store` implementation (`map` + `sync.RWMutex`) |
| `option.go` | Functional options (`WithStore(s Store)`) |
| `service.go` | `Service` interface for inter-module calls |
| `impl.go` | Route handler implementations (`handleXxx(ctx node.Context)`) |
| `module.go` | `Module()` constructor returning `stack.Module`, registers routes/events in `Init()` |

### `stack.Module` Interface

```go
type Module interface {
    Name() string
    Init(proxy *node.Proxy) error
}
```

`Init()` is called during app startup. Modules register routes via `proxy.AddRouteHandler()`, events via `proxy.AddEventHandler()`, and expose services via `stack.RegisterService()`.

### Application Bootstrap Flow

1. `stack.NewApplication(options...)` — configures node name, locator, registry, transporter, modules
2. `app.Run()` — creates `due.Container`, builds `node.Node`, calls `Init()` on each module, calls `container.Serve()` (blocks until SIGTERM)

### Due v2.5.5 API Patterns

**Route registration:**
```go
// RouteHandler = func(ctx Context)
proxy.AddRouteHandler(route int32, handler RouteHandler, opts ...RouteOptions)
// Key opts: node.AuthorizedRoute (requires auth), node.StatefulRoute, node.InternalRoute
```

**Event registration:**
```go
// EventHandler = func(ctx Context)
proxy.AddEventHandler(cluster.Connect, handler)  // cluster.Disconnect, cluster.Reconnect
// IMPORTANT: event handlers CANNOT call ctx.Response() — it returns an error
```

**Context interface (`node.Context`):**
- `Parse(v any) error` — deserialize request body
- `Response(message any) error` — send response to client (routes only)
- `UID() int64`, `CID() int64`, `GID() string` — session identifiers
- `BindGate(uid ...int64) error` — bind session to user (marks as authenticated)
- `Disconnect(force ...bool) error` — force disconnect

### Inter-Module Communication

- **Same node**: `stack.RegisterService(name, svc)` in `Init()` → `stack.GetService(name)` with type assertion
- **Cross-node**: Redis EventBus via event topic constants in `stack/event.go` (e.g., `EventPlayerLevelUp = "player:level_up"`)

### Route Numbering

All route IDs are centralized in `stack/route.go`. Each module gets a 100-number block:

| Module | Range |
|--------|-------|
| auth | 1–99 |
| player | 101–199 |
| chat | 201–299 |
| match | 301–399 |
| room | 401–499 |
| inventory | 501–599 |
| quest | 601–699 |
| combat | 701–799 |
| guild | 801–899 |
| mail | 901–999 |
| shop | 1001–1099 |
| leaderboard | 1101–1199 |
| activity | 1201–1299 |
| social | 1301–1399 |

### Error Codes

System errors (0–999) in `stack/errcode.go`, business errors per module (1000+). All responses use the envelope `{code, message, data}` via `stack.Respond()` helpers.

### Adding a New Module

1. Create `protocol/<name>/message.go` — request/response structs
2. Create `module/<name>/` with the 6-file pattern
3. Add route constants to `stack/route.go` (within its block)
4. Add error codes to `stack/errcode.go` (if any)
5. Import and add to `cmd/node/main.go`

### Store Pattern

Every module accepts a custom `Store` implementation via `WithStore(s Store)`. The default is an in-memory store suitable for development. Production stores (Redis, MySQL) implement the same interface.

### Due Framework Dependencies

Core module: `github.com/dobyte/due/v2` (v2.5.5, all cluster/ packages under this module)
Sub-modules (independent versioning, use `@main`): `due/locate/redis/v2`, `due/network/ws/v2`, `due/registry/etcd/v2`, `due/transport/grpc/v2`, `due/eventbus/redis/v2`, `due/component/http/v2`

Note: `due/v2/cluster/master` does NOT exist in v2.5.5 (removed after v2.2.3). Use `cluster/mesh` for stateless services.
