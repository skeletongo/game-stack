package auth

import (
	"context"

	"github.com/skeletongo/game-stack/module/auth/domain"
)

// ---- 框架适配器 ----
//
// 适配器是模块与框架之间的胶水层：
//   - cleanableAdapter → 适配 Repository 到框架的生命周期接口（CleanableService）
//   - svcAdapter       → 适配 Repository 到框架的服务注册接口（跨模块调用）

// cleanableAdapter 适配 AccountRepository → stack.CleanableService。
type cleanableAdapter struct {
	repo domain.AccountRepository
}

func (a *cleanableAdapter) CleanPlayerData(uid int64) error {
	return a.repo.Delete(context.Background(), uid)
}

// svcAdapter 是 auth 模块对外的服务适配器。
// 其他模块通过 stack.GetService("auth") 获取，类型断言为 *svcAdapter。
//
// 提供的能力：
//   - Authenticate(token) → 验证令牌并返回 userID
//   - IsOnline(uid) → 检查用户是否在线
type svcAdapter struct {
	repo domain.AccountRepository
}

// Authenticate 验证令牌有效性，返回对应的用户 ID。
func (s *svcAdapter) Authenticate(token string) (int64, error) {
	acc, err := s.repo.FindByToken(context.Background(), token)
	if err != nil {
		return 0, err
	}
	return acc.ID(), nil
}

// IsOnline 检查用户是否在线（有活跃的 Gate 连接）。
func (s *svcAdapter) IsOnline(uid int64) bool {
	acc, err := s.repo.Load(context.Background(), uid)
	if err != nil {
		return false
	}
	return acc.IsOnline()
}
