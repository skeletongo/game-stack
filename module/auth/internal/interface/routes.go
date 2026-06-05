// Package interfaces 提供 auth 模块的接口层实现。
//
// 职责：
//   - 路由处理器：解析 proto 消息 → 通过 CommandBus 分发 → 构建响应
//   - 框架适配：BindGate/BindNode 等 due 框架操作
//
// 注意：Connect/Disconnect 事件已移至 module/clean 统一处理。
package interfaces

import (
	"time"

	"github.com/dobyte/due/v2/cluster"
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"
	"github.com/dobyte/due/v2/session"
	"github.com/skeletongo/game-stack/ddd"

	"github.com/skeletongo/game-stack/module/auth/internal/application"
	"github.com/skeletongo/game-stack/proto/auth"
	"github.com/skeletongo/game-stack/proto/common"
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

// HandleLogin 处理登录请求（无状态路由）。
//
// 流程：验证凭证 → 网关绑定 → 检查重连状态 → 节点绑定/Actor创建 → 响应。
func (h *Handlers) HandleLogin(ctx node.Context) {
	log.Infof("[auth] HandleLogin called: uid=%d cid=%d", ctx.UID(), ctx.CID())

	req := &auth.LoginRequest{}
	if err := ctx.Parse(req); err != nil {
		log.Errorf("[auth] HandleLogin parse failed: %v", err)
		stack.ProtoResponse(ctx, &auth.LoginResponse{Code: int32(common.SysError_INVALID_PARAM), Message: err.Error()})
		return
	}
	if req.Username == "" || req.Password == "" {
		stack.ProtoResponse(ctx, &auth.LoginResponse{Code: int32(common.SysError_INVALID_PARAM), Message: "username and password required"})
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
		stack.ProtoResponse(ctx, &auth.LoginResponse{Code: int32(common.SysError_INTERNAL_ERROR), Message: err.Error()})
		return
	}
	boundNid, err := h.proxy.LocateNode(ctx.Context(), result.UserID, h.proxy.GetName())
	isReconnect := err == nil && boundNid != ""
	if !isReconnect {
		// 首次登录：绑定当前节点 + 创建 Actor
		if err := ctx.BindNode(result.UserID); err != nil {
			log.Errorf("[auth] HandleLogin BindNode failed: %v", err)
			stack.ProtoResponse(ctx, &auth.LoginResponse{Code: int32(common.SysError_INTERNAL_ERROR), Message: err.Error()})
			return
		}
	}
	// 重连且绑定在其他节点：不操作，旧节点通过 Connect 事件感知
	stack.ProtoResponse(ctx, &auth.LoginResponse{
		Code:      stack.CodeOK,
		Token:     result.Token,
		PlayerId:  result.PlayerID,
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	})
}

// HandleRegister 处理注册请求（无状态路由）。
func (h *Handlers) HandleRegister(ctx node.Context) {
	log.Infof("[auth] HandleRegister called: uid=%d cid=%d route=%d", ctx.UID(), ctx.CID(), ctx.Route())

	req := &auth.RegisterRequest{}
	if err := ctx.Parse(req); err != nil {
		log.Errorf("[auth] HandleRegister parse failed: %v", err)
		stack.ProtoResponse(ctx, &auth.RegisterResponse{Code: int32(common.SysError_INVALID_PARAM), Message: err.Error()})
		return
	}
	if req.Username == "" || req.Password == "" {
		stack.ProtoResponse(ctx, &auth.RegisterResponse{Code: int32(common.SysError_INVALID_PARAM), Message: "username and password required"})
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
	if err := ctx.BindGate(result.UserID); err != nil {
		log.Errorf("[auth] HandleRegister BindGate failed: %v", err)
		stack.ProtoResponse(ctx, &auth.RegisterResponse{Code: int32(common.SysError_INTERNAL_ERROR), Message: err.Error()})
		return
	}
	if err := ctx.BindNode(result.UserID); err != nil {
		log.Errorf("[auth] HandleRegister BindNode failed: %v", err)
		stack.ProtoResponse(ctx, &auth.RegisterResponse{Code: int32(common.SysError_INTERNAL_ERROR), Message: err.Error()})
		return
	}
	stack.ProtoResponse(ctx, &auth.RegisterResponse{Code: stack.CodeOK, Token: result.Token, PlayerId: result.PlayerID})
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
		stack.ProtoResponse(ctx, &auth.TokenRefreshResponse{Code: int32(common.SysError_INVALID_PARAM), Message: err.Error()})
		return
	}
	result, err := ddd.Dispatch[*application.RefreshTokenResult](ctx.Context(), h.cmdBus, application.RefreshTokenCmd{UserID: ctx.UID(), Token: req.Token})
	if err != nil {
		log.Errorf("[auth] HandleRefresh failed: %v", err)
		stack.ProtoResponse(ctx, &auth.TokenRefreshResponse{Code: stack.ErrCode(err), Message: err.Error()})
		return
	}
	stack.ProtoResponse(ctx, &auth.TokenRefreshResponse{Code: stack.CodeOK, Token: result.Token})
}
