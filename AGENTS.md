# AGENTS.md

本文件为 AI Agent 在此仓库中工作时提供指导。

代码注释必须使用中文  
绝对不要删跟本次任务无关的注释  
新增代码必须加注释，简要说明方法或字段含义  
请求响应结构体每个字段必须要加中文注释  
proto文件每个字段和rpc方法必须加中文注释  
git提交遵守格式规范，但描述内容必须用中文  
不确定的逻辑主动提问  

## 项目概述

`game-stack` 是一个分布式游戏服务器框架，基于 **due v2.5.8**（`github.com/dobyte/due/v2`）构建。采用 DDD（领域驱动设计）分层架构，将 due 的基础组件封装为可插拔的模块系统。

## 构建与验证

```bash
bash update_due.sh          # 获取/重置依赖
bash gen_proto.sh           # 生成所有模块的 proto 代码
go build ./...              # 构建所有包
go vet ./...                # 检查所有包
docker-compose -f docker/docker-compose.yaml up -d  # 启动开发基础设施（etcd + Redis）
```

## 架构

```
ddd/         → DDD 核心抽象（Aggregate、ValueObject、Repository[T]、CommandBus、EventBus、Snapshot、NoResult）
docs/        → 详细设计文档
cmd/         → 入口点（gate、node、gsc CLI）
stack/       → 核心框架 —— Application、Module 接口、路由、错误、中间件、服务注册
  debug/     → Debug HTTP 服务（运行时查询/修改聚合数据）
module/      → 可插拔游戏模块，DDD 四层架构（internal/domain / application / infrastructure / interface）
  actor/     → Actor 串行化引擎（InvokePlayer + InvokePlayerSync）
proto/       → 客户端通信协议（protobuf，客户端与服务端共用）
example/     → 示例项目（gate + server + test client）
```

## 模块开发（DDD 四层架构）

每个模块按 4 层组织，四层代码放在 `internal/` 子目录（Go 编译器边界），对外接口放在 `svc/`，RPC 适配放在顶层 `rpc/`。以 `module/player/` 或 `module/auth/` 为规范示例：

```
module/<name>/
├── internal/
│   ├── domain/            # 领域层（核心，不依赖任何外层）
│   │   ├── aggregate.go   # 聚合根 + 行为方法
│   │   ├── value_objects.go # 值对象定义 + 校验
│   │   ├── events.go      # 领域事件定义
│   │   ├── repository.go  # 仓储接口（内嵌 ddd.Repository[T]）
│   │   └── service.go     # 领域服务（纯函数）
│   ├── application/       # 应用层（编排，不包含领域规则）
│   │   ├── commands.go    # 命令定义
│   │   ├── handlers.go    # CommandHandler[C, T] 实现
│   │   └── dto.go         # proto ↔ 领域对象转换
│   ├── infrastructure/    # 基础设施层
│   │   └── repo_memory.go # 内存仓储实现
│   ├── interface/         # 接口层（最薄层）
│   │   └── routes.go      # 路由处理器
├── svc/                   # 对外接口和跨模块 Service
│   ├── interface.go       # IXxx + DTO
│   └── server/            # Service 实现
│       └── server.go      # 实现 svc.IXxx 接口
├── rpc/                   # RPC 适配（如有）
│   ├── <name>.proto       # RPC proto 定义（如有）
│   ├── client/            # RPC 客户端
│   │   └── client.go
│   ├── server/            # RPC 服务端
│   │   └── server.go
│   └── grpc/              # proto 生成代码
│       ├── <name>.pb.go
│       └── <name>_grpc.pb.go
├── module.go              # Module 构造函数 + 依赖注入装配
└── option.go              # 函数式选项（WithRepository）
```

详见 `docs/模块开发规范.md`、`docs/DDD设计文档.md`。

## 关键 DDD 抽象（ddd/）

| 接口 | 说明 |
|------|------|
| `ddd.Aggregate` | 聚合根，拥有业务不变量 |
| `ddd.ValueObject` | 无标识的值对象，按值比较 |
| `ddd.Scalar` | 可选接口，VO 实现后自动展开为 JSON 原始类型 |
| `ddd.Repository[T]` | 泛型仓储接口（Load/Save/Delete） |
| `ddd.CommandBus` | 命令总线，`ddd.Register(cmdBus, name, handler)` 注册，`ddd.Dispatch[T](ctx, bus, cmd)` 类型安全分发 |
| `ddd.NoResult` | 无返回值命令的占位结果类型 |
| `ddd.EventBus` | 同步领域事件总线（BC 内部使用） |
| `ddd.Snapshot(agg)` | 反射读取聚合所有非导出字段 |
| `ddd.ApplyPatch(agg, m)` | unsafe 直接修改聚合字段（仅调试用） |

## Context 使用约定

| 层 | Context 类型 | 原因 |
|---|---|---|
| Interface（routes.go） | `node.Context` | 框架依赖：Parse、Response、UID、BindGate |
| Application（handlers.go） | `context.Context` | 框架无关，通过 `ctx.Context()` 获取 |
| Infrastructure | `context.Context` | 标准库接口 |

传参时用 `ctx.Context()` 而非 `context.Background()`，以传递框架的超时和追踪。

## 响应模式（proto codec）

proto codec 下通过 `stack.ProtoResponse(ctx, msg)` 发送 proto 消息。

每个 Response 前两个字段固定为 `code`(int32) + `message`(string)。成功时 `Code: stack.CodeOK`，错误时通过 `stack.ErrCode(err)` 从 `*stack.Code` 解包错误码：

```go
// 成功
stack.ProtoResponse(ctx, &auth.LoginResponse{Code: stack.CodeOK, Token: t, PlayerId: uid})

// 错误
log.Errorf("xxx failed: %v", err)
stack.ProtoResponse(ctx, &auth.LoginResponse{Code: stack.ErrCode(err), Message: err.Error()})
```

`stack.Response` 封装仅适用于 JSON codec（`stack.JSONRespond*` 系列函数）。proto codec 下该结构体不实现 `proto.Message` 会导致序列化失败。

## 路由编号

公式：**模块号 × 1000 + 子协议号**。0–999 系统预留。常量在 `stack/route.go`。

## 错误码

系统错误（0–999）复用 HTTP 语义，业务错误 `模块号 × 1000 + 子码`。定义在 `stack/errcode.go`。

## 关键设计决策

- **Actor 与聚合正交**：Actor 提供物理串行化，Aggregate 提供逻辑不变量
- **RouteToActor** 不检查归属权：due 的 StatefulRoute 已保证消息投递到正确节点
- **InvokePlayer** 检查归属权，fire-and-forget 异步执行
- **InvokePlayerSync** 同步执行并返回结果，用于跨模块 Service 调用
- **RPC 放在顶层 rpc 子包**：`module/player/rpc/server/` 注册服务端，`module/player/rpc/client/` 创建客户端
- **Debug 默认关闭**：`stack.WithDebug(":6060")` 启用，模块一行 `debug.Register[*Player]("player", repo, cmdBus)` 注册
- **玩家节点归属稳定**：普通断线不解绑节点；已有节点归属只由显式玩家迁移流程修改

## 参考文档

- `docs/due-api.md` — Due v2.5.8 API 参考
- `docs/DDD设计文档.md` — DDD 战略设计（子域/BC/上下文映射）+ 战术设计 + Actor 与聚合关系
- `docs/模块开发规范.md` — DDD 四层架构 + module.go 模板 + 路由模式 + 添加新模块流程
- `docs/store设计.md` — 仓储接口设计、实现约束（删除幂等、查询重加载）
- `docs/debug调试服务设计.md` — Debug HTTP 服务设计
- `docs/客户端服务端通信协议设计.md` — 路由编号方案、错误码体系
- `docs/用户绑定节点设计.md` — 玩家节点绑定、有状态/无状态路由、断线重连
- `docs/用户数据并发修改安全设计.md` — Actor 串行化、RouteToActor/InvokePlayer、归属权校验
- `docs/go语言规范.md` — Go 编码规范
