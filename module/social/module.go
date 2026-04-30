package social

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/stack"
)

const name = "social"

func Module(opts ...Option) stack.Module {
	return &socialModule{opts: opts}
}

type socialModule struct {
	opts []Option
}

func (m *socialModule) Name() string { return name }

func (m *socialModule) Init(proxy *node.Proxy) error {
	o := defaultOptions()
	for _, opt := range m.opts {
		opt(o)
	}

	impl := newImpl(o.store)

	proxy.AddRouteHandler(stack.RouteSocialFriendList, impl.handleFriendList, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteSocialFriendAdd, impl.handleAdd, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteSocialFriendRemove, impl.handleRemove, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteSocialBlock, impl.handleBlock, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteSocialUnblock, impl.handleUnblock, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteSocialBlacklist, impl.handleBlacklist, stack.StatefulAuthorizedRoute)

	stack.RegisterService(name, impl.svc)

	log.Infof("[social] module initialized")
	return nil
}
