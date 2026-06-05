package svc

import (
	"context"

	"github.com/dobyte/due/v2/cluster/node"

	"github.com/skeletongo/game-stack/ddd"
	"github.com/skeletongo/game-stack/module/actor"
	"github.com/skeletongo/game-stack/module/player/internal/application"
	"github.com/skeletongo/game-stack/module/player/internal/domain"
	"github.com/skeletongo/game-stack/module/player/svc"
)

var _ svc.IPlayer = (*server)(nil)

// server 是 player 模块对外提供的接口
// 其他模块通过 stack.GetService("player") 获取，类型断言为 svc.IPlayer
type server struct {
	repo   domain.PlayerRepository
	cmdBus *ddd.CommandBus
	proxy  *node.Proxy
}

// New 创建 player 模块的跨模块 Service 实例。
//
// proxy 用于 InvokePlayerSync 将命令投递到玩家 Actor 中串行执行。
func New(repo domain.PlayerRepository, cmdBus *ddd.CommandBus, proxy *node.Proxy) svc.IPlayer {
	return &server{
		repo:   repo,
		cmdBus: cmdBus,
		proxy:  proxy,
	}
}

// GetPlayer 查询玩家（直接读仓储，不走 Actor）。
func (s *server) GetPlayer(ctx context.Context, id int64) (*svc.Player, error) {
	p, err := s.repo.Load(ctx, id)
	if err != nil {
		return nil, err
	}
	return playerToAPI(p), nil
}

// CreatePlayer 创建玩家档案（在玩家 Actor 中同步执行，注册账号时调用）。
func (s *server) CreatePlayer(ctx context.Context, id int64, nickname string) error {
	_, err := actor.InvokePlayerSync[*ddd.NoResult](ctx, s.proxy, id, func(ctx context.Context) (*ddd.NoResult, error) {
		_, err := s.cmdBus.Dispatch(ctx, application.CreatePlayerCmd{PlayerID: id, Nickname: nickname})
		return &ddd.NoResult{}, err
	})
	return err
}

// DeletePlayer 删除玩家档案（在玩家 Actor 中同步执行，注册补偿或内部清理时调用）。
func (s *server) DeletePlayer(ctx context.Context, id int64) error {
	_, err := actor.InvokePlayerSync[*ddd.NoResult](ctx, s.proxy, id, func(ctx context.Context) (*ddd.NoResult, error) {
		_, err := s.cmdBus.Dispatch(ctx, application.DeletePlayerCmd{PlayerID: id})
		return &ddd.NoResult{}, err
	})
	return err
}

// AddExp 增加经验值（在玩家 Actor 中同步执行），返回最新经验值。
func (s *server) AddExp(ctx context.Context, id int64, exp int64) (int64, error) {
	return actor.InvokePlayerSync[int64](ctx, s.proxy, id, func(ctx context.Context) (int64, error) {
		return ddd.Dispatch[int64](ctx, s.cmdBus, application.AddExpCmd{PlayerID: id, Amount: exp})
	})
}

// AddGold 增加金币（在玩家 Actor 中同步执行），返回最新金币数量。
func (s *server) AddGold(ctx context.Context, id int64, gold int32) (int32, error) {
	return actor.InvokePlayerSync[int32](ctx, s.proxy, id, func(ctx context.Context) (int32, error) {
		return ddd.Dispatch[int32](ctx, s.cmdBus, application.AddGoldCmd{PlayerID: id, Amount: gold})
	})
}

// DeductGold 扣除金币（在玩家 Actor 中同步执行），返回最新金币数量。
func (s *server) DeductGold(ctx context.Context, id int64, gold int32) (int32, error) {
	return actor.InvokePlayerSync[int32](ctx, s.proxy, id, func(ctx context.Context) (int32, error) {
		return ddd.Dispatch[int32](ctx, s.cmdBus, application.DeductGoldCmd{PlayerID: id, Amount: gold})
	})
}

// AddDiamond 增加钻石（在玩家 Actor 中同步执行），返回最新钻石数量。
func (s *server) AddDiamond(ctx context.Context, id int64, diamond int32) (int32, error) {
	return actor.InvokePlayerSync[int32](ctx, s.proxy, id, func(ctx context.Context) (int32, error) {
		return ddd.Dispatch[int32](ctx, s.cmdBus, application.AddDiamondCmd{PlayerID: id, Amount: diamond})
	})
}

// DeductDiamond 扣除钻石（在玩家 Actor 中同步执行），返回最新钻石数量。
func (s *server) DeductDiamond(ctx context.Context, id int64, diamond int32) (int32, error) {
	return actor.InvokePlayerSync[int32](ctx, s.proxy, id, func(ctx context.Context) (int32, error) {
		return ddd.Dispatch[int32](ctx, s.cmdBus, application.DeductDiamondCmd{PlayerID: id, Amount: diamond})
	})
}

func playerToAPI(p *domain.Player) *svc.Player {
	return &svc.Player{
		ID:        p.ID(),
		Nickname:  p.Nickname().String(),
		Level:     p.Level().Int32(),
		Exp:       p.Exp().Int64(),
		Avatar:    p.Avatar().String(),
		Gold:      p.Gold().Int32(),
		Diamond:   p.Diamond().Int32(),
		CreatedAt: p.CreatedAt(),
		UpdatedAt: p.UpdatedAt(),
	}
}
