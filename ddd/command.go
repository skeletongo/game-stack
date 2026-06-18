package ddd

import (
	"context"
	"fmt"
	"sync"

	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/internal/logfmt"
)

// NoResult 是命令无返回值时使用的占位结果类型。
type NoResult struct{}

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
// 泛型参数 C 约束为具体的 Command 类型，T 为命令执行的结果类型。
//
// Handle 在 Actor goroutine 中执行（通过 RouteToActor 投递），
// 处理跨聚合编排、领域事件发布、仓储调用等应用层职责。
type CommandHandler[C Command, T any] interface {
	// Handle 执行命令。返回 (T, error)，T 为命令执行结果，error 表示业务失败。
	Handle(ctx context.Context, cmd C) (T, error)
}

// cmdEntry 存储命令工厂和调用闭包。
type cmdEntry struct {
	newCmd func() any
	invoke func(context.Context, Command) (any, error)
}

// CommandBus 是限界上下文内部的命令总线。
//
// 职责：
//   - 路由命令到对应的 CommandHandler
//   - 类型安全：Register 时校验 handler 类型，Dispatch 时自动类型断言
//
// 线程安全：Register 通常在 Init 阶段调用（单线程），Dispatch 在 Actor goroutine 中调用。
type CommandBus struct {
	mu      sync.RWMutex
	entries map[string]*cmdEntry // command name → entry
}

// NewCommandBus 创建新的命令总线。
func NewCommandBus() *CommandBus {
	return &CommandBus{
		entries: make(map[string]*cmdEntry),
	}
}

// Register 注册命令处理器。
//
// C 由 handler 参数自动推导。同时存入命令工厂函数，供 debug 服务反序列化参数。
//
// 使用示例：
//
//	ddd.Register(cmdBus, "UpdateProfile", &UpdateProfileHandler{...})
func Register[C Command, T any](bus *CommandBus, cmdName string, handler CommandHandler[C, T]) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	if _, ok := bus.entries[cmdName]; ok {
		panic(fmt.Sprintf("command handler already registered: %s", cmdName))
	}
	bus.entries[cmdName] = &cmdEntry{
		newCmd: func() any { return new(C) },
		invoke: func(ctx context.Context, cmd Command) (any, error) {
			c, ok := cmd.(C)
			if !ok {
				return nil, fmt.Errorf("command type mismatch")
			}
			log.Debugf("==> cmd=%s type=%T payload=%s", cmd.CommandName(), cmd, logfmt.JSON(cmd))
			return handler.Handle(ctx, c)
		},
	}
}

// Dispatch 将命令分发到对应的处理器。
//
// 自动根据命令的 CommandName() 查找处理器并执行。
// 如果无匹配处理器，记录警告日志并返回 nil（静默忽略，不中断流程）。
func (b *CommandBus) Dispatch(ctx context.Context, cmd Command) (any, error) {
	b.mu.RLock()
	entry, ok := b.entries[cmd.CommandName()]
	b.mu.RUnlock()

	if !ok {
		log.Warnf("[ddd] no handler registered for command: %s", cmd.CommandName())
		return nil, nil
	}

	return entry.invoke(ctx, cmd)
}

// Dispatch 是类型安全的命令分发辅助函数。
//
// 通过泛型推导返回类型，调用方可直接获取具体类型结果：
//
//	result, err := ddd.Dispatch[*LoginResult](ctx, cmdBus, cmd)
func Dispatch[T any](ctx context.Context, bus *CommandBus, cmd Command) (T, error) {
	var zero T
	ret, err := bus.Dispatch(ctx, cmd)
	if err != nil {
		return zero, err
	}
	if ret == nil {
		return zero, nil
	}
	result, ok := ret.(T)
	if !ok {
		return zero, fmt.Errorf("command result type mismatch: expected %T, got %T", zero, ret)
	}
	return result, nil
}

// NewCommand 返回指定命令的零值指针（类型为 *Cmd）。不存在时返回 false。
// 返回值可直接传给 json.Unmarshal 填充参数，再断言为 Command 传给 Dispatch。
func (b *CommandBus) NewCommand(name string) (any, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	entry, ok := b.entries[name]
	if !ok {
		return nil, false
	}
	return entry.newCmd(), true
}

// CommandNames 返回所有已注册命令名称。
func (b *CommandBus) CommandNames() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	names := make([]string, 0, len(b.entries))
	for name := range b.entries {
		names = append(names, name)
	}
	return names
}
