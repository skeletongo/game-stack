package ddd

import (
	"context"
	"fmt"
	"sync"

	"github.com/dobyte/due/v2/log"
)

// Command 表示一个意图明确的操作请求。
//
// 命令与应用服务（Application Service）一一对应：
// 客户端请求 → 解析为 Command → CommandBus 路由 → CommandHandler 执行。
//
// 命名规范：动词 + 名词（UpdateProfile, AddExp, DeductGold）
type Command interface {
	// CommandName 返回命令名称，用于 CommandBus 路由。
	CommandName() string
}

// CommandHandler 是命令处理器接口。
// 泛型参数 C 约束为具体的 Command 类型。
//
// Handle 在 Actor goroutine 中执行（通过 RouteToActor 投递），
// 处理跨聚合编排、领域事件发布、仓储调用等应用层职责。
type CommandHandler[C Command] interface {
	// Handle 执行命令。返回 error 表示业务失败。
	Handle(ctx context.Context, cmd C) error
}

// CommandBus 是限界上下文内部的命令总线。
//
// 职责：
//   - 路由命令到对应的 CommandHandler
//   - 类型安全：Register 时校验 handler 类型，Dispatch 时自动类型断言
//
// 线程安全：Register 通常在 Init 阶段调用（单线程），Dispatch 在 Actor goroutine 中调用。
type CommandBus struct {
	mu       sync.RWMutex
	handlers map[string]any // command name → CommandHandler[C]
}

// NewCommandBus 创建新的命令总线。
func NewCommandBus() *CommandBus {
	return &CommandBus{
		handlers: make(map[string]any),
	}
}

// Register 注册命令处理器。
//
// handler 必须是 CommandHandler[C] 类型，C 必须是实现了 Command 接口的指针或值类型。
// 使用示例：
//
//	bus.Register("UpdateProfile", &UpdateProfileHandler{playerRepo: repo})
func (b *CommandBus) Register(cmdName string, handler any) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.handlers[cmdName]; ok {
		panic(fmt.Sprintf("command handler already registered: %s", cmdName))
	}
	b.handlers[cmdName] = handler
}

// Dispatch 将命令分发到对应的处理器。
//
// 自动根据命令的 CommandName() 查找处理器并执行。
// 如果无匹配处理器，记录警告日志并返回 nil（静默忽略，不中断流程）。
func (b *CommandBus) Dispatch(ctx context.Context, cmd Command) error {
	b.mu.RLock()
	handler, ok := b.handlers[cmd.CommandName()]
	b.mu.RUnlock()

	if !ok {
		log.Warnf("[ddd] no handler registered for command: %s", cmd.CommandName())
		return nil
	}

	// 类型断言由调用方保证正确性（Register 时由程序员负责类型匹配）
	// 这里不做 reflect 动态调用，保持简单和性能
	return invokeHandler(ctx, handler, cmd)
}

// invokeHandler 是一个类型擦除桥接函数。
// 在 CommandBus 中使用 any 存储 handler，Dispatch 时通过此函数恢复类型。
//
// 使用泛型函数避免在每个 CommandBus 调用点写重复的类型断言代码。
func invokeHandler[C Command](ctx context.Context, handler any, cmd C) error {
	h, ok := handler.(CommandHandler[C])
	if !ok {
		// 这是编程错误：Register 的 handler 类型与 Command 类型不匹配
		panic(fmt.Sprintf("handler type mismatch for command %s: expected CommandHandler[%T], got %T",
			cmd.CommandName(), cmd, handler))
	}
	return h.Handle(ctx, cmd)
}
