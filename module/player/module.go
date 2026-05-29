// Package player 是玩家限界上下文。
//
// 职责：玩家信息查询、头像修改、属性/货币管理。
//
// 战略分类：支撑域（Supporting）—— 支撑核心玩法，定制开发。
//
// 聚合根：Player（玩家属性 + 货币 + 等级）。
package player

import (
	"context"

	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/ddd"
	"github.com/skeletongo/game-stack/module/actor"
	"github.com/skeletongo/game-stack/module/player/application"
	"github.com/skeletongo/game-stack/module/player/domain"
	interfaces "github.com/skeletongo/game-stack/module/player/interface"
	"github.com/skeletongo/game-stack/stack"
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
	cmdBus.Register(application.CmdSetAvatar, &application.SetAvatarHandler{Repo: repo})
	cmdBus.Register(application.CmdAddExp, &application.AddExpHandler{Repo: repo, EventBus: eventBus})
	cmdBus.Register(application.CmdAddGold, &application.AddGoldHandler{Repo: repo, EventBus: eventBus})
	cmdBus.Register(application.CmdDeductGold, &application.DeductGoldHandler{Repo: repo, EventBus: eventBus})
	cmdBus.Register(application.CmdAddDiamond, &application.AddDiamondHandler{Repo: repo, EventBus: eventBus})
	cmdBus.Register(application.CmdDeductDiamond, &application.DeductDiamondHandler{Repo: repo, EventBus: eventBus})
	cmdBus.Register(application.CmdDeletePlayer, &application.DeletePlayerHandler{Repo: repo})

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
	RegisterGRPC(proxy, repo)

	// 注册清理回调（Grace Period 到期后清除玩家内存数据）
	if c, ok := stack.GetService("cleaner").(*stack.PlayerDoneCleaner); ok {
		c.Register(&cleanableAdapter{repo: repo})
	}

	// 注册跨模块 Service（供其他模块通过 stack.GetService("player") 获取）
	stack.RegisterService(name, &svcAdapter{repo: repo, cmdBus: cmdBus})

	log.Infof("[player] module initialized (DDD)")
	return nil
}

// ---- 适配器 ----

// cleanableAdapter 适配 PlayerRepository → stack.CleanableService。
type cleanableAdapter struct {
	repo domain.PlayerRepository
}

// CleanPlayerData 清理玩家内存数据（断线 Grace Period 到期时调用）。
func (a *cleanableAdapter) CleanPlayerData(uid int64) error {
	return a.repo.Delete(context.Background(), uid)
}

// svcAdapter 是 player 模块对外的服务适配器。
// 其他模块通过 stack.GetService("player") 获取，类型断言为 *svcAdapter。
//
// 提供的能力：
//   - GetPlayer(id) → 查询玩家信息
//   - AddExp / AddGold / DeductGold / AddDiamond / DeductDiamond → 修改玩家资源
type svcAdapter struct {
	repo   domain.PlayerRepository
	cmdBus *ddd.CommandBus
}

// GetPlayer 查询玩家（直接读仓储，不走 Actor）。
func (s *svcAdapter) GetPlayer(id int64) (*domain.Player, error) {
	return s.repo.Load(context.Background(), id)
}

// AddExp 增加经验值（通过命令总线，需在 Actor 内调用）。
func (s *svcAdapter) AddExp(id int64, exp int64) error {
	return s.cmdBus.Dispatch(context.Background(), application.AddExpCmd{PlayerID: id, Amount: exp})
}

// AddGold 增加金币。
func (s *svcAdapter) AddGold(id int64, gold int32) error {
	return s.cmdBus.Dispatch(context.Background(), application.AddGoldCmd{PlayerID: id, Amount: gold})
}

// DeductGold 扣除金币。
func (s *svcAdapter) DeductGold(id int64, gold int32) error {
	return s.cmdBus.Dispatch(context.Background(), application.DeductGoldCmd{PlayerID: id, Amount: gold})
}

// AddDiamond 增加钻石。
func (s *svcAdapter) AddDiamond(id int64, diamond int32) error {
	return s.cmdBus.Dispatch(context.Background(), application.AddDiamondCmd{PlayerID: id, Amount: diamond})
}

// DeductDiamond 扣除钻石。
func (s *svcAdapter) DeductDiamond(id int64, diamond int32) error {
	return s.cmdBus.Dispatch(context.Background(), application.DeductDiamondCmd{PlayerID: id, Amount: diamond})
}
