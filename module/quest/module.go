package quest

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/stack"
)

const name = "quest"

func Module(opts ...Option) stack.Module {
	return &questModule{opts: opts}
}

type questModule struct {
	opts []Option
}

func (m *questModule) Name() string { return name }

func (m *questModule) Init(proxy *node.Proxy) error {
	o := defaultOptions()
	for _, opt := range m.opts {
		opt(o)
	}

	impl := newImpl(o.store)

	proxy.AddRouteHandler(stack.RouteQuestList, impl.handleList, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteQuestAccept, impl.handleAccept, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteQuestSubmit, impl.handleSubmit, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteQuestAbandon, impl.handleAbandon, stack.StatefulAuthorizedRoute)

	stack.RegisterService(name, impl.svc)

	log.Infof("[quest] module initialized")
	return nil
}
