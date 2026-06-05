package svc

import "context"

// IPlayer 是 player 模块对外暴露的跨模块服务接口。
//
// 其他模块通过 stack.GetService("player") 获取，类型断言为 IPlayer 后调用。
// 所有写操作（Create/Delete/Add/Deduct）均在玩家 Actor 中同步串行执行，
// 保证数据一致性和并发安全。
type IPlayer interface {
	// GetPlayer 查询玩家信息（直接读仓储，不走 Actor）。
	GetPlayer(ctx context.Context, id int64) (*Player, error)
	// CreatePlayer 创建玩家档案（注册账号时由 auth 模块调用）。
	CreatePlayer(ctx context.Context, id int64, nickname string) error
	// DeletePlayer 删除玩家档案（注册补偿或内部清理时调用）。
	DeletePlayer(ctx context.Context, id int64) error
	// AddExp 增加经验值，返回最新经验值。
	AddExp(ctx context.Context, id int64, exp int64) (int64, error)
	// AddGold 增加金币，返回最新金币数量。
	AddGold(ctx context.Context, id int64, gold int32) (int32, error)
	// DeductGold 扣除金币，返回最新金币数量。
	DeductGold(ctx context.Context, id int64, gold int32) (int32, error)
	// AddDiamond 增加钻石，返回最新钻石数量。
	AddDiamond(ctx context.Context, id int64, diamond int32) (int32, error)
	// DeductDiamond 扣除钻石，返回最新钻石数量。
	DeductDiamond(ctx context.Context, id int64, diamond int32) (int32, error)
}

// Player 是 player 模块对外暴露的玩家数据传输对象。
type Player struct {
	ID        int64  // 玩家唯一 ID
	Nickname  string // 昵称
	Level     int32  // 等级
	Exp       int64  // 经验值
	Avatar    string // 头像 URL
	Gold      int32  // 金币数量
	Diamond   int32  // 钻石数量
	CreatedAt int64  // 创建时间戳
	UpdatedAt int64  // 最后更新时间戳
}
