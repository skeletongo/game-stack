package ddd

import (
	"sync"
	"time"

	"github.com/dobyte/due/v2/log"
)

// DomainEvent 表示领域中发生的有意义的事件。
//
// 领域事件用于限界上下文内部的组件间通信（解耦聚合之间的直接依赖）。
// 与 stack.EventBus（跨节点）不同，DomainEvent 在同一个 BC 内同步派发。
//
// 命名规范：过去式（PlayerLeveledUp, GoldDeducted, ItemAcquired）
type DomainEvent interface {
	// AggregateID 返回触发此事件的聚合 ID。
	AggregateID() int64
	// EventName 返回事件名称（过去式动词短语）。
	EventName() string
	// OccurredAt 返回事件发生的时间戳。
	OccurredAt() time.Time
}

// EventHandler 是领域事件的处理器函数类型。
type EventHandler func(event DomainEvent)

// EventBus 是限界上下文内部的同步领域事件总线。
//
// 特性：
//   - 同步派发：所有处理器在当前 goroutine 中依次执行
//   - 类型安全：通过 EventName 字符串路由（建议用常量）
//   - 线程安全：Subscribe 和 Publish 可并发调用
//
// 同步派发是刻意设计：
//   - Actor 已是单 goroutine，异步派发反而增加复杂度
//   - 同步派发保证事件处理完成后才返回客户端响应
//   - 如需异步处理，处理器内部自行启动 goroutine
type EventBus struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
}

// NewEventBus 创建新的事件总线。
func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]EventHandler),
	}
}

// Subscribe 注册事件处理器。可对同一事件注册多个处理器。
func (b *EventBus) Subscribe(eventName string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], handler)
}

// Publish 同步派发领域事件。所有注册的处理器依次执行。
// 某个处理器的 panic 会被 recover，不会中断后续处理器的执行。
func (b *EventBus) Publish(event DomainEvent) {
	log.Debugf("** event=%s type=%T aggregate=%d occurred_at=%s",
		event.EventName(), event, event.AggregateID(), event.OccurredAt().Format(time.RFC3339Nano))
	b.mu.RLock()
	handlers := b.handlers[event.EventName()]
	// 复制一份避免长时间持锁
	h := make([]EventHandler, len(handlers))
	copy(h, handlers)
	b.mu.RUnlock()

	for _, handler := range h {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("[ddd] event handler panic: event=%s aggregate=%d panic=%v",
						event.EventName(), event.AggregateID(), r)
				}
			}()
			handler(event)
		}()
	}
}
