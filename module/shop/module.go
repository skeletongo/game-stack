package shop

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/stack"
)

const name = "shop"

func Module(opts ...Option) stack.Module {
	return &shopModule{opts: opts}
}

type shopModule struct {
	opts []Option
}

func (m *shopModule) Name() string { return name }

func (m *shopModule) Init(proxy *node.Proxy) error {
	o := defaultOptions()
	for _, opt := range m.opts {
		opt(o)
	}

	impl := newImpl(o.store)

	// ShopList 无需有状态（商品列表全局共享），仅需授权
	proxy.AddRouteHandler(stack.RouteShopList, impl.handleList, node.AuthorizedRoute)
	// ShopBuy 必须有状态（购买需要扣减玩家金币，必须路由到玩家所在节点）
	proxy.AddRouteHandler(stack.RouteShopBuy, impl.handleBuy, stack.StatefulAuthorizedRoute)

	stack.RegisterService(name, impl.svc)

	log.Infof("[shop] module initialized")
	return nil
}
