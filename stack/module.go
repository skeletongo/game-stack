package stack

import "github.com/dobyte/due/v2/cluster/node"

// Module 是game-stack框架中所有游戏模块必须实现的接口。
// 每个模块在Init方法中通过proxy注册路由、事件处理器和钩子。
type Module interface {
	// Name 返回模块的唯一名称（用于日志和诊断）。
	Name() string

	// Init 在应用启动时调用。模块通过proxy注册路由处理器、
	// 事件处理器和钩子监听器。
	Init(proxy *node.Proxy) error
}
