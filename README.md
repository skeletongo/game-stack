# game-stack

基于 [due](https://github.com/dobyte/due) v2.5.5 封装的分布式游戏服务端框架，开箱即用。

## 快速开始

```bash
# 获取依赖
bash update_due.sh

# 启动基础设施（etcd + Redis）
docker compose -f docker/docker-compose.yaml up -d

# 启动网关（WebSocket）
go run cmd/gate/main.go

# 启动逻辑服
go run cmd/node/main.go
```

## 设计要点

- **模块化架构** — 14 个游戏模块（auth / player / chat / match / room / inventory / quest / combat / guild / mail / shop / leaderboard / activity / social），遵循统一的 6 文件模式，可插拔
- **Actor 串行化** — 同一玩家的所有状态修改在单 goroutine 中排队执行，杜绝读-改-写并发竞争
- **断线无缝重连** — 30s Grace Period 内重连保留内存数据，避免频繁清理重建
- **编号统一** — 路由和错误码共用 `模块号 × 1000 + 子码` 公式，心智负担低
- **Store 抽象** — 接口隔离存储实现，开发用内存，生产切换 Redis/MySQL/MongoDB

## 架构

```
cmd/        入口点（gate、node）
stack/      核心框架：应用启动、路由、错误码、Module 接口、中间件、延迟清理器
module/     14 个可插拔游戏模块
protocol/   消息结构体（Go 结构体 + json/msgpack 标签，无需 protoc）
docs/       设计文档
```

## 后续计划

### 并发安全

- [ ] Actor 写操作全覆盖（当前仅 ShopBuy 已接入 RouteToActor，其余 9 个模块待改造）

### 模块完善

- [ ] Token 共享存储（Redis 实现，支持跨节点令牌验证）
- [ ] 其余模块接入 CleanableService（shop / guild / leaderboard / chat / mail / room / match）

### 基础设施

- [ ] 生产环境 Store 实现（Redis / MySQL / MongoDB）
- [ ] 测试覆盖

### 功能模块

- [ ] 邮件系统完整实现（附件、批量操作、过期清理）
- [ ] 排行榜分段加载、赛季重置
- [ ] 活动系统时间线管理、条件任务

### 文档

- [ ] `go语言规范` 补充（命名规范、代码组织、并发）
- [ ] 各模块 Store 接口文档
