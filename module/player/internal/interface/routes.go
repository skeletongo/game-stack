// Package interfaces 提供 player 模块的接口层实现。
//
// 对外暴露 2 个客户端路由：
//   - GetInfo：查询玩家信息（不走 Actor）
//   - SetAvatar：修改头像（走 Actor 串行化）
package interfaces

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/ddd"
	"github.com/skeletongo/game-stack/module/player/internal/application"
	"github.com/skeletongo/game-stack/module/player/internal/domain"
	"github.com/skeletongo/game-stack/proto/common"
	"github.com/skeletongo/game-stack/proto/player"
	"github.com/skeletongo/game-stack/stack"
)

// RouteHandlers 持有路由处理器所需的依赖。
type RouteHandlers struct {
	CmdBus *ddd.CommandBus
	Repo   domain.PlayerRepository
}

// NewRouteHandlers 创建路由处理器。
func NewRouteHandlers(cmdBus *ddd.CommandBus, repo domain.PlayerRepository) *RouteHandlers {
	return &RouteHandlers{
		CmdBus: cmdBus,
		Repo:   repo,
	}
}

// ---- 查询路由（不走 Actor）----

// HandleGetInfo 查询玩家信息。
func (h *RouteHandlers) HandleGetInfo(ctx node.Context) {
	req := &player.GetInfoRequest{}
	if err := ctx.Parse(req); err != nil {
		log.Errorf("[player] HandleGetInfo parse failed: %v", err)
		stack.ProtoResponse(ctx, &player.GetInfoResponse{Code: int32(common.SysError_INVALID_PARAM), Message: err.Error()})
		return
	}
	pid := req.PlayerId
	if pid == 0 {
		pid = ctx.UID()
	}
	p, err := ddd.Dispatch[*domain.Player](ctx.Context(), h.CmdBus,
		application.GetPlayerCmd{PlayerID: ctx.UID(), TargetID: pid})
	if err != nil {
		log.Errorf("[player] HandleGetInfo load failed: %v", err)
		stack.ProtoResponse(ctx, &player.GetInfoResponse{Code: stack.ErrCode(err), Message: err.Error()})
		return
	}
	stack.ProtoResponse(ctx, &player.GetInfoResponse{Code: stack.CodeOK, Player: application.PlayerToProto(p)})
}

// ---- Actor 路由处理器（走 Actor 串行化）----

// HandleSetAvatarActor 在 Actor 内处理头像修改。
func (h *RouteHandlers) HandleSetAvatarActor(ctx node.Context) {
	req := &player.SetAvatarRequest{}
	if err := ctx.Parse(req); err != nil {
		log.Errorf("[player] HandleSetAvatar parse failed: %v", err)
		stack.ProtoResponse(ctx, &player.SetAvatarResponse{Code: int32(common.SysError_INVALID_PARAM), Message: err.Error()})
		return
	}
	cmd := application.SetAvatarCmd{PlayerID: ctx.UID(), Avatar: req.Avatar}
	if _, err := h.CmdBus.Dispatch(ctx.Context(), cmd); err != nil {
		log.Errorf("[player] set avatar failed: uid=%d err=%v", ctx.UID(), err)
		stack.ProtoResponse(ctx, &player.SetAvatarResponse{Code: stack.ErrCode(err), Message: err.Error()})
		return
	}
	stack.ProtoResponse(ctx, &player.SetAvatarResponse{Code: stack.CodeOK})
}
