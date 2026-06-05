package application

import (
	"context"

	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/ddd"
	"github.com/skeletongo/game-stack/module/player/internal/domain"
)

// CreatePlayerHandler 处理玩家创建。
type CreatePlayerHandler struct {
	Repo domain.PlayerRepository
}

var _ ddd.CommandHandler[CreatePlayerCmd, ddd.NoResult] = (*CreatePlayerHandler)(nil)

func (h *CreatePlayerHandler) Handle(ctx context.Context, cmd CreatePlayerCmd) (ddd.NoResult, error) {
	if _, err := h.Repo.Load(ctx, cmd.PlayerID); err == nil {
		return ddd.NoResult{}, nil
	}
	p, err := domain.NewPlayer(cmd.PlayerID, cmd.Nickname, 1, 0, 0, 0, "")
	if err != nil {
		return ddd.NoResult{}, err
	}
	if err := h.Repo.Save(ctx, p); err != nil {
		return ddd.NoResult{}, err
	}
	log.Infof("[player] player created: uid=%d nickname=%s", cmd.PlayerID, cmd.Nickname)
	return ddd.NoResult{}, nil
}

// GetPlayerHandler 处理玩家查询。
// 直接使用仓储（查询不需要 Actor 串行化）。
type GetPlayerHandler struct {
	Repo domain.PlayerRepository
}

var _ ddd.CommandHandler[GetPlayerCmd, *domain.Player] = (*GetPlayerHandler)(nil)

// Handle 实现 CommandHandler，通过命令总线分发查询。
func (h *GetPlayerHandler) Handle(ctx context.Context, cmd GetPlayerCmd) (*domain.Player, error) {
	return h.Repo.Load(ctx, cmd.TargetID)
}

// SetAvatarHandler 处理设置头像命令。
type SetAvatarHandler struct {
	Repo domain.PlayerRepository
}

var _ ddd.CommandHandler[SetAvatarCmd, ddd.NoResult] = (*SetAvatarHandler)(nil)

// Handle 执行头像设置：加载聚合 → 修改头像（保留昵称）→ 保存。返回 ddd.NoResult。
func (h *SetAvatarHandler) Handle(ctx context.Context, cmd SetAvatarCmd) (ddd.NoResult, error) {
	p, err := h.Repo.Load(ctx, cmd.PlayerID)
	if err != nil {
		return ddd.NoResult{}, err
	}
	p.UpdateProfile(p.Nickname(), domain.Avatar(cmd.Avatar))
	return ddd.NoResult{}, h.Repo.Save(ctx, p)
}

// AddExpHandler 处理增加经验值命令。
type AddExpHandler struct {
	Repo     domain.PlayerRepository
	EventBus *ddd.EventBus
}

var _ ddd.CommandHandler[AddExpCmd, int64] = (*AddExpHandler)(nil)

// Handle 执行经验增加：加载聚合 → AddExp → 保存 → 若升级则发布事件。返回最新经验值。
func (h *AddExpHandler) Handle(ctx context.Context, cmd AddExpCmd) (int64, error) {
	p, err := h.Repo.Load(ctx, cmd.PlayerID)
	if err != nil {
		return 0, err
	}
	oldLevel := p.Level().Int32()
	leveledUp := p.AddExp(cmd.Amount)
	if err := h.Repo.Save(ctx, p); err != nil {
		return 0, err
	}
	if leveledUp {
		h.EventBus.Publish(domain.NewPlayerLeveledUp(cmd.PlayerID, oldLevel, p.Level().Int32()))
		log.Infof("[player] leveled up: uid=%d %d->%d", cmd.PlayerID, oldLevel, p.Level().Int32())
	}
	return p.Exp().Int64(), nil
}

// AddGoldHandler 处理增加金币命令。
type AddGoldHandler struct {
	Repo     domain.PlayerRepository
	EventBus *ddd.EventBus
}

var _ ddd.CommandHandler[AddGoldCmd, int32] = (*AddGoldHandler)(nil)

// Handle 执行金币增加：返回最新金币数量。
func (h *AddGoldHandler) Handle(ctx context.Context, cmd AddGoldCmd) (int32, error) {
	p, err := h.Repo.Load(ctx, cmd.PlayerID)
	if err != nil {
		return 0, err
	}
	if err := p.AddGold(cmd.Amount); err != nil {
		return 0, err
	}
	if err := h.Repo.Save(ctx, p); err != nil {
		return 0, err
	}
	h.EventBus.Publish(domain.NewGoldChanged(cmd.PlayerID, cmd.Amount, p.Gold().Int32()))
	return p.Gold().Int32(), nil
}

// DeductGoldHandler 处理扣除金币命令。
type DeductGoldHandler struct {
	Repo     domain.PlayerRepository
	EventBus *ddd.EventBus
}

var _ ddd.CommandHandler[DeductGoldCmd, int32] = (*DeductGoldHandler)(nil)

// Handle 执行金币扣除：返回最新金币数量。
func (h *DeductGoldHandler) Handle(ctx context.Context, cmd DeductGoldCmd) (int32, error) {
	p, err := h.Repo.Load(ctx, cmd.PlayerID)
	if err != nil {
		return 0, err
	}
	if err := p.DeductGold(cmd.Amount); err != nil {
		return 0, err
	}
	if err := h.Repo.Save(ctx, p); err != nil {
		return 0, err
	}
	h.EventBus.Publish(domain.NewGoldChanged(cmd.PlayerID, -cmd.Amount, p.Gold().Int32()))
	return p.Gold().Int32(), nil
}

// AddDiamondHandler 处理增加钻石命令。
type AddDiamondHandler struct {
	Repo     domain.PlayerRepository
	EventBus *ddd.EventBus
}

var _ ddd.CommandHandler[AddDiamondCmd, int32] = (*AddDiamondHandler)(nil)

// Handle 执行钻石增加：返回最新钻石数量。
func (h *AddDiamondHandler) Handle(ctx context.Context, cmd AddDiamondCmd) (int32, error) {
	p, err := h.Repo.Load(ctx, cmd.PlayerID)
	if err != nil {
		return 0, err
	}
	if err := p.AddDiamond(cmd.Amount); err != nil {
		return 0, err
	}
	if err := h.Repo.Save(ctx, p); err != nil {
		return 0, err
	}
	h.EventBus.Publish(domain.NewDiamondChanged(cmd.PlayerID, cmd.Amount, p.Diamond().Int32()))
	return p.Diamond().Int32(), nil
}

// DeductDiamondHandler 处理扣除钻石命令。
type DeductDiamondHandler struct {
	Repo     domain.PlayerRepository
	EventBus *ddd.EventBus
}

var _ ddd.CommandHandler[DeductDiamondCmd, int32] = (*DeductDiamondHandler)(nil)

// Handle 执行钻石扣除：返回最新钻石数量。
func (h *DeductDiamondHandler) Handle(ctx context.Context, cmd DeductDiamondCmd) (int32, error) {
	p, err := h.Repo.Load(ctx, cmd.PlayerID)
	if err != nil {
		return 0, err
	}
	if err := p.DeductDiamond(cmd.Amount); err != nil {
		return 0, err
	}
	if err := h.Repo.Save(ctx, p); err != nil {
		return 0, err
	}
	h.EventBus.Publish(domain.NewDiamondChanged(cmd.PlayerID, -cmd.Amount, p.Diamond().Int32()))
	return p.Diamond().Int32(), nil
}

// DeletePlayerHandler 处理删除玩家命令。
type DeletePlayerHandler struct {
	Repo domain.PlayerRepository
}

var _ ddd.CommandHandler[DeletePlayerCmd, ddd.NoResult] = (*DeletePlayerHandler)(nil)

// Handle 执行玩家删除。返回 ddd.NoResult。
func (h *DeletePlayerHandler) Handle(ctx context.Context, cmd DeletePlayerCmd) (ddd.NoResult, error) {
	return ddd.NoResult{}, h.Repo.Delete(ctx, cmd.PlayerID)
}
