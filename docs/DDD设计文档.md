# DDD 设计文档

本文档定义 game-stack 框架的领域驱动设计（DDD）开发规范。在阅读本文前，请先理解现有架构（参考 `docs/` 下其他设计文档）。

## 目录

- [1. 战略设计](#1-战略设计)
  - [1.1 子域划分](#11-子域划分)
  - [1.2 限界上下文](#12-限界上下文)
  - [1.3 上下文映射](#13-上下文映射)
- [2. 战术设计](#2-战术设计)
  - [2.1 聚合](#21-聚合)
  - [2.2 实体](#22-实体)
  - [2.3 值对象](#23-值对象)
  - [2.4 领域事件](#24-领域事件)
  - [2.5 仓储](#25-仓储)
  - [2.6 领域服务](#26-领域服务)
  - [2.7 命令与应用服务](#27-命令与应用服务)
- [3. 分层架构](#3-分层架构)
  - [3.1 四层模型](#31-四层模型)
  - [3.2 目录结构](#32-目录结构)
  - [3.3 依赖方向](#33-依赖方向)
- [4. Actor 与聚合的关系](#4-actor-与聚合的关系)
- [5. 开发规范](#5-开发规范)
  - [5.1 模块文件清单](#51-模块文件清单)
  - [5.2 命名规范](#52-命名规范)
  - [5.3 添加新模块流程](#53-添加新模块流程)


---

## 1. 战略设计

### 1.1 子域划分

根据业务价值和技术复杂度，game-stack 的 14 个模块分为三类子域：

| 类型 | 模块 | 投入策略 |
|------|------|----------|
| ⭐ **核心域** | combat（战斗）、match（匹配）、guild（公会） | 自研，持续打磨差异化，投入最高人力 |
| 🔧 **支撑域** | player、inventory、quest、room、activity、leaderboard、social、mail、shop | 定制开发支撑核心域，投入适中 |
| 📦 **通用域** | auth（认证）、chat（聊天） | 标准化实现，可采购/复用，投入最低 |

### 1.2 限界上下文

每个模块对应一个限界上下文（Bounded Context）。BC 内部拥有独立的通用语言和领域模型。

| BC | 聚合根 | 核心职责 |
|----|--------|----------|
| auth | Account | 认证、令牌管理、会话生命周期 |
| player | Player | 玩家属性、等级、货币 |
| chat | ChatChannel | 消息收发、频道管理 |
| inventory | Backpack | 物品存储、使用、装备 |
| quest | QuestJournal | 任务接取、进度、完成 |
| combat | CombatInstance | 战斗逻辑、伤害计算、状态同步 |
| match | MatchQueue | 匹配排队、Elo 计算 |
| room | Room | 房间创建、加入、离开 |
| guild | Guild | 公会创建、管理、成员 |
| mail | Mailbox | 邮件收发、附件 |
| shop | Shop | 商品浏览、购买 |
| leaderboard | Leaderboard | 排名查询、更新 |
| activity | Activity | 限时活动、奖励领取 |
| social | SocialGraph | 好友、黑名单 |

### 1.3 上下文映射

限界上下文之间的集成关系：

```
Auth ──遵奉者──→ 所有 BC     （所有 BC 接受 Auth 提供的 UID 身份模型）
Player ←─合作关系─→ Inventory  （双向依赖，货币操作需跨 BC 协调）
Combat ──C/S──→ Player        （战斗修改玩家状态通过明确接口）
Shop ──C/S──→ Inventory,Player（商城是上游，定义扣费/发货契约）
Guild ──C/S──→ Player         （公会操作需确认玩家存在）
外部服务 ──ACL──→ 各 BC         （第三方支付/推送通过防腐层隔离）
EventBus ──发布/订阅──→ 跨 BC   （跨节点事件通过 due EventBus 传递）
gRPC ──C/S──→ 跨节点 BC       （跨节点服务调用通过 gRPC + Service Provider）
```

---

## 2. 战术设计

### 2.1 聚合

聚合是业务一致性边界。对聚合内任何实体的修改都必须通过聚合根。

**核心原则：**

- **小聚合**：聚合越小越好，一个 BC 可以有多个聚合
- **通过 ID 引用**：跨聚合引用只用 ID，不持有指针
- **最终一致性**：跨聚合操作通过领域事件实现最终一致
- **单次事务只改一个聚合**：一个命令只修改一个聚合实例

**game-stack 聚合示例：**

```
Player 聚合 (根: Player)
├── PlayerID (VO)
├── Nickname (VO, 1-16字符)
├── Level (VO, ≥1)
├── Gold (VO, ≥0)
├── Diamond (VO, ≥0)
├── Exp (VO, ≥0)
├── Avatar (VO)
└── 不变量: Gold≥0, Diamond≥0, Level=calcByExp(Exp)

Session 聚合 (根: Session)
├── UserID (VO)
├── Token (VO, 32字节随机 hex)
├── OnlineStatus (VO, online/offline)
└── 不变量: Token 全局唯一，同一 UserID 只有一个活跃 Session
```

### 2.2 实体

实体是拥有唯一标识的领域对象，标识在整个生命周期中保持不变。

```go
// ItemSlot 是背包中的一个物品槽位，属于 Backpack 聚合内部。
type ItemSlot struct {
    itemID   int64
    count    int32
    equipped bool
}

func (s *ItemSlot) ID() int64 { return s.itemID }
```

### 2.3 值对象

值对象无独立标识，由属性值定义相等性，创建后不可变。

```go
// Gold 是一个值对象，封装了金币的不变量（≥0）。
type Gold int32

func NewGold(amount int32) (Gold, error) {
    if amount < 0 {
        return 0, errors.New("gold must be ≥ 0")
    }
    return Gold(amount), nil
}

func (g Gold) Add(delta int32) (Gold, error) {
    return NewGold(int32(g) + delta)
}

func (g Gold) Equals(other ddd.ValueObject) bool {
    o, ok := other.(Gold)
    return ok && g == o
}
```

**何时用值对象而非原始类型：**

- 有业务校验规则（如 `Gold ≥ 0`）
- 有业务行为（如 `Level` 的升级计算）
- 需要防止 ID 混淆（如 `PlayerID` vs `ItemID`）

### 2.4 领域事件

领域事件解耦同一 BC 内不同组件。例如：`PlayerLeveledUp` 事件可触发 quest 模块检查新任务解锁。

**事件命名：过去式**（`PlayerLeveledUp`、`GoldDeducted`）。

```go
// PlayerLeveledUp 玩家升级领域事件。
type PlayerLeveledUp struct {
    playerID  int64
    oldLevel  int32
    newLevel  int32
    occurredAt time.Time
}

func (e PlayerLeveledUp) AggregateID() int64  { return e.playerID }
func (e PlayerLeveledUp) EventName() string   { return "player.leveled_up" }
func (e PlayerLeveledUp) OccurredAt() time.Time { return e.occurredAt }
```

**EventBus 使用方式：**

```go
// 在 module Init 中订阅
bus := ddd.NewEventBus()
bus.Subscribe("player.leveled_up", func(event ddd.DomainEvent) {
    e := event.(PlayerLeveledUp)
    // 触发 quest 进度更新等副作用
})

// 在聚合方法中发布
func (p *Player) AddExp(amount int64, bus *ddd.EventBus) error {
    oldLevel := p.level
    p.exp += amount
    p.level = calcLevel(p.exp)
    if p.level != oldLevel {
        bus.Publish(PlayerLeveledUp{
            playerID:   p.id,
            oldLevel:  int32(oldLevel),
            newLevel:  int32(p.level),
            occurredAt: time.Now(),
        })
    }
    return nil
}
```

### 2.5 仓储

仓储以聚合为单位进行持久化。接口定义在 `domain/` 层，实现在 `infrastructure/` 层。

```go
// domain/repository.go — 接口定义（领域层）

// PlayerRepository 是 Player 聚合的仓储接口。
type PlayerRepository interface {
    ddd.Repository[*Player]  // 内嵌泛型仓储接口
    // BC 特有查询
    FindByNickname(ctx context.Context, nickname string) (*Player, error)
}
```

```go
// infrastructure/repo_memory.go — 内存实现（基础设施层）

type playerMemoryRepo struct {
    mu      sync.RWMutex
    players map[int64]*Player
}

func (r *playerMemoryRepo) Load(ctx context.Context, id int64) (*Player, error) { ... }
func (r *playerMemoryRepo) Save(ctx context.Context, p *Player) error { ... }
func (r *playerMemoryRepo) Delete(ctx context.Context, id int64) error { ... }
```

**仓储实现约束（继承自旧 Store 模式）：**

- **删除幂等**：已删除的数据再次删除返回 nil
- **查询重加载**：内存中不存在时从持久存储恢复

### 2.6 领域服务

当操作不属于任一聚合时，使用领域服务（无状态）。

```go
// CalcLevelService 计算等级（纯函数，无副作用）。
// 放在 domain 层，因为它是领域逻辑的一部分。
type CalcLevelService struct{}

func (CalcLevelService) CalcLevel(exp int64) int32 {
    level := int32(1)
    needed := int64(100)
    for exp >= needed {
        exp -= needed
        level++
        needed = int64(level) * 100
    }
    return level
}
```

**何时用领域服务：**
- 操作涉及多个聚合（协调逻辑在应用层，纯领域计算在领域服务）
- 计算逻辑复杂且不属于任一实体/VO

### 2.7 命令与应用服务

命令封装了客户端意图。应用服务（CommandHandler）编排领域对象完成操作。

`CommandHandler[C, T any]` 接受命令类型 `C` 和返回结果类型 `T`。无返回值的命令使用 `ddd.NoResult`，有返回值的命令直接返回具体类型。

```go
// application/commands.go — 命令定义

// AddExpCmd 增加经验值命令。
type AddExpCmd struct {
    PlayerID int64
    Amount   int64
}

func (c AddExpCmd) CommandName() string { return "player.add_exp" }
```

```go
// application/handlers.go — 应用服务（返回最新经验值）

type AddExpHandler struct {
    playerRepo PlayerRepository
    eventBus   *ddd.EventBus
}

var _ ddd.CommandHandler[AddExpCmd, int64] = (*AddExpHandler)(nil)

func (h *AddExpHandler) Handle(ctx context.Context, cmd AddExpCmd) (int64, error) {
    p, err := h.playerRepo.Load(ctx, cmd.PlayerID)
    if err != nil {
        return 0, err
    }
    leveledUp := p.AddExp(cmd.Amount)
    if err := h.playerRepo.Save(ctx, p); err != nil {
        return 0, err
    }
    if leveledUp {
        h.eventBus.Publish(PlayerLeveledUp{...})
    }
    return p.Exp().Int64(), nil
}
```

**类型安全的命令分发**：

```go
// 接口层使用 ddd.Dispatch[T] 获取强类型结果
result, err := ddd.Dispatch[*LoginResult](ctx, cmdBus, cmd)

// 无返回值命令用 _ 丢弃 ddd.NoResult
_, err := cmdBus.Dispatch(ctx, cmd)
```

---

## 3. 分层架构

### 3.1 四层模型

```
┌──────────────────────────────────────────┐
│ Interface 层（接口层）                    │  ← 最薄层
│ 路由注册、proto 消息 ↔ 命令/DTO 转换      │
│ 不包含任何业务逻辑                        │
├──────────────────────────────────────────┤
│ Application 层（应用层）                  │
│ CommandHandler：编排聚合、仓储、事件总线    │
│ 不包含领域规则，只做任务协调               │
├──────────────────────────────────────────┤
│ Domain 层（领域层）                       │  ← 核心，不依赖任何外层
│ Aggregate、Entity、ValueObject            │
│ Domain Event、Repository 接口             │
│ Domain Service（无状态领域计算）           │
├──────────────────────────────────────────┤
│ Infrastructure 层（基础设施层）            │
│ Repository 实现（memory/redis/mysql）      │
│ EventBus 适配、外部服务适配器              │
└──────────────────────────────────────────┘
```

### 3.2 目录结构

四层代码放在 `internal/` 子目录中（Go 编译器强制边界），对外接口放在 `svc/`，RPC 适配放在顶层 `rpc/`。以 player 为例：

```
module/player/
├── internal/
│   ├── domain/                # 领域层
│   │   ├── aggregate.go       # Player 聚合根 + 行为方法
│   │   ├── value_objects.go   # PlayerID、Nickname、Level、Gold 等 VO
│   │   ├── events.go          # 领域事件定义
│   │   ├── repository.go      # PlayerRepository 接口
│   │   └── service.go         # 领域服务（CalcLevelService 等）
│   ├── application/           # 应用层
│   │   ├── commands.go        # 命令定义（AddExpCmd、UpdateProfileCmd 等）
│   │   ├── handlers.go        # CommandHandler 实现
│   │   └── dto.go             # DTO（proto 消息 ↔ 领域对象的转换）
│   ├── infrastructure/        # 基础设施层
│   │   ├── repo_memory.go     # 内存仓储实现
│   │   └── repo_redis.go      # Redis 仓储实现（可选）
│   ├── interface/             # 接口层
│   │   └── routes.go          # 路由处理器（薄层，解析 proto → 构建 Command → 投递）
├── svc/                       # 对外接口和跨模块 Service
│   ├── interface.go           # IPlayer + Player DTO
│   └── server/                # 跨模块 Service 实现
│       └── server.go          # 实现 svc.IPlayer 接口
├── rpc/                       # RPC 适配（如有）
│   ├── player.proto           # RPC proto 定义（如有）
│   ├── client/                # RPC 客户端
│   │   └── client.go
│   ├── server/                # RPC 服务端
│   │   └── server.go
│   └── grpc/                  # proto 生成代码
│       ├── player.pb.go
│       └── player_grpc.pb.go
├── module.go                  # Module 构造函数，装配依赖注入
└── option.go                  # 函数式选项
```

### 3.3 依赖方向

```
Interface ──→ Application ──→ Domain ←── Infrastructure
     ↓              ↓            ↑            ↑
  proto 消息    命令/DTO     聚合/VO/事件   仓储实现
```

- **Domain 不依赖任何外层**：不 import proto、不 import due/network
- **Application 依赖 Domain**：不依赖 Infrastructure（通过接口注入）
- **Interface 依赖 Application**：不做业务逻辑
- **Infrastructure 依赖 Domain**：实现仓储接口

---

## 4. Actor 与聚合的关系

这是 game-stack DDD 设计中最关键的架构决策。

### 两个正交边界

| 维度 | Actor | Aggregate |
|------|-------|-----------|
| 边界性质 | **物理**串行化边界 | **逻辑**一致性边界 |
| 关注点 | 并发安全（goroutine 隔离） | 业务规则（不变量保护） |
| 保护范围 | 同一玩家的所有操作 | 单个聚合的属性 |
| 生命周期 | 登录创建 → 断线杀死 | 业务操作创建 → 业务操作删除 |

### 协作模型

```
同一个 Actor goroutine 内：
  │
  ├── CommandBus.Dispatch(ctx, AddExpCmd{...})
  │     │
  │     ├── Player.Load(id)           ← 仓储加载 Player 聚合
  │     ├── Player.AddExp(amount)     ← 聚合内部修改 + 校验不变量
  │     ├── EventBus.Publish(...)     ← 发布领域事件（同步）
  │     └── Player.Save()             ← 仓储保存聚合
  │
  ├── CommandBus.Dispatch(ctx, AddGoldCmd{...})
  │     │
  │     ├── Player.Load(id)           ← 同一聚合，同一 Actor 保证串行
  │     └── Player.AddGold(amount)    ← 不变量校验：Gold ≥ 0
  │
  └── ... (同一玩家的所有命令串行执行)
```

### 跨模块同步调用

其他模块通过 `InvokePlayerSync` 将命令投递到玩家 Actor 中同步执行并获取返回值：

```go
// auth 模块注册时调用 player 模块
newExp, err := actor.InvokePlayerSync[int64](ctx, proxy, uid, func(ctx context.Context) (int64, error) {
    return ddd.Dispatch[int64](ctx, cmdBus, application.AddExpCmd{...})
})
```

**关键规则：**

1. **Actor 不感知 Aggregate**：Actor 只需知道玩家的命令队列，不需要知道有多少聚合
2. **Aggregate 不依赖 Actor**：聚合是纯领域对象，可独立单测
3. **跨聚合操作通过 EventBus**：同一 Actor 内的跨聚合通信走 DomainEvent
4. **命令处理在 Actor 内**：`RouteToActor` → Actor dispatch → CommandBus.Dispatch → CommandHandler.Handle
5. **跨模块写操作走 Actor**：通过 `InvokePlayerSync`/`InvokePlayer` 在玩家 Actor 中串行执行

---

## 5. 开发规范

### 5.1 模块文件清单

| 层级 | 文件 | 职责 |
|------|------|------|
| Domain | `internal/domain/aggregate.go` | 聚合根结构体 + 行为方法（AddExp、DeductGold 等） |
| Domain | `internal/domain/value_objects.go` | 值对象定义 + 校验 + Equals |
| Domain | `internal/domain/events.go` | 领域事件结构体定义 |
| Domain | `internal/domain/repository.go` | 仓储接口（内嵌 `ddd.Repository[T]`） |
| Domain | `internal/domain/service.go` | 领域服务（跨聚合计算逻辑，无状态） |
| Application | `internal/application/commands.go` | 命令结构体定义 |
| Application | `internal/application/handlers.go` | CommandHandler[C, T] 实现（编排层） |
| Application | `internal/application/dto.go` | proto ↔ 领域对象转换函数 |
| Infrastructure | `internal/infrastructure/repo_memory.go` | 内存仓储实现 |
| Infrastructure | `internal/infrastructure/repo_redis.go` | Redis 仓储实现（可选） |
| Interface | `internal/interface/routes.go` | 路由处理器（薄层） |
| RPC Client | `rpc/client/client.go` | RPC 客户端 |
| RPC Server | `rpc/server/server.go` | RPC 服务端 |
| Service | `svc/server/server.go` | 跨模块 Service 实现（Actor 串行化） |
| Public API | `svc/interface.go` | 对外接口定义 + DTO |
| Root | `module.go` | 模块构造函数 + Init（依赖注入装配） |

### 5.2 命名规范

| 元素 | 命名规则 | 示例 |
|------|----------|------|
| 聚合根 | 名词，首字母大写 | `Player`、`Session`、`Backpack` |
| 值对象 | 名词，首字母大写 | `PlayerID`、`Gold`、`Nickname` |
| 领域事件 | 名词 + 过去式动词 | `PlayerLeveledUp`、`GoldDeducted` |
| 命令 | 动词 + 名词 | `AddExpCmd`、`UpdateProfileCmd` |
| 命令处理器 | 命令 + Handler | `AddExpHandler`、`UpdateProfileHandler` |
| 仓储接口 | 聚合名 + Repository | `PlayerRepository`、`SessionRepository` |
| 仓储实现 | 前缀 + Repo | `playerMemoryRepo`、`playerRedisRepo` |
| 领域服务 | 功能描述 + Service | `CalcLevelService` |

### 5.3 添加新模块流程

1. **战略设计** — 确定模块属于核心域/支撑域/通用域
2. **聚合设计** — 识别聚合根、实体、值对象，定义不变量
3. **创建目录结构** — `internal/domain/`、`internal/application/`、`internal/infrastructure/`、`internal/interface/`
4. **对外接口** — `svc/interface.go`（接口 + DTO）
5. **domain 层** — `value_objects.go` → `aggregate.go` → `events.go` → `repository.go`
6. **infrastructure 层** — `repo_memory.go`（至少提供内存实现）
7. **application 层** — `commands.go` → `handlers.go` → `dto.go`
8. **interface 层** — `routes.go`（路由注册 + 薄 handler）
9. **svc/server** — 实现 `svc/` 中的接口
10. **rpc/client + rpc/server** — RPC 客户端和服务端（如有）
11. **module.go** — 装配依赖注入
12. **注册到 cmd/hall/main.go** — `stack.WithModules(xxx.Module())`
13. **proto 定义** — 客户端协议放在 `proto/`，模块 RPC 协议放在 `module/<name>/rpc/<name>.proto`

---

## 参考

- `ddd/` — DDD 核心抽象实现
- `docs/模块开发规范.md` — DDD 四层架构开发规范
- `docs/store设计.md` — 仓储接口设计、实现约束
- `docs/用户数据并发修改安全设计.md` — Actor 串行化模型（DDD 中保持不变）
