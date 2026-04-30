package chat

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/stack"
)

const name = "chat"

// Module 创建聊天模块。
func Module(opts ...Option) stack.Module {
	return &chatModule{opts: opts}
}

type chatModule struct {
	opts []Option
}

func (m *chatModule) Name() string { return name }

func (m *chatModule) Init(proxy *node.Proxy) error {
	o := defaultOptions()
	for _, opt := range m.opts {
		opt(o)
	}

	impl := newImpl(o.store, o)

	proxy.AddRouteHandler(stack.RouteChatSend, impl.handleSend, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteChatHistory, impl.handleHistory, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteChatWorld, impl.handleWorldChat, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteChatPrivate, impl.handlePrivateChat, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteChatGuild, impl.handleGuildChat, stack.StatefulAuthorizedRoute)

	stack.RegisterService(name, impl.svc)

	log.Infof("[chat] module initialized")
	return nil
}
