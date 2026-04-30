package activity

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/stack"
)

const name = "activity"

func Module(opts ...Option) stack.Module {
	return &activityModule{opts: opts}
}

type activityModule struct {
	opts []Option
}

func (m *activityModule) Name() string { return name }

func (m *activityModule) Init(proxy *node.Proxy) error {
	o := defaultOptions()
	for _, opt := range m.opts {
		opt(o)
	}

	impl := newImpl(o.store)

	proxy.AddRouteHandler(stack.RouteActivityList, impl.handleList, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteActivityClaim, impl.handleClaim, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteActivityInfo, impl.handleInfo, stack.StatefulAuthorizedRoute)

	if c, ok := stack.GetService("cleaner").(*stack.PlayerDoneCleaner); ok {
		c.Register(impl.svc)
	}

	stack.RegisterService(name, impl.svc)

	log.Infof("[activity] module initialized")
	return nil
}
