// Package interfaces 提供 player 模块的接口层实现。
//
// 对外暴露 2 个客户端路由：
//   - GetInfo：查询玩家信息（不走 Actor）
//   - SetAvatar：修改头像（走 Actor 串行化）
package interfaces

import (
	"context"

	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/ddd"
	"github.com/skeletongo/game-stack/module/player/application"
	"github.com/skeletongo/game-stack/module/player/domain"
	"github.com/skeletongo/game-stack/proto/player"
	"github.com/skeletongo/game-stack/stack"
)

// RouteHandlers 持有路由处理器所需的依赖。
type RouteHandlers struct {
	CmdBus  *ddd.CommandBus
	Repo    domain.PlayerRepository
	GetInfo *application.GetPlayerHandler
}

// NewRouteHandlers 创建路由处理器。
func NewRouteHandlers(cmdBus *ddd.CommandBus, repo domain.PlayerRepository) *RouteHandlers {
	return &RouteHandlers{
		CmdBus:  cmdBus,
		Repo:    repo,
		GetInfo: &application.GetPlayerHandler{Repo: repo},
	}
}

// ---- 查询路由（不走 Actor）----

// HandleGetInfo 查询玩家信息。
func (h *RouteHandlers) HandleGetInfo(ctx node.Context) {
	req := &player.GetInfoRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}
	pid := req.PlayerId
	if pid == 0 {
		pid = ctx.UID()
	}
	p, err := h.GetInfo.GetPlayer(context.Background(), pid)
	if err != nil {
		stack.RespondError(ctx, stack.ErrPlayerNotFound)
		return
	}
	stack.RespondData(ctx, &player.GetInfoResponse{Player: application.PlayerToProto(p)})
}

// ---- Actor 路由处理器（走 Actor 串行化）----

// HandleSetAvatarActor 在 Actor 内处理头像修改。
func (h *RouteHandlers) HandleSetAvatarActor(ctx node.Context) {
	req := &player.SetAvatarRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}
	cmd := application.SetAvatarCmd{PlayerID: ctx.UID(), Avatar: req.Avatar}
	if err := h.CmdBus.Dispatch(context.Background(), cmd); err != nil {
		log.Errorf("[player] set avatar failed: uid=%d err=%v", ctx.UID(), err)
		stack.RespondError(ctx, stack.ErrInternalError)
		return
	}
	stack.RespondOK(ctx)
}
