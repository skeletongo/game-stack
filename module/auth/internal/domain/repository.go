package domain

import (
	"context"

	"github.com/skeletongo/game-stack/ddd"
)

// AccountRepository 是 Account 聚合的仓储接口。
// 定义在领域层，实现在基础设施层。
type AccountRepository interface {
	ddd.Repository[*Account]

	// FindByUsername 按用户名查找账户。
	FindByUsername(ctx context.Context, username string) (*Account, error)
}
