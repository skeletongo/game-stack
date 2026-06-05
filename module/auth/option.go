package auth

import (
	"github.com/skeletongo/game-stack/module/auth/internal/domain"
	"github.com/skeletongo/game-stack/module/auth/internal/infrastructure"
)

// options auth 模块配置。
type options struct {
	repo domain.AccountRepository
}

// defaultOptions 返回默认配置（内存仓储）。
func defaultOptions() *options {
	return &options{
		repo: infrastructure.NewMemoryRepo(),
	}
}

// Option 函数式选项。
type Option func(o *options)

// WithRepository 注入自定义仓储实现（如 Redis/MySQL）。
//
// 使用示例：
//
//	auth.Module(auth.WithRepository(myRedisRepo))
func WithRepository(r domain.AccountRepository) Option {
	return func(o *options) { o.repo = r }
}
