package guild

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/stack"
)

const name = "guild"

func Module(opts ...Option) stack.Module {
	return &guildModule{opts: opts}
}

type guildModule struct {
	opts []Option
}

func (m *guildModule) Name() string { return name }

func (m *guildModule) Init(proxy *node.Proxy) error {
	o := defaultOptions()
	for _, opt := range m.opts {
		opt(o)
	}

	impl := newImpl(o.store)

	proxy.AddRouteHandler(stack.RouteGuildCreate, impl.handleCreate, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteGuildJoin, impl.handleJoin, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteGuildLeave, impl.handleLeave, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteGuildKick, impl.handleKick, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteGuildList, impl.handleList, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteGuildInfo, impl.handleInfo, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteGuildDonate, impl.handleDonate, stack.StatefulAuthorizedRoute)

	stack.RegisterService(name, impl.svc)

	log.Infof("[guild] module initialized")
	return nil
}
