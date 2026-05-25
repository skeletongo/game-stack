package auth

import (
	"context"
	"time"

	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/module/actor"
	"github.com/skeletongo/game-stack/protocol/auth"
	"github.com/skeletongo/game-stack/stack"
)

type impl struct {
	svc     *service
	cleaner *stack.PlayerDoneCleaner
	proxy   *node.Proxy
}

func newImpl(store Store, cleaner *stack.PlayerDoneCleaner, proxy *node.Proxy) *impl {
	return &impl{svc: newService(store), cleaner: cleaner, proxy: proxy}
}

func (i *impl) handleLogin(ctx node.Context) {
	req := &auth.LoginRequest{}
	if err := ctx.Parse(req); err != nil {
		log.Errorf("parse login request failed: %v", err)
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	if req.Username == "" || req.Password == "" {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	user, err := i.svc.store.GetUserByUsername(context.Background(), req.Username)
	if err != nil {
		stack.RespondError(ctx, stack.ErrWrongPassword)
		return
	}

	if user.BannedAt > 0 {
		stack.RespondError(ctx, stack.ErrAccountBanned)
		return
	}

	if user.Password != hashPassword(req.Password) {
		stack.RespondError(ctx, stack.ErrWrongPassword)
		return
	}

	token := generateToken()
	if err := i.svc.store.SetToken(context.Background(), user.ID, token); err != nil {
		log.Errorf("set token failed: %v", err)
		stack.RespondError(ctx, stack.ErrInternalError)
		return
	}

	if err := ctx.BindGate(user.ID); err != nil {
		log.Errorf("bind gate failed: %v", err)
		stack.RespondError(ctx, stack.ErrInternalError)
		return
	}

	// 检查玩家是否已有节点绑定（断线重连场景）
	boundNid, err := i.proxy.LocateNode(context.Background(), user.ID, i.proxy.GetName())
	if err != nil {
		log.Errorf("locate node failed: %v", err)
		stack.RespondError(ctx, stack.ErrInternalError)
		return
	}
	isReconnect := boundNid != ""

	if !isReconnect {
		// 首次登录：绑定当前节点
		if err := ctx.BindNode(user.ID); err != nil {
			log.Errorf("bind node failed: %v", err)
			stack.RespondError(ctx, stack.ErrInternalError)
			return
		}
		i.cleaner.OnLogin(user.ID)
		if _, err := actor.SpawnPlayer(i.proxy, user.ID); err != nil {
			log.Errorf("spawn player actor failed: %v", err)
		}
	} else if boundNid == i.proxy.GetID() {
		// 重连回到原节点：取消清理 + 恢复 Actor
		i.cleaner.OnLogin(user.ID)
		if act := actor.GetPlayer(i.proxy, user.ID); act == nil {
			if _, err := actor.SpawnPlayer(i.proxy, user.ID); err != nil {
				log.Errorf("re-spawn actor failed: uid=%d err=%v", user.ID, err)
			}
		}
	}
	// 重连且绑定在其他节点：不操作，旧节点通过 cluster.Connect 事件处理后续

	stack.RespondData(ctx, &auth.LoginResponse{
		Token:       token,
		PlayerID:    user.ID,
		ExpiresAt:   time.Now().Add(24 * time.Hour).Unix(),
		IsNewPlayer: false,
	})
}

func (i *impl) handleRegister(ctx node.Context) {
	req := &auth.RegisterRequest{}
	if err := ctx.Parse(req); err != nil {
		log.Errorf("parse register request failed: %v", err)
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	if req.Username == "" || req.Password == "" {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	if req.Nickname == "" {
		req.Nickname = req.Username
	}

	if _, err := i.svc.store.GetUserByUsername(context.Background(), req.Username); err == nil {
		stack.RespondError(ctx, stack.ErrAccountExists)
		return
	}

	user := &User{
		ID:        time.Now().UnixNano(),
		Username:  req.Username,
		Password:  hashPassword(req.Password),
		Nickname:  req.Nickname,
		CreatedAt: time.Now().Unix(),
	}

	if err := i.svc.store.CreateUser(context.Background(), user); err != nil {
		log.Errorf("create user failed: %v", err)
		stack.RespondError(ctx, stack.ErrInternalError)
		return
	}

	token := generateToken()
	if err := i.svc.store.SetToken(context.Background(), user.ID, token); err != nil {
		log.Errorf("set token failed: %v", err)
		stack.RespondError(ctx, stack.ErrInternalError)
		return
	}

	if err := ctx.BindGate(user.ID); err != nil {
		log.Errorf("bind gate failed: %v", err)
		stack.RespondError(ctx, stack.ErrInternalError)
		return
	}

	if err := ctx.BindNode(user.ID); err != nil {
		log.Errorf("bind node failed: %v", err)
		stack.RespondError(ctx, stack.ErrInternalError)
		return
	}

	i.cleaner.OnLogin(user.ID)

	if _, err := actor.SpawnPlayer(i.proxy, user.ID); err != nil {
		log.Errorf("spawn player actor failed: %v", err)
	}

	stack.RespondData(ctx, &auth.RegisterResponse{
		Token:    token,
		PlayerID: user.ID,
	})
}

func (i *impl) handleLogout(ctx node.Context) {
	_ = i.svc.store.DeleteToken(context.Background(), ctx.UID())
	stack.RespondOK(ctx)
}

func (i *impl) handleRefresh(ctx node.Context) {
	req := &auth.TokenRefreshRequest{}
	if err := ctx.Parse(req); err != nil {
		stack.RespondError(ctx, stack.ErrInvalidParam)
		return
	}

	uid, err := i.svc.Authenticate(req.Token)
	if err != nil || uid != ctx.UID() {
		stack.RespondError(ctx, stack.ErrInvalidToken)
		return
	}

	token := generateToken()
	if err := i.svc.store.SetToken(context.Background(), uid, token); err != nil {
		stack.RespondError(ctx, stack.ErrInternalError)
		return
	}

	stack.RespondData(ctx, &auth.TokenRefreshResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	})
}

// handleConnect 连接事件处理器，不可调用 ctx.Response。
// cluster.Connect 为全集群广播，旧节点收到后取消延迟清理、恢复 Actor。
func (i *impl) handleConnect(ctx node.Context) {
	uid := ctx.UID()
	if uid == 0 {
		return
	}
	log.Infof("[auth] player connected: uid=%d cid=%d gid=%s", uid, ctx.CID(), ctx.GID())

	// 取消延迟清理（如果有待处理的 cleaner 定时器）
	i.cleaner.OnLogin(uid)

	// 如果 Actor 已被杀死（断线场景），重新创建
	if act := actor.GetPlayer(i.proxy, uid); act == nil {
		if _, err := actor.SpawnPlayer(i.proxy, uid); err != nil {
			log.Errorf("[auth] re-spawn actor on connect failed: uid=%d err=%v", uid, err)
		} else {
			log.Infof("[auth] actor re-spawned on reconnect: uid=%d", uid)
		}
	}
}

// handleDisconnect 断开事件处理器，不可调用 ctx.Response。
//
// 两阶段清理策略：
//   - 立即：杀 Actor（关闭 mailbox，丢弃排队消息）
//   - 延迟（30s Grace Period）：清除 token + 解绑节点 + 清理模块数据
//
// Grace Period 期间保留 token 和节点绑定，玩家重连可无缝恢复。
func (i *impl) handleDisconnect(ctx node.Context) {
	uid := ctx.UID()
	if uid != 0 {
		_ = i.svc.store.SetOffline(context.Background(), uid)
		actor.KillPlayer(i.proxy, uid)
		i.cleaner.OnDisconnect(uid)
	}
	log.Infof("[auth] player disconnected: uid=%d cid=%d", uid, ctx.CID())
}
