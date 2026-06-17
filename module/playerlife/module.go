package playerlife

import (
	"github.com/dobyte/due/v2/cluster"
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/skeletongo/game-stack/stack"
)

var global *Manager

// Get 返回进程级玩家生命周期管理器。
func Get() *Manager {
	if global == nil {
		panic("playerlife manager not init")
	}
	return global
}

// Module 创建玩家生命周期组件。
func Module(opts ...Option) stack.Module {
	return &lifeModule{opts: opts}
}

type lifeModule struct {
	opts []Option
}

func (m *lifeModule) Name() string {
	return "playerlife"
}

func (m *lifeModule) Init(proxy *node.Proxy) error {
	o := defaultOptions()
	for _, opt := range m.opts {
		opt(o)
	}

	global = NewManager(proxy, o.delay, o.maxRetries)
	stack.AddEventHandler(cluster.Connect, global.handleConnect)
	stack.AddEventHandler(cluster.Disconnect, global.handleDisconnect)
	return nil
}
