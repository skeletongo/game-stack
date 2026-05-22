package stack

import "github.com/dobyte/due/v2/cluster/node"

// Module 是 game-stack 框架中所有游戏模块必须实现的接口。
// 每个模块在 Init 方法中通过 proxy 注册路由、事件处理器和钩子。
type Module interface {
	// Name 返回模块的唯一名称（用于日志和诊断）。
	Name() string

	// Init 在应用启动时调用。模块通过 proxy 注册路由处理器、
	// 事件处理器和钩子监听器。
	Init(proxy *node.Proxy) error
}

// CleanableService 是模块 Service 可选实现的接口。
// 实现了此接口的模块，会在玩家断线时被调用以清理该玩家的内存数据。
//
// CleanPlayerData 会重试直到成功（最多 maxRetries 次），
// 全部成功后才会解除节点绑定，防止清理失败导致数据丢失。
type CleanableService interface {
	CleanPlayerData(uid int64) error
}

// 类型断言辅助
var _ interface{} = CleanableService(nil)
