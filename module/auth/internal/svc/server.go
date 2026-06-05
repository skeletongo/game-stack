package svc

import (
	"context"

	"github.com/skeletongo/game-stack/module/auth/internal/domain"
	"github.com/skeletongo/game-stack/module/auth/svc"
)

// server 是 auth 模块对外提供的接口
// 其他模块通过 stack.GetService("auth") 获取，类型断言为 svc.IAuth
type server struct {
	repo domain.AccountRepository
}

func New(repo domain.AccountRepository) svc.IAuth {
	return &server{
		repo: repo,
	}
}

// Authenticate 验证令牌有效性，返回对应的用户 ID。
func (s *server) Authenticate(ctx context.Context, token string) (int64, error) {
	acc, err := s.repo.FindByToken(ctx, token)
	if err != nil {
		return 0, err
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
