package player

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/stack"
)

const name = "player"

// Module 创建玩家模块。
func Module(opts ...Option) stack.Module {
	return &playerModule{opts: opts}
}

type playerModule struct {
	opts []Option
}

func (m *playerModule) Name() string { return name }

func (m *playerModule) Init(proxy *node.Proxy) error {
	o := defaultOptions()
	for _, opt := range m.opts {
		opt(o)
	}

	impl := newImpl(o.store)

	// 注册路由（有状态+授权，保证玩家数据在同一节点）
	proxy.AddRouteHandler(stack.RoutePlayerGetInfo, impl.handleGetInfo, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RoutePlayerUpdate, impl.handleUpdate, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RoutePlayerDelete, impl.handleDelete, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RoutePlayerSearch, impl.handleSearch, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RoutePlayerSetAvatar, impl.handleSetAvatar, stack.StatefulAuthorizedRoute)

	// 注册到延迟清理器（断线 30 秒后清除玩家内存数据）
	if c, ok := stack.GetService("cleaner").(*stack.PlayerDoneCleaner); ok {
		c.Register(impl.svc)
	}

	// 注册服务供其他模块使用
	stack.RegisterService(name, impl.svc)

	log.Infof("[player] module initialized")
	return nil
}
