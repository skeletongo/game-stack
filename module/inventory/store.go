package inventory

import "context"

// Item 道具数据。
type Item struct {
	ID       int64
	ItemID   int32
	Name     string
	Type     int32
	Count    int32
	MaxStack int32
	Level    int32
	Quality  int32
	Equipped bool
}

// Store 背包模块数据存储接口。
type Store interface {
	ListItems(ctx context.Context, uid int64, bagType int32) ([]*Item, error)
	AddItem(ctx context.Context, uid int64, item *Item) error
	RemoveItem(ctx context.Context, uid int64, itemID int64, count int32) error
	UseItem(ctx context.Context, uid int64, itemID int64, count int32) error
	EquipItem(ctx context.Context, uid int64, itemID int64) error
	UnequipItem(ctx context.Context, uid int64, itemID int64) error
	GetBagSize(ctx context.Context, uid int64) (int32, error)
	// RemovePlayerBag 删除玩家背包数据（断线清理用）。
	RemovePlayerBag(ctx context.Context, uid int64) error
}
