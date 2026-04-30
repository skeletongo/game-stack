package auth

import (
	"time"

	"github.com/dobyte/due/v2/cluster"
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/stack"
)

const name = "auth"

// Module 创建认证模块。
func Module(opts ...Option) stack.Module {
	return &authModule{opts: opts}
}

type authModule struct {
	opts []Option
}

func (m *authModule) Name() string { return name }

func (m *authModule) Init(proxy *node.Proxy) error {
	o := defaultOptions()
	for _, opt := range m.opts {
		opt(o)
	}

	// 创建玩家断线延迟清理器（30 秒 Grace Period）
	cleaner := stack.NewPlayerDoneCleaner(proxy, 30*time.Second)
	stack.RegisterService("cleaner", cleaner)

	impl := newImpl(o.store, cleaner)

	// 注册路由（无需授权的路由不传 RouteOptions）
	proxy.AddRouteHandler(stack.RouteAuthLogin, impl.handleLogin)
	proxy.AddRouteHandler(stack.RouteAuthRegister, impl.handleRegister)
	// 需要授权的路由使用有状态+授权路由
	proxy.AddRouteHandler(stack.RouteAuthLogout, impl.handleLogout, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteAuthTokenRefresh, impl.handleRefresh, stack.StatefulAuthorizedRoute)

	// 注册连接事件处理器（事件不能调用 ctx.Response）
	proxy.AddEventHandler(cluster.Connect, impl.handleConnect)
	proxy.AddEventHandler(cluster.Disconnect, impl.handleDisconnect)

	// 注册服务供其他模块使用
	stack.RegisterService(name, impl.svc)

	log.Infof("[auth] module initialized")
	return nil
}
