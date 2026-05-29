# 仓储设计

## 概述

仓储（Repository）是聚合的持久化抽象，以**聚合为单位**进行加载和保存，不暴露内部实体。

- 接口定义在 `domain/repository.go`（领域层）
- 实现在 `infrastructure/repo_memory.go`（基础设施层）
- 通过泛型 `ddd.Repository[T]` 提供标准的 Load/Save/Delete

## DDD 四层架构中的位置

```
module/<name>/
├── domain/
│   ├── aggregate.go       # 聚合根 + 数据类型
│   └── repository.go      # ← 仓储接口（内嵌 ddd.Repository[T]）
├── infrastructure/
│   └── repo_memory.go     # ← 内存仓储实现
└── module.go              # 依赖注入：repo := o.repo
```

## 接口定义（以 player 为例）

```go
// domain/repository.go
package domain

import (
    "context"
    "github.com/skeletongo/game-stack/ddd"
)

// PlayerRepository 是 Player 聚合的仓储接口。
type PlayerRepository interface {
    ddd.Repository[*Player]  // 内嵌泛型接口：Load / Save / Delete

    // BC 特有查询
    FindByNickname(ctx context.Context, nickname string) (*Player, error)
}
```

`ddd.Repository[T]` 定义为：

```go
// ddd/repository.go
type Repository[T Aggregate] interface {
    Load(ctx context.Context, id int64) (T, error)
    Save(ctx context.Context, aggregate T) error
    Delete(ctx context.Context, id int64) error
}
```

## 内存实现

```go
// infrastructure/repo_memory.go
package infrastructure

type MemoryRepo struct {
    mu      sync.RWMutex
    players map[int64]*domain.Player
    nickname map[string]int64
}

func NewMemoryRepo() *MemoryRepo {
    return &MemoryRepo{
        players:  make(map[int64]*domain.Player),
        nickname: make(map[string]int64),
    }
}

func (r *MemoryRepo) Load(_ context.Context, id int64) (*domain.Player, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    p, ok := r.players[id]
    if !ok {
        return nil, fmt.Errorf("player %d not found", id)
    }
    return p, nil
}

func (r *MemoryRepo) Save(_ context.Context, p *domain.Player) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.players[p.ID()] = p
    r.nickname[p.Nickname().String()] = p.ID()
    return nil
}

func (r *MemoryRepo) Delete(_ context.Context, id int64) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    p, ok := r.players[id]
    if !ok {
        return nil  // 幂等
    }
    delete(r.nickname, p.Nickname().String())
    delete(r.players, id)
    return nil
}
```

## 函数式选项注入

```go
// module/<name>/option.go（模块根目录）
package player

type options struct {
    repo domain.PlayerRepository
}

func defaultOptions() *options {
    return &options{
        repo: infrastructure.NewMemoryRepo(),
    }
}

type Option func(o *options)

func WithRepository(r domain.PlayerRepository) Option {
    return func(o *options) { o.repo = r }
}
```

使用：

```go
player.Module(player.WithRepository(myRedisRepo))
```

## 仓储与 Actor 的职责分工

仓储 **不负责串行化**。`MemoryRepo` 的 `RWMutex` 只保护 map 的单次读写安全，不保护 read-modify-write 的原子性。

| 职责 | 负责方 |
|------|--------|
| 单次读写的线程安全 | 仓储的 `sync.RWMutex` |
| RMW 操作的原子性 | Actor 串行化 |
| 跨聚合操作的事务性 | Actor 串行化 |
| 并发修改安全 | Actor 串行化 |

写操作用 `RouteToActor` 进入 Actor 上下文，在 Actor goroutine 内调用仓储方法。详见 `docs/用户数据并发修改安全设计.md`。

## 实现约束

### 1. 删除操作需幂等

`CleanPlayerData` 会重试，已删除的聚合再次删除不应报错：

```go
func (r *MemoryRepo) Delete(_ context.Context, id int64) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    if _, ok := r.data[id]; !ok {
        return nil // 已不存在，不是错误
    }
    delete(r.data, id)
    return nil
}
```

### 2. 查询需兼容重新加载

Grace Period 过期后内存数据可能已被清理，查询时若内存中不存在应从持久化存储重新加载：

```go
func (r *Repo) Load(ctx context.Context, id int64) (*Player, error) {
    r.mu.RLock()
    p, ok := r.players[id]
    r.mu.RUnlock()
    if ok {
        return p, nil
    }
    return r.loadFromDB(ctx, id)  // 从持久存储恢复
}
```

详见 `docs/用户延迟登出设计.md`。

### 3. 生产环境替换

生产环境通过实现 `domain.Repository` 接口，注入对应实现：

```go
// Redis 实现
type RedisRepo struct {
    rdb *redis.Client
}

func NewRedisRepo(rdb *redis.Client) domain.PlayerRepository {
    return &RedisRepo{rdb: rdb}
}
```

## 相关文档

- `docs/DDD设计文档.md` — DDD 分层架构和开发规范
- `docs/用户数据并发修改安全设计.md` — Actor 串行化模型
- `docs/用户延迟登出设计.md` — Grace Period 与 CleanPlayerData
