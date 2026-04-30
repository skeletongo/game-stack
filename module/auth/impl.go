package auth

import (
	"context"
	"time"

	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/protocol/auth"
	"github.com/skeletongo/game-stack/stack"
)

type impl struct {
	svc *service
}

func newImpl(store Store) *impl {
	return &impl{svc: newService(store)}
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

	if err := ctx.BindNode(user.ID); err != nil {
		log.Errorf("bind node failed: %v", err)
		stack.RespondError(ctx, stack.ErrInternalError)
		return
	}

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
func (i *impl) handleConnect(ctx node.Context) {
	log.Infof("[auth] player connected: uid=%d cid=%d gid=%s", ctx.UID(), ctx.CID(), ctx.GID())
}

// handleDisconnect 断开事件处理器，不可调用 ctx.Response。
func (i *impl) handleDisconnect(ctx node.Context) {
	uid := ctx.UID()
	if uid != 0 {
		_ = i.svc.store.DeleteToken(context.Background(), uid)
		_ = i.svc.store.SetOffline(context.Background(), uid)
	}
	log.Infof("[auth] player disconnected: uid=%d cid=%d", uid, ctx.CID())
}
