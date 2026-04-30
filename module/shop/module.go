package shop

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/module/actor"
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

	// ShopList — 全局数据，无状态路由，仅需授权
	proxy.AddRouteHandler(stack.RouteShopList, impl.handleList, node.AuthorizedRoute)

	// ShopBuy — 使用 RouteToActor（模式2）：
	//   ① Node 路由处理器将消息投递到 PlayerActor 的 mailbox
	//   ② PlayerActor 的 dispatch goroutine 串行处理
	//   ③ ctx.Response() 在 Actor 中同步返回结果
	proxy.AddRouteHandler(stack.RouteShopBuy,
		actor.RouteToActor(actor.KindPlayer),
		stack.StatefulAuthorizedRoute,
	)

	// 注册 Actor 路由初始化器：当 PlayerActor 被 Spawn 时，
	// 自动为 Actor 注册 ShopBuy 的实际处理器
	actor.RegisterRouteInitializer(func(act *node.Actor) {
		act.AddRouteHandler(stack.RouteShopBuy, impl.handleBuyActor)
		log.Debugf("[shop] registered actor route: ShopBuy")
	})

	stack.RegisterService(name, impl.svc)

	log.Infof("[shop] module initialized")
	return nil
}
