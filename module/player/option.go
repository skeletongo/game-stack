package player

import (
	"github.com/skeletongo/game-stack/module/player/domain"
	"github.com/skeletongo/game-stack/module/player/infrastructure"
)

// options Player 模块配置。
type options struct {
	repo domain.PlayerRepository
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
//	player.Module(player.WithRepository(myRedisRepo))
func WithRepository(r domain.PlayerRepository) Option {
	return func(o *options) { o.repo = r }
}
