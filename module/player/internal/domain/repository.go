package domain

import (
	"context"

	"github.com/skeletongo/game-stack/ddd"
)

// PlayerRepository 是 Player 聚合的仓储接口。
// 定义在领域层，实现在基础设施层。
type PlayerRepository interface {
	ddd.Repository[*Player]

	// FindByNickname 按昵称查找玩家。
	FindByNickname(ctx context.Context, nickname string) (*Player, error)
}
