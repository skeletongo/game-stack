// Package interfaces 提供 auth 模块的接口层实现。
//
// 职责：
//   - 路由处理器：解析 proto 消息 → 调用应用层 Handler → 构建响应
//   - 事件处理器：Connect/Disconnect 事件 → Actor 生命周期管理
//   - 框架适配：BindGate/BindNode/SpawnPlayer 等 due 框架操作
package interfaces

import (
	"time"

	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/module/actor"
	"github.com/skeletongo/game-stack/module/auth/application"
	"github.com/skeletongo/game-stack/proto/auth"
	"github.com/skeletongo/game-stack/stack"
)

// Handlers 持有路由和事件处理器所需的依赖。
type Handlers struct {
	cleaner  *stack.PlayerDoneCleaner
	proxy    *node.Proxy
	register *application.RegisterHandler
	login    *application.LoginHandler
	logout   *application.LogoutHandler
	refresh  *application.RefreshTokenHandler
}

// NewHandlers 创建处理器。
func NewHandlers(
	cleaner *stack.PlayerDoneCleaner,
	proxy *node.Proxy,
	register *application.RegisterHandler,
	login *application.LoginHandler,
	logout *application.LogoutHandler,
	refresh *application.RefreshTokenHandler,
) *Handlers {
	return &Handlers{
		cleaner: cleaner, proxy: proxy,
		register: register, login: login, logout: logout, refresh: refresh,
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
		ctx.Response(nil)
		return
	}
	if req.Username == "" || req.Password == "" {
		ctx.Response(nil)
		return
	}
	result, err := h.login.Handle(ctx.Context(), application.LoginCmd{Username: req.Username, Password: req.Password})
	if err != nil {
		log.Errorf("[auth] HandleLogin failed: %v", err)
		ctx.Response(nil)
		return
	}
	if err := ctx.BindGate(result.UserID); err != nil {
		log.Errorf("[auth] HandleLogin BindGate failed: %v", err)
		ctx.Response(nil)
		return
	}
	boundNid, err := h.proxy.LocateNode(ctx.Context(), result.UserID, h.proxy.GetName())
	isReconnect := err == nil && boundNid != ""
	if !isReconnect {
		// 首次登录：绑定当前节点 + 创建 Actor
		if err := ctx.BindNode(result.UserID); err != nil {
			log.Errorf("[auth] HandleLogin BindNode failed: %v", err)
			ctx.Response(nil)
			return
		}
		h.cleaner.OnLogin(result.UserID)
		if _, err := actor.SpawnPlayer(h.proxy, result.UserID); err != nil {
			log.Errorf("spawn player actor failed: uid=%d err=%v", result.UserID, err)
		}
	} else if boundNid == h.proxy.GetID() {
		// 重连回原节点：取消清理定时器 + 恢复 Actor
		h.cleaner.OnLogin(result.UserID)
		if act := actor.GetPlayer(h.proxy, result.UserID); act == nil {
			if _, err := actor.SpawnPlayer(h.proxy, result.UserID); err != nil {
				log.Errorf("re-spawn actor failed: uid=%d err=%v", result.UserID, err)
			}
		}
	}
	// 重连且绑定在其他节点：不操作，旧节点通过 Connect 事件感知
	ctx.Response(&auth.LoginResponse{
		Token:     result.Token,
		PlayerId:  result.UserID,
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	})
}

// HandleRegister 处理注册请求（无状态路由）。
func (h *Handlers) HandleRegister(ctx node.Context) {
	log.Infof("[auth] HandleRegister called: uid=%d cid=%d route=%d", ctx.UID(), ctx.CID(), ctx.Route())

	req := &auth.RegisterRequest{}
	if err := ctx.Parse(req); err != nil {
		log.Errorf("[auth] HandleRegister parse failed: %v", err)
		ctx.Response(nil)
		return
	}
	if req.Username == "" || req.Password == "" {
		ctx.Response(nil)
		return
	}
	if req.Nickname == "" {
		req.Nickname = req.Username
	}
	result, err := h.register.Handle(ctx.Context(), application.RegisterCmd{
		Username: req.Username, Password: req.Password, Nickname: req.Nickname,
	})
	if err != nil {
		log.Errorf("[auth] HandleRegister failed: %v", err)
		ctx.Response(nil)
		return
	}
	if err := ctx.BindGate(result.UserID); err != nil {
		log.Errorf("[auth] HandleRegister BindGate failed: %v", err)
		ctx.Response(nil)
		return
	}
	if err := ctx.BindNode(result.UserID); err != nil {
		log.Errorf("[auth] HandleRegister BindNode failed: %v", err)
		ctx.Response(nil)
		return
	}
	h.cleaner.OnLogin(result.UserID)
	if _, err := actor.SpawnPlayer(h.proxy, result.UserID); err != nil {
		log.Errorf("spawn player actor failed: uid=%d err=%v", result.UserID, err)
	}
	ctx.Response(&auth.RegisterResponse{Token: result.Token, PlayerId: result.UserID})
}

// HandleLogout 处理登出请求。
func (h *Handlers) HandleLogout(ctx node.Context) {
	_ = h.logout.Handle(ctx.Context(), application.LogoutCmd{UserID: ctx.UID()})
	ctx.Response(nil)
}

// HandleRefresh 处理令牌刷新请求。
func (h *Handlers) HandleRefresh(ctx node.Context) {
	req := &auth.TokenRefreshRequest{}
	if err := ctx.Parse(req); err != nil {
		log.Errorf("[auth] HandleRefresh parse failed: %v", err)
		ctx.Response(nil)
		return
	}
	result, err := h.refresh.Handle(ctx.Context(), application.RefreshTokenCmd{
		UserID: ctx.UID(), Token: req.Token,
	})
	if err != nil {
		log.Errorf("[auth] HandleRefresh failed: %v", err)
		ctx.Response(nil)
		return
	}
	ctx.Response(&auth.TokenRefreshResponse{Token: result.Token, ExpiresAt: result.ExpiresAt})
}

// HandleConnect 处理集群 Connect 事件。
//
// Connect 事件为全集群广播。旧节点收到后取消玩家的延迟清理定时器，
// 若 Actor 已被杀死则重新创建（断线重连场景）。
func (h *Handlers) HandleConnect(ctx node.Context) {
	uid := ctx.UID()
	if uid == 0 {
		return
	}
	log.Infof("[auth] player connected: uid=%d cid=%d gid=%s", uid, ctx.CID(), ctx.GID())
	h.cleaner.OnLogin(uid)
	if act := actor.GetPlayer(h.proxy, uid); act == nil {
		if _, err := actor.SpawnPlayer(h.proxy, uid); err != nil {
			log.Errorf("[auth] re-spawn actor on connect failed: uid=%d err=%v", uid, err)
		} else {
			log.Infof("[auth] actor re-spawned on reconnect: uid=%d", uid)
		}
	}
}

// HandleDisconnect 处理集群 Disconnect 事件。
//
// 两阶段清理：
//
//	立即：杀 Actor（关闭 mailbox，丢弃排队消息）
//	延迟 30s：Grace Period 到期后清理 token + 解绑节点
func (h *Handlers) HandleDisconnect(ctx node.Context) {
	uid := ctx.UID()
	if uid != 0 {
		actor.KillPlayer(h.proxy, uid)
		h.cleaner.OnDisconnect(uid)
	}
	log.Infof("[auth] player disconnected: uid=%d cid=%d", uid, ctx.CID())
}
