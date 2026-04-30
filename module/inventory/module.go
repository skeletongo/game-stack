package inventory

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/stack"
)

const name = "inventory"

// Module 创建背包模块。
func Module(opts ...Option) stack.Module {
	return &inventoryModule{opts: opts}
}

type inventoryModule struct {
	opts []Option
}

func (m *inventoryModule) Name() string { return name }

func (m *inventoryModule) Init(proxy *node.Proxy) error {
	o := defaultOptions()
	for _, opt := range m.opts {
		opt(o)
	}

	impl := newImpl(o.store, o)

	proxy.AddRouteHandler(stack.RouteInvList, impl.handleList, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteInvUse, impl.handleUse, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteInvEquip, impl.handleEquip, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteInvUnequip, impl.handleUnequip, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteInvDrop, impl.handleDrop, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteInvSell, impl.handleSell, stack.StatefulAuthorizedRoute)

	if c, ok := stack.GetService("cleaner").(*stack.PlayerDoneCleaner); ok {
		c.Register(impl.svc)
	}

	stack.RegisterService(name, impl.svc)

	log.Infof("[inventory] module initialized")
	return nil
}
