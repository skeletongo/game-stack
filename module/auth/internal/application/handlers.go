package application

import (
	"context"
	"strconv"
	"time"

	"github.com/dobyte/due/v2/log"
	"github.com/dobyte/due/v2/utils/xconv"
	dobytejwt "github.com/dobyte/jwt"

	"github.com/skeletongo/game-stack/ddd"
	"github.com/skeletongo/game-stack/internal/component/jwt"
	"github.com/skeletongo/game-stack/module/auth/internal/domain"
	"github.com/skeletongo/game-stack/stack"
)

// PlayerRegistrar 定义 auth 注册流程依赖的玩家创建接口。
type PlayerRegistrar interface {
	CreatePlayer(ctx context.Context, id int64, nickname string) error
	DeletePlayer(ctx context.Context, id int64) error
}

// RegisterHandler 处理注册命令。
type RegisterHandler struct {
	Repo     domain.AccountRepository // 账号仓储
	EventBus *ddd.EventBus            // 领域事件总线
	Players  PlayerRegistrar          // 玩家注册服务
}

// RegisterResult 注册命令的返回结果。
type RegisterResult struct {
	UserID   int64 // 用户id
	PlayerID int64 // 玩家id
}

// Handle 执行注册：校验用户名唯一、创建账户、生成令牌、发布事件。
func (h *RegisterHandler) Handle(ctx context.Context, cmd RegisterCmd) (*RegisterResult, error) {
	if _, err := h.Repo.FindByUsername(ctx, cmd.Username); err == nil {
		return nil, stack.ErrAccountExists
	}
	userID := time.Now().UnixNano()
	playerID := userID
	account, err := domain.NewAccount(userID, playerID, cmd.Username, cmd.Password, cmd.Nickname)
	if err != nil {
		return nil, err
	}
	if err := h.Players.CreatePlayer(ctx, playerID, cmd.Nickname); err != nil {
		return nil, err
	}
	if err := h.Repo.Save(ctx, account); err != nil {
		_ = h.Players.DeletePlayer(ctx, playerID)
		return nil, err
	}
	h.EventBus.Publish(domain.NewAccountCreated(userID, cmd.Username))
	log.Infof("[auth] account registered: uid=%d player_id=%d username=%s", userID, playerID, cmd.Username)
	return &RegisterResult{UserID: userID, PlayerID: playerID}, nil
}

// LoginHandler 处理登录命令。
type LoginHandler struct {
	Repo     domain.AccountRepository // 账号仓储
	EventBus *ddd.EventBus            // 领域事件总线
	Jwt      *jwt.JWT                 // JWT 组件
}

// LoginResult 登录命令的返回结果。
type LoginResult struct {
	UserID    int64  // 用户id
	PlayerID  int64  // 玩家id
	Token     string // 登录令牌
	ExpiresAt int64  // token过期时间
	Nickname  string // 昵称
}

// Handle 执行登录：查找账户、验证密码和封禁状态、生成令牌、发布事件。
func (h *LoginHandler) Handle(ctx context.Context, cmd LoginCmd) (*LoginResult, error) {
	account, err := h.Repo.FindByUsername(ctx, cmd.Username)
	if err != nil {
		return nil, stack.ErrWrongPassword
	}
	if account.IsBanned() {
		return nil, stack.ErrAccountBanned
	}
	if !account.VerifyPassword(cmd.Password) {
		return nil, stack.ErrWrongPassword
	}
	token, err := h.Jwt.GenerateToken(xconv.String(account.ID()), jwt.Payload{jwt.ClaimGameID: cmd.GameID})
	if err != nil {
		return nil, stack.ErrInternalError
	}
	h.EventBus.Publish(domain.NewAccountLoggedIn(account.ID()))
	log.Infof("[auth] account login verified: uid=%d player_id=%d username=%s", account.ID(), account.PlayerID(), cmd.Username)
	return &LoginResult{
		UserID:    account.ID(),
		PlayerID:  account.PlayerID(),
		Token:     token.Token,
		ExpiresAt: token.ExpiredAt.Unix(),
		Nickname:  account.Nickname().String(),
	}, nil
}

// TokenLoginHandler 处理 token 登录长连接命令。
type TokenLoginHandler struct {
	Repo     domain.AccountRepository // 账号仓储
	EventBus *ddd.EventBus            // 领域事件总线
	Jwt      *jwt.JWT                 // JWT 组件
}

// TokenLoginResult token 登录长连接的返回结果。
type TokenLoginResult struct {
	UserID    int64 // 用户id
	PlayerID  int64 // 玩家id
	ExpiresAt int64 // token过期时间
}

// Handle 执行 token 登录：验证 token、加载账户、检查封禁状态。
func (h *TokenLoginHandler) Handle(ctx context.Context, cmd TokenLoginCmd) (*TokenLoginResult, error) {
	payload, err := h.Jwt.ParseToken(cmd.Token)
	if err != nil {
		return nil, tokenError(err)
	}
	uid, err := strconv.ParseInt(payload.Subject(), 10, 64)
	if err != nil || uid == 0 {
		return nil, stack.ErrInvalidToken
	}
	account, err := h.Repo.Load(ctx, uid)
	if err != nil {
		return nil, stack.ErrInvalidToken
	}
	if account.IsBanned() {
		return nil, stack.ErrAccountBanned
	}
	h.EventBus.Publish(domain.NewAccountLoggedIn(account.ID()))
	log.Infof("[auth] token login verified: uid=%d player_id=%d", account.ID(), account.PlayerID())
	return &TokenLoginResult{
		UserID:    account.ID(),
		PlayerID:  account.PlayerID(),
		ExpiresAt: payload.Expired().Unix(),
	}, nil
}

// LogoutHandler 处理登出命令。
type LogoutHandler struct {
	Repo     domain.AccountRepository // 账号仓储
	EventBus *ddd.EventBus            // 领域事件总线
	Jwt      *jwt.JWT                 // JWT 组件
}

// Handle 执行登出：清除账户的令牌，发布事件。
func (h *LogoutHandler) Handle(ctx context.Context, cmd LogoutCmd) (ddd.NoResult, error) {
	if _, err := h.Repo.Load(ctx, cmd.UserID); err != nil {
		return ddd.NoResult{}, err
	}
	if h.Jwt != nil {
		_ = h.Jwt.DestroyTokenBySubject(strconv.FormatInt(cmd.UserID, 10))
	}
	h.EventBus.Publish(domain.NewAccountLoggedOut(cmd.UserID))
	log.Infof("[auth] account logged out: uid=%d", cmd.UserID)
	return ddd.NoResult{}, nil
}

// RefreshTokenHandler 处理令牌刷新命令。
type RefreshTokenHandler struct {
	Repo domain.AccountRepository // 账号仓储
	Jwt  *jwt.JWT                 // JWT 组件
}

// RefreshTokenResult 令牌刷新的返回结果。
type RefreshTokenResult struct {
	Token     string // 新登录令牌
	ExpiresAt int64  // token过期时间
}

// Handle 执行令牌刷新：验证旧令牌并生成新令牌。
func (h *RefreshTokenHandler) Handle(ctx context.Context, cmd RefreshTokenCmd) (*RefreshTokenResult, error) {
	payload, err := h.Jwt.ParseToken(cmd.Token, true)
	if err != nil {
		return nil, tokenError(err)
	}
	if payload.Subject() != strconv.FormatInt(cmd.UserID, 10) {
		return nil, stack.ErrInvalidToken
	}

	if _, err := h.Repo.Load(ctx, cmd.UserID); err != nil {
		return nil, stack.ErrInvalidToken
	}

	newToken, err := h.Jwt.RefreshToken(cmd.Token, true)
	if err != nil {
		return nil, tokenError(err)
	}
	return &RefreshTokenResult{Token: newToken.Token, ExpiresAt: newToken.ExpiredAt.Unix()}, nil
}

func tokenError(err error) error {
	switch {
	case dobytejwt.IsExpiredToken(err):
		return stack.ErrTokenExpired
	case dobytejwt.IsMissingToken(err),
		dobytejwt.IsInvalidToken(err),
		dobytejwt.IsAuthElsewhere(err),
		dobytejwt.IsIdentityMissing(err),
		dobytejwt.IsInvalidSignAlgorithm(err):
		return stack.ErrInvalidToken
	default:
		return err
	}
}
