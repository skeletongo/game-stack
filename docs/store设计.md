# Store 设计

## 概述

Store 是模块的数据持久化接口，抽象了数据存取方式。每个模块通过 `Store` 接口隔离具体存储实现，默认提供内存实现（开发环境），生产环境可替换为 Redis/MySQL 等外部存储。

## 6 文件模式中的 Store

| 文件 | 职责 |
|------|------|
| `store.go` | 数据类型定义 + `Store` 接口 |
| `store_memory.go` | 默认内存实现（`map` + `sync.RWMutex`） |
| `option.go` | `WithStore(s Store)` 注入自定义实现 |

## 接口定义（以 auth 为例）

```go
// store.go
package auth

type User struct {
    ID        int64
    Username  string
    Password  string
    Nickname  string
    BannedAt  int64
    CreatedAt int64
}

type Store interface {
    CreateUser(ctx context.Context, user *User) error
    GetUserByID(ctx context.Context, id int64) (*User, error)
    GetUserByUsername(ctx context.Context, username string) (*User, error)
    UpdateUser(ctx context.Context, user *User) error
    BanUser(ctx context.Context, id int64) error
    UnbanUser(ctx context.Context, id int64) error
    SetToken(ctx context.Context, uid int64, token string) error
    GetToken(ctx context.Context, uid int64) (string, error)
    DeleteToken(ctx context.Context, uid int64) error
    GetTokenByValue(ctx context.Context, token string) (int64, error)
    SetOnline(ctx context.Context, uid int64, gid string) error
    SetOffline(ctx context.Context, uid int64) error
    IsOnline(ctx context.Context, uid int64) (bool, error)
    OnlineCount(ctx context.Context) (int64, error)
}
```

## 内存实现

```go
// store_memory.go
package auth

type memoryStore struct {
    mu       sync.RWMutex
    users    map[int64]*User
    username map[string]int64
    tokens   map[int64]string
    tokenRev map[string]int64
    online   map[int64]string
}

func newMemoryStore() *memoryStore {
    return &memoryStore{
        users:    make(map[int64]*User),
        username: make(map[string]int64),
        tokens:   make(map[int64]string),
        tokenRev: make(map[string]int64),
        online:   make(map[int64]string),
    }
}

func (s *memoryStore) CreateUser(_ context.Context, user *User) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.users[user.ID] = user
    s.username[user.Username] = user.ID
    return nil
}
```

## 函数式选项注入

```go
// option.go
package auth

type options struct {
    store Store
}

func defaultOptions() *options {
    return &options{store: newMemoryStore()}
}

type Option func(o *options)

func WithStore(s Store) Option {
    return func(o *options) { o.store = s }
}
```

使用时：

```go
auth.Module(auth.WithStore(myRedisStore))
```

## Store 与 Actor 的职责分工

Store **不负责串行化**。`memoryStore` 的 `RWMutex` 只保护 map 的单次读写安全，不保护 read-modify-write 的原子性。

| 职责 | 负责方 |
|------|--------|
| 单次读写的线程安全 | Store 的 `sync.RWMutex` |
| RMW 操作的原子性 | Actor 串行化 |
| 跨 Store 操作的事务性 | Actor 串行化 |
| 并发修改安全 | Actor 串行化 |

改写操作用 `RouteToActor` 或 `Invoke` 进入 Actor 上下文，在 Actor goroutine 内调用 Store 方法。详见 `docs/用户数据并发修改安全设计.md`。

## 实现约束

### 1. 删除操作需幂等

`CleanPlayerData` 会重试，已删除的数据再次删除不应报错：

```go
func (s *store) RemoveData(ctx context.Context, uid int64) error {
    if _, ok := s.data[uid]; !ok {
        return nil // 已不存在，不是错误
    }
    delete(s.data, uid)
    return nil
}
```

### 2. 查询需兼容重新加载

Grace Period 过期后内存数据可能已被清理，查询时若内存中不存在应从持久化存储重新加载：

```go
func (s *store) GetData(ctx context.Context, uid int64) (*Data, error) {
    s.mu.RLock()
    d, ok := s.data[uid]
    s.mu.RUnlock()
    if ok {
        return d, nil
    }
    return s.loadFromDB(ctx, uid)
}
```

详见 `docs/用户延迟登出设计.md`。

### 3. 生产环境替换

生产环境通过实现 `Store` 接口，注入对应实现即可。进程重启时数据自然恢复：

```go
// Redis 实现
type redisStore struct {
    rdb *redis.Client
}

func NewRedisStore(rdb *redis.Client) Store {
    return &redisStore{rdb: rdb}
}
```

## 相关文档

- `docs/模块开发规范.md` — 模块 6 文件模式
- `docs/用户数据并发修改安全设计.md` — Actor 串行化模型
- `docs/用户延迟登出设计.md` — Grace Period 与 CleanPlayerData
