package mail

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/stack"
)

const name = "mail"

// Module 创建邮件模块。
func Module(opts ...Option) stack.Module {
	return &mailModule{opts: opts}
}

type mailModule struct {
	opts []Option
}

func (m *mailModule) Name() string { return name }

func (m *mailModule) Init(proxy *node.Proxy) error {
	o := defaultOptions()
	for _, opt := range m.opts {
		opt(o)
	}

	impl := newImpl(o.store)

	// 所有邮件路由都需要有状态+授权（玩家邮箱数据在同一节点）
	proxy.AddRouteHandler(stack.RouteMailList, impl.handleList, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteMailRead, impl.handleRead, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteMailReceiveAttach, impl.handleReceiveAttach, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteMailDelete, impl.handleDelete, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteMailSend, impl.handleSend, stack.StatefulAuthorizedRoute)

	// 注册服务供其他模块使用
	stack.RegisterService(name, impl.svc)

	log.Infof("[mail] module initialized")
	return nil
}
