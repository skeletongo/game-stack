package leaderboard

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/stack"
)

const name = "leaderboard"

func Module(opts ...Option) stack.Module {
	return &lbModule{opts: opts}
}

type lbModule struct {
	opts []Option
}

func (m *lbModule) Name() string { return name }

func (m *lbModule) Init(proxy *node.Proxy) error {
	o := defaultOptions()
	for _, opt := range m.opts {
		opt(o)
	}

	impl := newImpl(o.store)

	proxy.AddRouteHandler(stack.RouteLeaderboardGet, impl.handleGet, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteLeaderboardRank, impl.handleRank, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteLeaderboardTop, impl.handleTop, stack.StatefulAuthorizedRoute)

	stack.RegisterService(name, impl.svc)

	log.Infof("[leaderboard] module initialized")
	return nil
}
