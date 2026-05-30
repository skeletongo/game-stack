// Package player 是玩家限界上下文。
//
// 职责：玩家信息查询、头像修改、属性/货币管理。
//
// 战略分类：支撑域（Supporting）—— 支撑核心玩法，定制开发。
//
// 聚合根：Player（玩家属性 + 货币 + 等级）。
package player

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/ddd"
	"github.com/skeletongo/game-stack/module/actor"
	"github.com/skeletongo/game-stack/module/player/application"
	"github.com/skeletongo/game-stack/module/player/domain"
	grpc "github.com/skeletongo/game-stack/module/player/grpc"
	interfaces "github.com/skeletongo/game-stack/module/player/interface"
	"github.com/skeletongo/game-stack/stack"
	"github.com/skeletongo/game-stack/stack/debug"
)

const name = "player"

// Module 创建 Player 模块（DDD 四层架构）。
//
// 使用方式：
//
//	stack.WithModules(player.Module(player.WithRepository(myRedisRepo)))
func Module(opts ...Option) stack.Module {
	return &playerModule{opts: opts}
}

type playerModule struct {
	opts []Option
}

func (m *playerModule) Name() string { return name }

// Init 初始化模块：装配四层依赖、注册路由和 gRPC 服务。
//
// 客户端路由（2 个）：
//   - GetInfo：查询玩家信息
//   - SetAvatar：修改头像（Actor 串行化）
func (m *playerModule) Init(proxy *node.Proxy) error {
	o := defaultOptions()
	for _, opt := range m.opts {
		opt(o)
	}

	// ---- 基础设施层 ----
	repo := o.repo

	// ---- 领域层 ----
	eventBus := ddd.NewEventBus()
	cmdBus := ddd.NewCommandBus()

	// ---- 应用层：注册命令处理器（内部 + 跨模块）----
	ddd.Register(cmdBus, application.CmdSetAvatar, &application.SetAvatarHandler{Repo: repo})
	ddd.Register(cmdBus, application.CmdAddExp, &application.AddExpHandler{Repo: repo, EventBus: eventBus})
	ddd.Register(cmdBus, application.CmdAddGold, &application.AddGoldHandler{Repo: repo, EventBus: eventBus})
	ddd.Register(cmdBus, application.CmdDeductGold, &application.DeductGoldHandler{Repo: repo, EventBus: eventBus})
	ddd.Register(cmdBus, application.CmdAddDiamond, &application.AddDiamondHandler{Repo: repo, EventBus: eventBus})
	ddd.Register(cmdBus, application.CmdDeductDiamond, &application.DeductDiamondHandler{Repo: repo, EventBus: eventBus})
	ddd.Register(cmdBus, application.CmdDeletePlayer, &application.DeletePlayerHandler{Repo: repo})

	// ---- 接口层 ----
	routes := interfaces.NewRouteHandlers(cmdBus, repo)

	// 查询路由（不走 Actor）
	proxy.AddRouteHandler(stack.RoutePlayerGetInfo, routes.HandleGetInfo, stack.StatefulAuthorizedRoute)

	// 写操作路由（走 Actor 串行化）
	proxy.AddRouteHandler(stack.RoutePlayerSetAvatar, actor.RouteToActor(actor.KindPlayer), stack.StatefulAuthorizedRoute)

	// 注册 Actor 路由处理器（每个 Actor Spawn 时自动应用）
	actor.RegisterRouteInitializer(func(act *node.Actor) {
		act.AddRouteHandler(stack.RoutePlayerSetAvatar, routes.HandleSetAvatarActor)
	})

	// 注册 gRPC 服务（供其他节点调用）
	grpc.Register(name, proxy, repo)

	// 注册清理回调（Grace Period 到期后清除玩家内存数据）
	if c, ok := stack.GetService("cleaner").(*stack.PlayerDoneCleaner); ok {
		c.Register(&cleanableAdapter{repo: repo})
	}

	// 注册跨模块 Service（供其他模块通过 stack.GetService("player") 获取）
	stack.RegisterService(name, &svcAdapter{repo: repo, cmdBus: cmdBus})

	// 注册到 debug 服务（运行时查询/修改数据）
	debug.Register[*domain.Player](name, repo, cmdBus)

	log.Infof("[player] module initialized (DDD)")
	return nil
}
