package application

import (
	"context"

	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/ddd"
	"github.com/skeletongo/game-stack/module/player/domain"
)

// GetPlayerHandler 处理玩家查询。
// 直接使用仓储（查询不需要 Actor 串行化）。
type GetPlayerHandler struct {
	Repo domain.PlayerRepository
}

// GetPlayer 加载玩家聚合。
func (h *GetPlayerHandler) GetPlayer(ctx context.Context, playerID int64) (*domain.Player, error) {
	return h.Repo.Load(ctx, playerID)
}

// SetAvatarHandler 处理设置头像命令。
type SetAvatarHandler struct {
	Repo domain.PlayerRepository
}

var _ ddd.CommandHandler[SetAvatarCmd] = (*SetAvatarHandler)(nil)

// Handle 执行头像设置：加载聚合 → 修改头像（保留昵称）→ 保存。
func (h *SetAvatarHandler) Handle(ctx context.Context, cmd SetAvatarCmd) error {
	p, err := h.Repo.Load(ctx, cmd.PlayerID)
	if err != nil {
		return err
	}
	p.UpdateProfile(p.Nickname(), domain.Avatar(cmd.Avatar))
	return h.Repo.Save(ctx, p)
}

// AddExpHandler 处理增加经验值命令。
type AddExpHandler struct {
	Repo     domain.PlayerRepository
	EventBus *ddd.EventBus
}

var _ ddd.CommandHandler[AddExpCmd] = (*AddExpHandler)(nil)

// Handle 执行经验增加：加载聚合 → AddExp → 保存 → 若升级则发布事件。
func (h *AddExpHandler) Handle(ctx context.Context, cmd AddExpCmd) error {
	p, err := h.Repo.Load(ctx, cmd.PlayerID)
	if err != nil {
		return err
	}
	oldLevel := p.Level().Int32()
	leveledUp := p.AddExp(cmd.Amount)
	if err := h.Repo.Save(ctx, p); err != nil {
		return err
	}
	if leveledUp {
		h.EventBus.Publish(domain.NewPlayerLeveledUp(cmd.PlayerID, oldLevel, p.Level().Int32()))
		log.Infof("[player] leveled up: uid=%d %d->%d", cmd.PlayerID, oldLevel, p.Level().Int32())
	}
	return nil
}

// AddGoldHandler 处理增加金币命令。
type AddGoldHandler struct {
	Repo     domain.PlayerRepository
	EventBus *ddd.EventBus
}

var _ ddd.CommandHandler[AddGoldCmd] = (*AddGoldHandler)(nil)

func (h *AddGoldHandler) Handle(ctx context.Context, cmd AddGoldCmd) error {
	p, err := h.Repo.Load(ctx, cmd.PlayerID)
	if err != nil {
		return err
	}
	if err := p.AddGold(cmd.Amount); err != nil {
		return err
	}
	if err := h.Repo.Save(ctx, p); err != nil {
		return err
	}
	h.EventBus.Publish(domain.NewGoldChanged(cmd.PlayerID, cmd.Amount, p.Gold().Int32()))
	return nil
}

// DeductGoldHandler 处理扣除金币命令。
type DeductGoldHandler struct {
	Repo     domain.PlayerRepository
	EventBus *ddd.EventBus
}

var _ ddd.CommandHandler[DeductGoldCmd] = (*DeductGoldHandler)(nil)

func (h *DeductGoldHandler) Handle(ctx context.Context, cmd DeductGoldCmd) error {
	p, err := h.Repo.Load(ctx, cmd.PlayerID)
	if err != nil {
		return err
	}
	if err := p.DeductGold(cmd.Amount); err != nil {
		return err
	}
	if err := h.Repo.Save(ctx, p); err != nil {
		return err
	}
	h.EventBus.Publish(domain.NewGoldChanged(cmd.PlayerID, -cmd.Amount, p.Gold().Int32()))
	return nil
}

// AddDiamondHandler 处理增加钻石命令。
type AddDiamondHandler struct {
	Repo     domain.PlayerRepository
	EventBus *ddd.EventBus
}

var _ ddd.CommandHandler[AddDiamondCmd] = (*AddDiamondHandler)(nil)

func (h *AddDiamondHandler) Handle(ctx context.Context, cmd AddDiamondCmd) error {
	p, err := h.Repo.Load(ctx, cmd.PlayerID)
	if err != nil {
		return err
	}
	if err := p.AddDiamond(cmd.Amount); err != nil {
		return err
	}
	if err := h.Repo.Save(ctx, p); err != nil {
		return err
	}
	h.EventBus.Publish(domain.NewDiamondChanged(cmd.PlayerID, cmd.Amount, p.Diamond().Int32()))
	return nil
}

// DeductDiamondHandler 处理扣除钻石命令。
type DeductDiamondHandler struct {
	Repo     domain.PlayerRepository
	EventBus *ddd.EventBus
}

var _ ddd.CommandHandler[DeductDiamondCmd] = (*DeductDiamondHandler)(nil)

func (h *DeductDiamondHandler) Handle(ctx context.Context, cmd DeductDiamondCmd) error {
	p, err := h.Repo.Load(ctx, cmd.PlayerID)
	if err != nil {
		return err
	}
	if err := p.DeductDiamond(cmd.Amount); err != nil {
		return err
	}
	if err := h.Repo.Save(ctx, p); err != nil {
		return err
	}
	h.EventBus.Publish(domain.NewDiamondChanged(cmd.PlayerID, -cmd.Amount, p.Diamond().Int32()))
	return nil
}

// DeletePlayerHandler 处理删除玩家命令。
type DeletePlayerHandler struct {
	Repo domain.PlayerRepository
}

var _ ddd.CommandHandler[DeletePlayerCmd] = (*DeletePlayerHandler)(nil)

func (h *DeletePlayerHandler) Handle(ctx context.Context, cmd DeletePlayerCmd) error {
	return h.Repo.Delete(ctx, cmd.PlayerID)
}
