// Package inventory 定义背包道具相关消息类型。
package inventory

// Item 道具信息。
type Item struct {
	ID       int64  `json:"id" msgpack:"id"`
	ItemID   int32  `json:"itemId" msgpack:"itemId"`
	Name     string `json:"name" msgpack:"name"`
	Type     int32  `json:"type" msgpack:"type"` // 1:消耗 2:装备 3:材料
	Count    int32  `json:"count" msgpack:"count"`
	MaxStack int32  `json:"maxStack" msgpack:"maxStack"`
	Level    int32  `json:"level" msgpack:"level"`
	Quality  int32  `json:"quality" msgpack:"quality"` // 品质: 白绿蓝紫橙
	Equipped bool   `json:"equipped" msgpack:"equipped"`
}

// ListRequest 获取背包列表请求。
type ListRequest struct {
	BagType int32 `json:"bagType" msgpack:"bagType"` // 1:材料 2:装备 3:全部
}

// ListResponse 获取背包列表响应。
type ListResponse struct {
	Items     []*Item `json:"items" msgpack:"items"`
	BagSlots  int32   `json:"bagSlots" msgpack:"bagSlots"`
	UsedSlots int32   `json:"usedSlots" msgpack:"usedSlots"`
}

// UseRequest 使用道具请求。
type UseRequest struct {
	ID    int64 `json:"id" msgpack:"id"`
	Count int32 `json:"count" msgpack:"count"`
}

// EquipRequest 装备道具请求。
type EquipRequest struct {
	ID int64 `json:"id" msgpack:"id"`
}

// UnequipRequest 卸下装备请求。
type UnequipRequest struct {
	ID   int64 `json:"id" msgpack:"id"`
	Slot int32 `json:"slot" msgpack:"slot"`
}

// DropRequest 丢弃道具请求。
type DropRequest struct {
	ID    int64 `json:"id" msgpack:"id"`
	Count int32 `json:"count" msgpack:"count"`
}

// SellRequest 出售道具请求。
type SellRequest struct {
	ID    int64 `json:"id" msgpack:"id"`
	Count int32 `json:"count" msgpack:"count"`
}

// SellResponse 出售道具响应。
type SellResponse struct {
	Gold int32 `json:"gold" msgpack:"gold"`
}

// ItemChangeEvent 道具变化事件（服务器推送）。
type ItemChangeEvent struct {
	Item  *Item `json:"item" msgpack:"item"`
	Delta int32 `json:"delta" msgpack:"delta"` // 正数为获得，负数为失去
}
