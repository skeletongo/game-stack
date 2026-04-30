package room

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/stack"
)

const name = "room"

// Module 创建房间模块。
func Module(opts ...Option) stack.Module {
	return &roomModule{opts: opts}
}

type roomModule struct {
	opts []Option
}

func (m *roomModule) Name() string { return name }

func (m *roomModule) Init(proxy *node.Proxy) error {
	o := defaultOptions()
	for _, opt := range m.opts {
		opt(o)
	}

	impl := newImpl(o.store, o)

	proxy.AddRouteHandler(stack.RouteRoomCreate, impl.handleCreate, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteRoomJoin, impl.handleJoin, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteRoomLeave, impl.handleLeave, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteRoomList, impl.handleList, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteRoomKick, impl.handleKick, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteRoomReady, impl.handleReady, stack.StatefulAuthorizedRoute)

	stack.RegisterService(name, impl.svc)

	log.Infof("[room] module initialized")
	return nil
}
