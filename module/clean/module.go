package clean

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/skeletongo/game-stack/stack"
)

var global *PlayerDoneCleaner

func Get() *PlayerDoneCleaner {
	if global == nil {
		panic("player cleaner not init!")
	}
	return global
}

func Module(opts ...Option) stack.Module {
	return &cleanModule{opts: opts}
}

type cleanModule struct {
	opts []Option
}

func (m *cleanModule) Name() string {
	return "player-cleaner"
}

func (m *cleanModule) Init(proxy *node.Proxy) error {
	o := defaultOptions()
	for _, opt := range m.opts {
		opt(o)
	}

	global = NewPlayerDoneCleaner(proxy, o.delay, o.maxRetries)
	return nil
}
