package server

import (
	"context"
	"strconv"

	"github.com/dobyte/due/v2/cluster/node"
	dobytejwt "github.com/dobyte/jwt"

	"github.com/skeletongo/game-stack/internal/component/jwt"
	"github.com/skeletongo/game-stack/module/auth/internal/domain"
	"github.com/skeletongo/game-stack/module/auth/svc"
	"github.com/skeletongo/game-stack/stack"
)

// server 是 auth 模块对外提供的接口
// 其他模块通过 stack.GetService("auth") 获取，类型断言为 svc.IAuth
type server struct {
	repo  domain.AccountRepository
	jwt   *jwt.JWT
	proxy *node.Proxy
}

func New(repo domain.AccountRepository, jt *jwt.JWT, proxy *node.Proxy) svc.IAuth {
	return &server{
		repo:  repo,
		jwt:   jt,
		proxy: proxy,
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
	if _, err := s.repo.Load(ctx, uid); err != nil {
		return 0, stack.ErrInvalidToken
	}
	return uid, nil
}

// IsOnline 检查用户是否在线（有活跃的 Gate 连接）。
func (s *server) IsOnline(ctx context.Context, uid int64) bool {
	if s.proxy == nil {
		return false
	}
	gid, err := s.proxy.LocateGate(ctx, uid)
	if err != nil {
		return false
	}
	return gid != ""
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
