// Package shop 定义商城相关消息类型。
package shop

// ShopItem 商城道具。
type ShopItem struct {
	ID           int64  `json:"id" msgpack:"id"`
	ItemID       int32  `json:"itemId" msgpack:"itemId"`
	ItemName     string `json:"itemName" msgpack:"itemName"`
	Price        int32  `json:"price" msgpack:"price"`
	CurrencyType int32  `json:"currencyType" msgpack:"currencyType"` // 1:金币 2:钻石
	Discount     int32  `json:"discount" msgpack:"discount"`         // 折扣百分比
	LimitCount   int32  `json:"limitCount" msgpack:"limitCount"`     // 限购数量(-1无限)
	SoldCount    int32  `json:"soldCount" msgpack:"soldCount"`
	LevelRequire int32  `json:"levelRequire" msgpack:"levelRequire"`
	TabIndex     int32  `json:"tabIndex" msgpack:"tabIndex"`
}

// ListRequest 获取商城列表请求。
type ListRequest struct {
	TabIndex int32 `json:"tabIndex" msgpack:"tabIndex"` // -1:全部
}

// ListResponse 获取商城列表响应。
type ListResponse struct {
	Items    []*ShopItem `json:"items" msgpack:"items"`
	TabIndex int32       `json:"tabIndex" msgpack:"tabIndex"`
}

// BuyRequest 购买请求。
type BuyRequest struct {
	ShopItemID int64 `json:"shopItemId" msgpack:"shopItemId"`
	Count      int32 `json:"count" msgpack:"count"`
}

// BuyResponse 购买响应。
type BuyResponse struct {
	ItemID  int32 `json:"itemId" msgpack:"itemId"`
	Count   int32 `json:"count" msgpack:"count"`
	Cost    int32 `json:"cost" msgpack:"cost"`
	Gold    int32 `json:"gold" msgpack:"gold"`
	Diamond int32 `json:"diamond" msgpack:"diamond"`
}

// RefreshRequest 刷新商城请求。
type RefreshRequest struct {
	TabIndex int32 `json:"tabIndex" msgpack:"tabIndex"`
}
