package shop

import "context"

type ShopItem struct {
	ID           int64
	ItemID       int32
	ItemName     string
	Price        int32
	CurrencyType int32
	Discount     int32
	LimitCount   int32
	SoldCount    int32
	LevelRequire int32
	TabIndex     int32
}

type Store interface {
	ListItems(ctx context.Context, tabIndex int32) ([]*ShopItem, error)
	BuyItem(ctx context.Context, uid int64, shopItemID int64, count int32) error
}
