package match

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/stack"
)

const name = "match"

// Module 创建匹配模块。
func Module(opts ...Option) stack.Module {
	return &matchModule{opts: opts}
}

type matchModule struct {
	opts []Option
}

func (m *matchModule) Name() string { return name }

func (m *matchModule) Init(proxy *node.Proxy) error {
	o := defaultOptions()
	for _, opt := range m.opts {
		opt(o)
	}

	impl := newImpl(o.store)

	proxy.AddRouteHandler(stack.RouteMatchJoin, impl.handleJoin, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteMatchLeave, impl.handleLeave, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteMatchCancel, impl.handleCancel, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteMatchStatus, impl.handleStatus, stack.StatefulAuthorizedRoute)

	stack.RegisterService(name, impl.svc)

	log.Infof("[match] module initialized")
	return nil
}
