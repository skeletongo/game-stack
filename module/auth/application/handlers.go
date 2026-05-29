package application

import (
	"context"
	"time"

	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/ddd"
	"github.com/skeletongo/game-stack/module/auth/domain"
)

// RegisterHandler 处理注册命令。
type RegisterHandler struct {
	Repo     domain.AccountRepository
	EventBus *ddd.EventBus
}

// RegisterResult 注册命令的返回结果。
type RegisterResult struct {
	UserID int64
	Token  string
}

// Handle 执行注册：校验用户名唯一、创建账户、生成令牌、发布事件。
func (h *RegisterHandler) Handle(ctx context.Context, cmd RegisterCmd) (*RegisterResult, error) {
	if _, err := h.Repo.FindByUsername(ctx, cmd.Username); err == nil {
		return nil, domain.ErrAccountExists
	}
	userID := time.Now().UnixNano()
	account, err := domain.NewAccount(userID, cmd.Username, cmd.Password, cmd.Nickname)
	if err != nil {
		return nil, err
	}
	token := domain.GenerateToken()
	if err := account.Login(token, ""); err != nil {
		return nil, err
	}
	if err := h.Repo.Save(ctx, account); err != nil {
		return nil, err
	}
	h.EventBus.Publish(domain.NewAccountCreated(userID, cmd.Username))
	h.EventBus.Publish(domain.NewAccountLoggedIn(userID))
	log.Infof("[auth] account registered: uid=%d username=%s", userID, cmd.Username)
	return &RegisterResult{UserID: userID, Token: token.String()}, nil
}

// LoginHandler 处理登录命令。
type LoginHandler struct {
	Repo     domain.AccountRepository
	EventBus *ddd.EventBus
}

// LoginResult 登录命令的返回结果。
type LoginResult struct {
	UserID   int64
	Token    string
	Nickname string
}

// Handle 执行登录：查找账户、验证密码和封禁状态、生成令牌、发布事件。
func (h *LoginHandler) Handle(ctx context.Context, cmd LoginCmd) (*LoginResult, error) {
	account, err := h.Repo.FindByUsername(ctx, cmd.Username)
	if err != nil {
		return nil, domain.ErrWrongPassword
	}
	if account.IsBanned() {
		return nil, domain.ErrAccountBanned
	}
	if !account.VerifyPassword(cmd.Password) {
		return nil, domain.ErrWrongPassword
	}
	token := domain.GenerateToken()
	if err := account.Login(token, ""); err != nil {
		return nil, err
	}
	if err := h.Repo.Save(ctx, account); err != nil {
		return nil, err
	}
	h.EventBus.Publish(domain.NewAccountLoggedIn(account.ID()))
	log.Infof("[auth] account logged in: uid=%d username=%s", account.ID(), cmd.Username)
	return &LoginResult{UserID: account.ID(), Token: token.String(), Nickname: account.Nickname().String()}, nil
}

// LogoutHandler 处理登出命令。
type LogoutHandler struct {
	Repo     domain.AccountRepository
	EventBus *ddd.EventBus
}

// Handle 执行登出：清除账户的令牌和在线状态，发布事件。
func (h *LogoutHandler) Handle(ctx context.Context, cmd LogoutCmd) error {
	account, err := h.Repo.Load(ctx, cmd.UserID)
	if err != nil {
		return err
	}
	account.Logout()
	if err := h.Repo.Save(ctx, account); err != nil {
		return err
	}
	h.EventBus.Publish(domain.NewAccountLoggedOut(cmd.UserID))
	log.Infof("[auth] account logged out: uid=%d", cmd.UserID)
	return nil
}

// RefreshTokenHandler 处理令牌刷新命令。
type RefreshTokenHandler struct {
	Repo domain.AccountRepository
}

// RefreshTokenResult 令牌刷新的返回结果。
type RefreshTokenResult struct {
	Token     string
	ExpiresAt int64
}

// Handle 执行令牌刷新：验证旧令牌、生成新令牌、持久化。
func (h *RefreshTokenHandler) Handle(ctx context.Context, cmd RefreshTokenCmd) (*RefreshTokenResult, error) {
	account, err := h.Repo.FindByToken(ctx, cmd.Token)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}
	if account.ID() != cmd.UserID {
		return nil, domain.ErrInvalidToken
	}
	newToken := domain.GenerateToken()
	account.RefreshToken(newToken)
	if err := h.Repo.Save(ctx, account); err != nil {
		return nil, err
	}
	return &RefreshTokenResult{Token: newToken.String(), ExpiresAt: time.Now().Add(24 * time.Hour).Unix()}, nil
}
