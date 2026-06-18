// Package interfaces 提供 auth 模块的接口层实现。
//
// 职责：
//   - 路由处理器：解析 proto 消息 → 通过 CommandBus 分发 → 构建响应
//   - 框架适配：BindGate/BindNode 等 due 框架操作
//
// 注意：Connect/Disconnect 事件已移至 module/playerlife 统一处理。
package interfaces

import (
	"time"

	"github.com/dobyte/due/v2/cluster"
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"
	"github.com/dobyte/due/v2/session"

	"github.com/skeletongo/game-stack/ddd"
	"github.com/skeletongo/game-stack/module/auth/internal/application"
	"github.com/skeletongo/game-stack/module/playerlife"
	"github.com/skeletongo/game-stack/proto/auth"
	"github.com/skeletongo/game-stack/stack"
)

// Handlers 持有路由处理器所需的依赖。
type Handlers struct {
	proxy  *node.Proxy
	cmdBus *ddd.CommandBus
}

// NewHandlers 创建处理器。
func NewHandlers(proxy *node.Proxy, cmdBus *ddd.CommandBus) *Handlers {
	return &Handlers{
		proxy:  proxy,
		cmdBus: cmdBus,
	}
}

// HandleRegister 处理注册请求（无状态路由）。
func (h *Handlers) HandleRegister(ctx node.Context) {
	log.Infof("[auth] HandleRegister called: uid=%d cid=%d route=%d", ctx.UID(), ctx.CID(), ctx.Route())

	req := &auth.RegisterRequest{}
	if err := ctx.Parse(req); err != nil {
		log.Errorf("[auth] HandleRegister parse failed: %v", err)
		stack.ProtoResponse(ctx, &auth.RegisterResponse{Code: stack.ErrInvalidParam.Code, Message: err.Error()})
		return
	}
	if req.Username == "" || req.Password == "" {
		stack.ProtoResponse(ctx, &auth.RegisterResponse{Code: stack.ErrInvalidParam.Code, Message: stack.ErrInvalidParam.Message})
		return
	}
	if req.Nickname == "" {
		req.Nickname = req.Username
	}

	result, err := ddd.Dispatch[*application.RegisterResult](ctx.Context(), h.cmdBus,
		application.RegisterCmd{Username: req.Username, Password: req.Password, Nickname: req.Nickname})
	if err != nil {
		log.Errorf("[auth] HandleRegister failed: %v", err)
		stack.ProtoResponse(ctx, &auth.RegisterResponse{Code: stack.ErrCode(err), Message: err.Error()})
		return
	}

	if result.UserID > 0 {
		stack.ProtoResponse(ctx, &auth.RegisterResponse{Code: stack.CodeOK})
		return
	}

	stack.ProtoResponse(ctx, &auth.RegisterResponse{Code: stack.ErrInternalError.Code, Message: stack.ErrInternalError.Message})
}

func (h *Handlers) markOnline(ctx node.Context, uid int64, token string) error {
	_, err := h.cmdBus.Dispatch(ctx.Context(), application.MarkOnlineCmd{UserID: uid, Token: token, GID: ctx.GID()})
	if err != nil {
		log.Errorf("[auth] mark online failed: uid=%d gid=%s err=%v", uid, ctx.GID(), err)
	}
	return err
}

// HandleLogin 处理登录请求（无状态路由）。
//
// 流程：验证凭证 → 网关绑定 → 检查重连状态 → 节点绑定/Actor创建 → 响应。
func (h *Handlers) HandleLogin(ctx node.Context) {
	log.Infof("[auth] HandleLogin called: uid=%d cid=%d", ctx.UID(), ctx.CID())

	req := &auth.LoginRequest{}
	if err := ctx.Parse(req); err != nil {
		log.Errorf("[auth] HandleLogin parse failed: %v", err)
		stack.ProtoResponse(ctx, &auth.LoginResponse{Code: stack.ErrInvalidParam.Code, Message: stack.ErrInvalidParam.Message})
		return
	}
	if req.Username == "" || req.Password == "" {
		stack.ProtoResponse(ctx, &auth.LoginResponse{Code: stack.ErrInvalidParam.Code, Message: "username and password required"})
		return
	}

	result, err := ddd.Dispatch[*application.LoginResult](ctx.Context(), h.cmdBus,
		application.LoginCmd{Username: req.Username, Password: req.Password})
	if err != nil {
		log.Errorf("[auth] HandleLogin failed: %v", err)
		stack.ProtoResponse(ctx, &auth.LoginResponse{Code: stack.ErrCode(err), Message: err.Error()})
		return
	}

	// 顶号：若用户已有活跃连接，推送踢出消息并断开旧连接
	if gid, err := h.proxy.LocateGate(ctx.Context(), result.UserID); err == nil && gid != "" {
		log.Infof("[auth] kicking old session: uid=%d gid=%s", result.UserID, gid)
		_ = h.proxy.Push(ctx.Context(), &cluster.PushArgs{
			Kind:       session.User,
			Target:     result.UserID,
			Message:    &cluster.Message{Route: stack.RouteAuthKick, Data: auth.KickResponse{Reason: auth.KickReason_LoginElseWhere}},
			Disconnect: true,
		})
	}

	if err := ctx.BindGate(result.UserID); err != nil {
		log.Errorf("[auth] HandleLogin BindGate failed: %v", err)
		stack.ProtoResponse(ctx, &auth.LoginResponse{Code: stack.ErrInternalError.Code, Message: stack.ErrInternalError.Message})
		return
	}

	boundNid, err := h.proxy.LocateNode(ctx.Context(), result.UserID, h.proxy.GetName())
	if err != nil {
		log.Errorf("[auth] HandleLogin LocateNode failed: %v", err)
		stack.ProtoResponse(ctx, &auth.LoginResponse{Code: stack.ErrInternalError.Code, Message: stack.ErrInternalError.Message})
		return
	}
	isReconnect := boundNid != ""
	if !isReconnect {
		// 首次登录：绑定当前节点 + 创建 Actor
		if err := ctx.BindNode(result.UserID); err != nil {
			log.Errorf("[auth] HandleLogin BindNode failed: %v", err)
			stack.ProtoResponse(ctx, &auth.LoginResponse{Code: stack.ErrInternalError.Code, Message: stack.ErrInternalError.Message})
			return
		}
		if err := h.markOnline(ctx, result.UserID, result.Token); err != nil {
			stack.ProtoResponse(ctx, &auth.LoginResponse{Code: stack.ErrInternalError.Code, Message: stack.ErrInternalError.Message})
			return
		}
		playerlife.Get().OnLogin(ctx.Context(), result.UserID)
	} else if boundNid == h.proxy.GetID() {
		if err := h.markOnline(ctx, result.UserID, result.Token); err != nil {
			stack.ProtoResponse(ctx, &auth.LoginResponse{Code: stack.ErrInternalError.Code, Message: stack.ErrInternalError.Message})
			return
		}
		playerlife.Get().OnLogin(ctx.Context(), result.UserID)
	} else {
		if err := h.markOnline(ctx, result.UserID, result.Token); err != nil {
			stack.ProtoResponse(ctx, &auth.LoginResponse{Code: stack.ErrInternalError.Code, Message: stack.ErrInternalError.Message})
			return
		}
	}

	// 重连且绑定在其他节点：不操作，旧节点通过 Connect 事件感知
	stack.ProtoResponse(ctx, &auth.LoginResponse{
		Code:      stack.CodeOK,
		Token:     result.Token,
		PlayerId:  result.PlayerID,
		ExpiresAt: result.ExpiresAt,
		UnixNano:  time.Now().UnixNano(),
	})
}

// HandleLogout 处理登出请求。
func (h *Handlers) HandleLogout(ctx node.Context) {
	_, _ = h.cmdBus.Dispatch(ctx.Context(), application.LogoutCmd{UserID: ctx.UID()})
	_ = ctx.Response(nil)
}

// HandleRefresh 处理令牌刷新请求。
func (h *Handlers) HandleRefresh(ctx node.Context) {
	req := &auth.TokenRefreshRequest{}
	if err := ctx.Parse(req); err != nil {
		log.Errorf("[auth] HandleRefresh parse failed: %v", err)
		stack.ProtoResponse(ctx, &auth.TokenRefreshResponse{Code: stack.ErrInvalidParam.Code, Message: err.Error()})
		return
	}
	result, err := ddd.Dispatch[*application.RefreshTokenResult](ctx.Context(), h.cmdBus, application.RefreshTokenCmd{UserID: ctx.UID(), Token: req.Token})
	if err != nil {
		log.Errorf("[auth] HandleRefresh failed: %v", err)
		stack.ProtoResponse(ctx, &auth.TokenRefreshResponse{Code: stack.ErrCode(err), Message: err.Error()})
		return
	}
	stack.ProtoResponse(ctx, &auth.TokenRefreshResponse{Code: stack.CodeOK, Token: result.Token, ExpiresAt: result.ExpiresAt})
}
