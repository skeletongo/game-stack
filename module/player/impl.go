package player

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/protocol/player"
	"github.com/skeletongo/game-stack/stack"
)

var (
	ErrNotEnoughGold    = errors.New("not enough gold")
	ErrNotEnoughDiamond = errors.New("not enough diamond")
)

type impl struct {
	svc *service
}

func newImpl(store Store) *impl {
	return &impl{svc: newService(store)}
}

func (i *impl) handleGetInfo(ctx node.Context) {
	req := &player.GetInfoRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	pid := req.PlayerID
	if pid == 0 {
		pid = ctx.UID()
	}

	p, err := i.svc.store.GetPlayer(context.Background(), pid)
	if err != nil {
		stack.RespondError(ctx, stack.ErrPlayerNotFound)
		return
	}

	stack.RespondData(ctx, &player.GetInfoResponse{
		Player: &player.PlayerInfo{
			ID:        p.ID,
			Nickname:  p.Nickname,
			Level:     p.Level,
			Exp:       p.Exp,
			Avatar:    p.Avatar,
			Gold:      p.Gold,
			Diamond:   p.Diamond,
			CreatedAt: p.CreatedAt,
		},
	})
}

func (i *impl) handleUpdate(ctx node.Context) {
	req := &player.UpdateProfileRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	p, err := i.svc.store.GetPlayer(context.Background(), ctx.UID())
	if err != nil {
		stack.RespondError(ctx, stack.ErrPlayerNotFound)
		return
	}

	if req.Nickname != "" {
		if len(req.Nickname) > 16 {
			stack.RespondError(ctx, stack.ErrNameTooLong)
			return
		}
		if existing, err := i.svc.store.GetPlayerByName(context.Background(), req.Nickname); err == nil && existing.ID != p.ID {
			stack.RespondError(ctx, stack.ErrNicknameExists)
			return
		}
		p.Nickname = req.Nickname
	}

	if req.Avatar != "" {
		p.Avatar = req.Avatar
	}

	p.UpdatedAt = time.Now().Unix()
	if err := i.svc.store.UpdatePlayer(context.Background(), p); err != nil {
		log.Errorf("update player failed: %v", err)
		stack.RespondError(ctx, stack.ErrInternalError)
		return
	}

	stack.RespondOK(ctx)
}

func (i *impl) handleSetAvatar(ctx node.Context) {
	req := &player.UpdateProfileRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	if req.Avatar == "" {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	p, err := i.svc.store.GetPlayer(context.Background(), ctx.UID())
	if err != nil {
		stack.RespondError(ctx, stack.ErrPlayerNotFound)
		return
	}

	p.Avatar = req.Avatar
	p.UpdatedAt = time.Now().Unix()

	if err := i.svc.store.UpdatePlayer(context.Background(), p); err != nil {
		stack.RespondError(ctx, stack.ErrInternalError)
		return
	}

	stack.RespondOK(ctx)
}

func (i *impl) handleSearch(ctx node.Context) {
	req := &player.GetInfoRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	if req.PlayerID == 0 {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	p, err := i.svc.store.GetPlayer(context.Background(), req.PlayerID)
	if err != nil {
		stack.RespondError(ctx, stack.ErrPlayerNotFound)
		return
	}

	stack.RespondData(ctx, &player.GetInfoResponse{
		Player: &player.PlayerInfo{
			ID:        p.ID,
			Nickname:  p.Nickname,
			Level:     p.Level,
			Avatar:    p.Avatar,
			CreatedAt: p.CreatedAt,
		},
	})
}

func (i *impl) handleDelete(ctx node.Context) {
	uid := ctx.UID()
	if uid == 0 {
		stack.RespondError(ctx, stack.ErrUnauthorized)
		return
	}
	_ = i.svc.store.UpdatePlayer(context.Background(), &Player{ID: uid})
	stack.RespondOK(ctx)
	fmt.Println("player deleted:", uid)
}
