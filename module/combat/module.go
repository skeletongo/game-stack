package combat

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/stack"
)

const name = "combat"

func Module(opts ...Option) stack.Module {
	return &combatModule{opts: opts}
}

type combatModule struct {
	opts []Option
}

func (m *combatModule) Name() string { return name }

func (m *combatModule) Init(proxy *node.Proxy) error {
	o := defaultOptions()
	for _, opt := range m.opts {
		opt(o)
	}

	impl := newImpl(o.store)

	proxy.AddRouteHandler(stack.RouteCombatSkillCast, impl.handleSkillCast, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteCombatMove, impl.handleMove, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteCombatTarget, impl.handleTarget, stack.StatefulAuthorizedRoute)

	stack.RegisterService(name, impl.svc)

	log.Infof("[combat] module initialized")
	return nil
}
