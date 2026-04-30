package player

import "context"

// Player 玩家数据。
type Player struct {
	ID        int64
	Nickname  string
	Level     int32
	Exp       int64
	Avatar    string
	Gold      int32
	Diamond   int32
	CreatedAt int64
	UpdatedAt int64
}

// Store 定义玩家模块的数据存储接口。
type Store interface {
	CreatePlayer(ctx context.Context, player *Player) error
	GetPlayer(ctx context.Context, id int64) (*Player, error)
	UpdatePlayer(ctx context.Context, player *Player) error
	GetPlayerByName(ctx context.Context, name string) (*Player, error)
	// RemovePlayer 删除玩家内存数据（断线清理用）。
	RemovePlayer(ctx context.Context, id int64) error
}
