# Hall 玩家迁移设计

## 目标

Hall 玩家迁移用于支持大厅逻辑服滚动更新、节点缩容、故障摘除和玩家负载重均衡。迁移的核心目标不是“把 Locator 改到新节点”这么简单，而是保证同一玩家在任何时刻只有一个 Hall 拥有写入权，并且玩家状态不会因为新旧 Hall 并存而丢失、覆盖或重复执行。

本文档定义玩家从旧 Hall 迁移到新 Hall 的统一协议。后续所有与 Hall 更新、玩家迁移、玩家状态归属相关的实现都必须以本文档为准。

## 术语

| 名称 | 说明 |
|------|------|
| 源 Hall | 当前拥有玩家节点归属的旧 Hall 实例 |
| 目标 Hall | 准备接管玩家归属的新 Hall 实例 |
| 玩家归属 | Locator 中记录的玩家 UID → Hall 实例 ID 映射 |
| 写入权 | 某个 Hall 对玩家状态执行写操作并保存的资格 |
| Ownership Epoch | 玩家归属版本号，每次迁移成功时递增，用于防止旧 Hall 继续写入 |
| Drain | 源 Hall 停止接收新写操作，并等待 Actor 中已排队任务执行到迁移屏障 |
| Snapshot | 从源 Hall 导出的玩家状态快照，主要用于内存仓储或缓存恢复 |
| Flush | 将源 Hall 中玩家状态持久化到共享仓储 |
| Fence | 写入栅栏，仓储只接受当前归属 epoch 的写入 |

## 核心原则

1. **Locator 切换不是迁移的全部**：Locator 只决定新消息投递到哪里，不负责玩家状态一致性。
2. **同一玩家同一时刻只能有一个写入者**：源 Hall 和目标 Hall 不能同时修改同一玩家状态。
3. **迁移必须经过玩家 Actor**：迁移命令必须投递到源玩家 Actor，由 Actor 串行化地执行迁移屏障。
4. **仓储必须提供最终防线**：即使旧 Hall 收到延迟消息，仓储层也要通过 ownership epoch 或版本号拒绝旧写入。
5. **状态权威来源必须明确**：共享持久仓储优先；内存仓储只能通过 snapshot/restore 支持开发级迁移。
6. **失败必须可恢复**：迁移流程每个阶段都要能重试、回滚或由后台任务修复。
7. **普通断线不是迁移**：断线、登出、token 失效不改变玩家节点归属。

## 当前基础

现有框架已经有三个关键基础：

- `StatefulRoute`：Gate 根据 Locator 将玩家有状态消息投递到绑定 Hall。
- `PlayerActor`：同一玩家的写操作通过 Actor 串行执行。
- `InvokePlayer` / `InvokePlayerSync`：跨模块或 RPC 写操作投递到玩家 Actor，并检查当前节点归属。

这些能力解决了“同一 Hall 内的串行化”和“普通请求路由到归属 Hall”，但还不足以解决热迁移。热迁移还需要额外能力：

- 源 Actor 进入迁移状态后拒绝或延迟新写入。
- 迁移屏障等待已排队写操作完成。
- Locator 切换和 ownership epoch 更新原子化。
- 仓储 Save 通过 epoch 或版本号做 fencing。
- 目标 Hall 恢复状态后才能开放写入。

## 状态是否需要移动

状态是否需要从旧 Hall 移到新 Hall，取决于基础设施层的仓储实现。

| 仓储类型 | 是否需要移动状态 | 迁移方式 | 说明 |
|----------|------------------|----------|------|
| 纯内存仓储 | 需要 | Snapshot → Restore | 当前 `MemoryRepo` 属于这种类型，新 Hall 无法直接读取旧 Hall 内存 |
| 共享持久仓储 | 不需要搬对象 | Drain → Flush → 新 Hall Load | Redis/MySQL/MongoDB 等共享仓储是生产推荐方案 |
| 本地缓存 + 共享持久仓储 | 需要处理缓存 | Flush + Invalidate/Reload | 源 Hall 缓存必须失效，目标 Hall 必须重载或恢复 |
| 分片持久仓储 | 视分片策略而定 | Flush + 迁移分片或访问路由 | 如果玩家数据分片也跟 Hall 绑定，需要独立数据迁移设计 |

生产环境推荐使用共享持久仓储，Hall 迁移时不搬 Go 对象，只保证源 Hall 已经把最新状态落到仓储，目标 Hall 再从仓储加载。

内存仓储只能作为开发或测试实现。如果要支持内存仓储迁移，必须让每个玩家模块实现 snapshot/restore 钩子；否则滚动更新时内存状态会丢失。

## 写一致性模型

### 为什么 Actor 不够

Actor 保证同一节点内同一玩家串行执行，但迁移时可能出现两个 Hall 都持有玩家 Actor：

```
旧 Hall Actor 还没销毁
    │
    ├── 延迟的 StatefulRoute 消息仍到达旧 Hall
    ├── 某个 gRPC 回调仍在旧 Hall 投递 InvokePlayer
    └── 目标 Hall 已经创建新 Actor 并开始接收新消息
```

如果没有额外防护，新旧 Hall 都可能执行写操作。即使每个 Hall 内部都是串行的，跨 Hall 仍然会产生并发写覆盖。

### Ownership Epoch

每个玩家维护一个 ownership 记录：

```text
player_owner:{uid} = {
  hall_name: "game-hall",
  node_id:   "hall-xxx",
  epoch:     42,
  state:     "active|migrating",
  updated_at: 1710000000
}
```

规则：

1. 玩家 Actor 创建时读取当前 owner record，保存 `node_id` 和 `epoch`。
2. 每次玩家状态写入仓储时带上 `epoch`。
3. 仓储只接受 `epoch == 当前 owner epoch` 的写入。
4. 迁移成功提交时，owner record 的 `node_id` 切到目标 Hall，`epoch` 加 1。
5. 旧 Hall 任何迟到写入都携带旧 epoch，必须被仓储拒绝。

这就是 fencing。它是防止旧写入污染新状态的最后一道防线。

## 迁移状态机

```
Active@源Hall
    │
    │  迁移管理器发起迁移(uid, target)
    ▼
Draining@源Hall
    │
    │  迁移命令进入源 Actor mailbox
    │  等待屏障前所有写操作完成
    │  源 Actor 标记 migrating，拒绝新写
    ▼
Flush/Snapshot@源Hall
    │
    │  共享仓储：保存最新状态
    │  内存仓储：导出所有玩家模块快照
    ▼
Prepare@目标Hall
    │
    │  目标 Hall 创建迁移占位
    │  共享仓储：暂不对外写，等待 commit 后 Load
    │  内存仓储：导入 snapshot，但仍禁止写
    ▼
Commit
    │
    │  原子更新 owner record:
    │    node_id = 目标Hall
    │    epoch = epoch + 1
    │    state = active
    │  同步更新 due Locator
    ▼
Activate@目标Hall
    │
    │  目标 Actor 使用新 epoch 开放写入
    │  后续 StatefulRoute 投递到目标 Hall
    ▼
Cleanup@源Hall
    │
    │  源 Actor 校验 owner 已失去归属
    │  清理本地 Actor、缓存、迁移状态
    ▼
Active@目标Hall
```

## 迁移流程

### 1. 选择目标 Hall

迁移管理器选择目标 Hall 时必须确认：

- 目标 Hall 已注册到服务发现。
- 目标 Hall 版本满足迁移兼容要求。
- 目标 Hall 可用且负载允许接收新玩家。
- 目标 Hall 能访问玩家所需的共享仓储或支持 snapshot restore。

如果目标 Hall 版本和源 Hall 的领域模型不兼容，必须先做数据兼容层或禁止在线迁移。

### 2. 在源 Actor 建立迁移屏障

迁移请求必须投递给源玩家 Actor，而不是直接由迁移管理器修改状态。

```
迁移管理器 → 源 Hall → actor.InvokePlayerSync(uid, BeginMigrate)
```

Actor 收到 `BeginMigrate` 时：

1. 由于 Actor mailbox 串行，`BeginMigrate` 前面的写操作已经执行完。
2. Actor 将玩家迁移状态设为 `migrating`。
3. Actor 后续收到普通写操作时返回迁移中错误或短暂排队。
4. Actor 禁止再启动新的跨模块玩家写事务。

建议优先“拒绝并让客户端重试”，而不是无限排队。迁移通常很短，客户端收到迁移中错误后可以退避重试。

### 3. Flush 或 Snapshot

共享持久仓储：

1. 源 Actor 将所有玩家模块状态保存到仓储。
2. 确认仓储版本号或 epoch 已写入。
3. 源 Hall 不再接受该玩家写操作。

内存仓储：

1. 源 Actor 调用每个玩家模块的 `Snapshot(uid)`。
2. 生成包含所有模块状态和版本的迁移包。
3. 迁移包发送给目标 Hall。

迁移包必须包含：

| 字段 | 说明 |
|------|------|
| `uid` | 玩家 ID |
| `source_node_id` | 源 Hall 实例 ID |
| `target_node_id` | 目标 Hall 实例 ID |
| `epoch` | 源 Hall 当前持有的 epoch |
| `modules` | 各玩家模块快照 |
| `created_at` | 迁移包创建时间 |
| `schema_version` | 快照结构版本 |

### 4. 目标 Hall 预恢复

目标 Hall 接收迁移请求后不能立即对外提供写服务。

目标 Hall 需要做：

- 创建迁移占位状态，标记 `uid` 正在导入。
- 如果是内存仓储，执行各模块 `Restore(uid, snapshot)`。
- 如果是共享仓储，等待 commit 后按需 Load，或提前预热但禁止写。
- 准备目标 Actor，但 Actor 必须处于 `pending` 状态。

此时 Locator 仍然指向源 Hall。目标 Hall 即使收到了误投递请求，也必须拒绝写入。

### 5. 原子 Commit

Commit 是迁移的关键点。必须原子完成：

1. 校验当前 owner 仍然是源 Hall。
2. 校验当前 epoch 等于迁移开始时读取的 epoch。
3. 将 owner node_id 更新为目标 Hall。
4. 将 epoch 加 1。
5. 将 state 改为 active。
6. 更新 due Locator 中玩家绑定节点。

如果使用 Redis，建议用 Lua 脚本完成 owner record 和 Locator 的原子更新。不能先改 Locator 再改 owner，也不能先改 owner 再改 Locator 后忽略失败。

Commit 失败时：

- 如果 owner 已不是源 Hall，说明被其他迁移或恢复流程接管，本次迁移必须中止。
- 如果目标 Hall 准备失败，源 Hall 可以恢复 active。
- 如果 Commit 状态不确定，必须通过读取 owner record 判断最终归属，不能盲目重试写 Locator。

### 6. 目标 Activate

Commit 成功后：

1. 目标 Hall 读取新的 owner record。
2. 目标 Actor 绑定新 epoch。
3. 目标 Actor 状态从 pending 改为 active。
4. 目标 Hall 开始接受该玩家写操作。

如果目标 Hall 在 Commit 后崩溃，恢复流程必须能根据 owner record 发现玩家归属已经切到目标 Hall，并在目标 Hall 重启或替代节点接管时重新加载状态。

### 7. 源 Cleanup

源 Hall 在确认 owner 已经不是自己后：

1. 销毁源玩家 Actor。
2. 清理本地缓存和 idle timer。
3. 清理迁移临时状态。
4. 忽略或拒绝该玩家后续迟到请求。

源 Hall Cleanup 失败不应该影响新 Hall 对外服务。旧 Actor 即使残留，也会因为 ownership epoch 失效无法成功写入。

## 模块迁移钩子

Hall 迁移是玩家级能力，但玩家状态通常分布在多个模块，例如 player、bag、quest、mail。迁移管理器不能只迁移 player 聚合。

建议定义玩家状态迁移钩子：

```go
// PlayerMigrationHook 定义模块参与玩家迁移的钩子。
type PlayerMigrationHook interface {
    // PrepareMigration 在源 Actor 中执行，阻止模块开启新的玩家写事务。
    PrepareMigration(ctx context.Context, uid int64) error

    // FlushMigration 在源 Actor 中执行，将模块状态持久化到共享仓储。
    FlushMigration(ctx context.Context, uid int64, epoch int64) error

    // SnapshotMigration 在源 Actor 中执行，为内存仓储导出模块状态。
    SnapshotMigration(ctx context.Context, uid int64) (any, error)

    // RestoreMigration 在目标 Hall 中执行，为内存仓储恢复模块状态。
    RestoreMigration(ctx context.Context, uid int64, snapshot any, epoch int64) error

    // CleanupMigration 在源 Hall 完成迁移后执行，清理本地缓存。
    CleanupMigration(ctx context.Context, uid int64) error
}
```

生产共享仓储场景下，核心是 `PrepareMigration`、`FlushMigration`、`CleanupMigration`。内存仓储场景才需要 `SnapshotMigration` 和 `RestoreMigration`。

模块钩子必须满足：

- 幂等：重复调用不会破坏状态。
- 有序：同一玩家的钩子在源 Actor 中串行执行。
- 可失败：失败时迁移流程可以中止或重试。
- 不跨玩家阻塞：迁移一个玩家不能阻塞所有玩家。

## 仓储能力要求

生产级玩家仓储必须支持以下能力：

| 能力 | 说明 |
|------|------|
| 按玩家 ID 加载 | 目标 Hall 能重新加载玩家状态 |
| 按玩家 ID 保存 | 源 Hall 能 flush 最新状态 |
| 版本号或 epoch 校验 | 防止旧 Hall 迟到写覆盖新状态 |
| 幂等保存 | 重试不会重复扣除、重复发奖或破坏索引 |
| 事务或批处理 | 同一模块多个聚合需要一致保存时使用 |
| 可观测错误 | Save 失败要能区分冲突、临时错误和数据损坏 |

写入建议形式：

```text
Save(uid, aggregate, expected_epoch, expected_version)
```

保存时必须校验：

- `expected_epoch == owner.epoch`
- `expected_version == current_version` 或使用模块自己的乐观锁版本

如果校验失败，返回明确的 ownership lost 或 version conflict 错误。

## 路由与 Actor 防线

迁移需要多层防护，而不是只依赖其中一层。

### Gate/Locator 层

Commit 后，Gate 应根据 Locator 将新消息投递到目标 Hall。由于 Locator 依赖 Pub/Sub 和本地缓存，短时间内可能存在旧缓存。

因此 Locator 层只能作为路由优化，不能作为唯一一致性保障。

### RouteToActor 层

玩家有状态路由进入 Actor 前应校验当前 Hall 是否仍拥有该玩家：

- 如果 owner 指向本 Hall，允许进入 Actor。
- 如果 owner 指向其他 Hall，拒绝处理并清理本地残留 Actor。
- 如果 owner state 为 migrating，返回迁移中错误。

这可以减少旧 Hall 继续处理迟到消息的概率。

### Actor 层

Actor 内部维护玩家状态：

```text
active      正常处理读写
migrating   拒绝新写，允许迁移钩子执行
pending     目标 Hall 已恢复但未 commit，不允许写
closed      已失去归属，所有写拒绝
```

Actor 是迁移的串行化边界。迁移命令必须排在 Actor mailbox 中执行，不能绕过 Actor 直接改仓储。

### 仓储层

仓储是最终防线。无论路由和 Actor 层是否出现延迟、缓存失效或 bug，旧 epoch 写入都必须失败。

## 读操作策略

读操作分三类：

| 读类型 | 迁移期间策略 |
|--------|--------------|
| Actor 内强一致读 | 迁移中返回重试，或等迁移完成后读目标 Hall |
| 仓储只读查询 | 可读共享仓储，但可能读到迁移前状态，需要业务接受 |
| 客户端状态查询 | 建议走 StatefulRoute；迁移中返回重试，避免读到旧缓存 |

查询虽然不修改状态，但如果读的是本地缓存或内存仓储，也必须受迁移状态约束。

## 失败处理

### 源 Hall 在 Drain 前崩溃

如果源 Hall 崩溃且 owner 仍指向源 Hall：

- 目标 Hall 不能直接接管内存状态，因为状态可能没有 flush。
- 如果使用共享持久仓储，可以由恢复流程判断源 Hall 不可用后重新绑定到可用 Hall，并从仓储加载最后持久状态。
- 如果使用纯内存仓储，未持久状态不可恢复，只能视为开发环境限制。

### 源 Hall Drain 后、Commit 前崩溃

如果源 Hall 已 flush，但尚未 commit：

- owner 仍指向源 Hall，state 可能是 migrating。
- 迁移管理器可以检查迁移事务记录。
- 如果确认 flush 完成，可以继续 commit 到目标 Hall。
- 如果无法确认 flush 完成，应回滚 state 为 active 或等待人工处理。

### 目标 Hall Restore 后、Commit 前崩溃

owner 仍指向源 Hall。迁移可以中止：

- 源 Hall 恢复 active。
- 清理目标 Hall 残留迁移占位。
- 后续可重新选择目标 Hall 发起迁移。

### Commit 成功后、目标 Hall 崩溃

owner 已指向目标 Hall，epoch 已递增：

- 旧 Hall 不能继续写，因为 epoch 已失效。
- 如果目标 Hall 不可用，恢复流程需要再次迁移或重新绑定到其他 Hall。
- 共享仓储可以从最新持久状态恢复。
- 内存仓储如果目标快照未持久化，会有丢失风险，因此不适合生产热更新。

### Cleanup 失败

Cleanup 失败只影响旧 Hall 本地资源。由于 owner 和 epoch 已经切换，旧 Actor 不能成功写入。后台清理任务应定期扫描并销毁失去归属的本地 Actor。

## 迁移管理器

迁移管理器负责调度玩家迁移，可以作为 Hall 内部组件，也可以作为独立管理服务。

职责：

1. 选择目标 Hall。
2. 按批次迁移玩家，控制并发。
3. 记录迁移事务状态。
4. 处理失败重试和超时回滚。
5. 对外暴露滚动更新进度。
6. 在旧 Hall 下线前确认该 Hall 无玩家归属。

迁移管理器不直接修改玩家业务状态。所有业务状态修改必须通过模块迁移钩子和玩家 Actor 完成。

## 滚动更新流程

推荐滚动更新流程：

```
1. 启动新 Hall 版本
2. 新 Hall 注册服务发现，但暂不承接旧玩家
3. 将旧 Hall 标记为 draining
   - 不再绑定首次登录玩家
   - 不再作为迁移目标
   - 已有玩家继续服务
4. 迁移管理器按批次迁移旧 Hall 上的玩家
5. 每批迁移后检查错误率、耗时、目标负载
6. 旧 Hall 玩家归属清零后停止旧 Hall
7. 清理旧 Hall 临时数据和服务发现记录
```

旧 Hall 进入 draining 后，登录流程处理建议：

- 已绑定在旧 Hall 的玩家继续走旧 Hall，直到迁移完成。
- 首次登录或原归属失效的玩家不再绑定到 draining Hall。
- 如果登录请求落到 draining Hall 且玩家无绑定，应该转发、重试或绑定到非 draining Hall。

## 批量迁移策略

玩家迁移不应一次性迁走所有在线玩家。

建议：

- 按玩家在线状态、活跃度、UID 分片分批迁移。
- 单 Hall 同时迁移玩家数设置上限。
- 单玩家迁移设置超时。
- 每批完成后检查失败率和平均耗时。
- 对高活跃玩家使用更小批次。

批量迁移期间必须保留限流，避免目标 Hall 同时加载大量玩家状态导致抖动。

## 客户端体验

迁移期间客户端可能遇到短暂重试：

- 普通写操作返回迁移中错误码。
- 客户端收到迁移中错误后延迟重试。
- 对于必须无感的关键操作，可以由服务端短暂排队，但必须设置超时。

建议迁移窗口内单玩家阻塞时间控制在几十到几百毫秒。超过阈值应中止迁移，恢复源 Hall 服务。

## 监控指标

迁移机制必须具备可观测性。

关键指标：

| 指标 | 说明 |
|------|------|
| `migration_total` | 迁移总次数 |
| `migration_success_total` | 成功次数 |
| `migration_failed_total` | 失败次数 |
| `migration_duration_ms` | 单玩家迁移耗时 |
| `migration_draining_players` | 正在迁移玩家数 |
| `migration_retry_total` | 重试次数 |
| `ownership_conflict_total` | owner/epoch 冲突次数 |
| `stale_write_rejected_total` | 旧 epoch 写入被拒绝次数 |
| `actor_cleanup_total` | 源 Actor 清理次数 |
| `actor_orphan_total` | 发现失去归属的残留 Actor 数 |

关键日志字段：

- `uid`
- `source_node_id`
- `target_node_id`
- `epoch`
- `migration_id`
- `phase`
- `duration_ms`
- `error`

## 安全边界

以下行为禁止：

1. 直接修改 Locator 迁移玩家，不经过源 Actor drain。
2. 目标 Hall 在 commit 前处理玩家写操作。
3. 源 Hall 在 commit 后继续接受玩家写操作。
4. 仓储 Save 不校验 ownership epoch。
5. 跨模块玩家状态只迁移 player 模块，忽略 bag、quest、mail 等模块。
6. 将普通断线、登出、token 失效作为迁移触发条件。
7. 迁移失败后不记录事务状态，导致无法判断最终归属。

## 与现有设计关系

- `docs/用户绑定节点设计.md` 定义普通登录、重连和 Locator 绑定规则。
- `docs/用户数据并发修改安全设计.md` 定义单 Hall 内玩家写操作必须通过 Actor 串行化。
- 本文档定义跨 Hall 迁移时的所有权切换、状态转移和写一致性规则。

三者关系：

```text
用户绑定节点设计
    解决：玩家请求路由到哪个 Hall

用户数据并发修改安全设计
    解决：同一 Hall 内玩家写操作如何串行

Hall 玩家迁移设计
    解决：玩家从旧 Hall 切到新 Hall 时如何不丢状态、不双写
```

## 分阶段实现建议

### 阶段 1：迁移防线

- 引入 owner record 和 ownership epoch。
- Actor 创建时记录 epoch。
- 玩家写入仓储时校验 epoch。
- RouteToActor 前增加归属校验。

### 阶段 2：共享仓储迁移

- 为生产仓储实现 epoch fencing。
- 实现源 Actor drain 和 flush。
- 实现目标 Hall activate。
- 支持单玩家手动迁移。

### 阶段 3：滚动更新

- 实现 Hall draining 状态。
- 实现迁移管理器批量调度。
- 实现迁移事务记录、失败重试和监控。
- 支持旧 Hall 玩家清零后自动下线。

### 阶段 4：内存仓储开发迁移

- 定义模块 snapshot/restore 钩子。
- MemoryRepo 支持导出和导入。
- 用于本地开发、集成测试和迁移流程验证。

生产环境不应依赖纯内存仓储完成热更新。

## 待确认决策

1. owner record 存储在 Redis Locator 同库，还是独立迁移状态库。
2. Commit 是否要求 owner record 和 due Locator 使用同一个 Redis Lua 脚本原子更新。
3. 玩家迁移中客户端写操作是立即返回重试，还是短暂排队等待。
4. 生产仓储首先支持 Redis、MySQL 还是 MongoDB。
5. MemoryRepo 是否只声明不支持生产迁移，还是实现 snapshot/restore 方便测试。
6. 多模块迁移钩子是注册到 `stack`，还是独立 `migration` 包统一管理。
7. 迁移事务记录保留时长和人工修复入口。

