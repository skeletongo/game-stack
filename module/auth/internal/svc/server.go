package svc

import (
	"context"
	"strconv"

	dobytejwt "github.com/dobyte/jwt"

	"github.com/skeletongo/game-stack/component/jwt"
	"github.com/skeletongo/game-stack/module/auth/internal/domain"
	"github.com/skeletongo/game-stack/module/auth/svc"
	"github.com/skeletongo/game-stack/stack"
)

// server 是 auth 模块对外提供的接口
// 其他模块通过 stack.GetService("auth") 获取，类型断言为 svc.IAuth
type server struct {
	repo domain.AccountRepository
	jwt  *jwt.JWT
}

func New(repo domain.AccountRepository, jt *jwt.JWT) svc.IAuth {
	return &server{
		repo: repo,
		jwt:  jt,
	}
}

// Authenticate 验证令牌有效性，返回对应的用户 ID。
func (s *server) Authenticate(ctx context.Context, token string) (int64, error) {
	payload, err := s.jwt.ParseToken(token)
	if err != nil {
		return 0, tokenError(err)
	}
	uid, err := strconv.ParseInt(payload.Subject(), 10, 64)
	if err != nil {
		return 0, stack.ErrInvalidToken
	}
	acc, err := s.repo.Load(ctx, uid)
	if err != nil {
		return 0, stack.ErrInvalidToken
	}
	if acc.Token().String() != token {
		return 0, stack.ErrInvalidToken
	}
	return acc.ID(), nil
}

// IsOnline 检查用户是否在线（有活跃的 Gate 连接）。
func (s *server) IsOnline(ctx context.Context, uid int64) bool {
	acc, err := s.repo.Load(ctx, uid)
	if err != nil {
		return false
	}
	return acc.IsOnline()
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
