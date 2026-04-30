package inventory

import "context"

// Service 背包模块对外服务接口。
type Service interface {
	// HasItem 检查玩家是否拥有某道具。
	HasItem(uid int64, itemID int64) bool
	// GetItemCount 获取玩家某道具数量。
	GetItemCount(uid int64, itemID int64) int32
}

type service struct {
	store Store
}

func newService(store Store) *service {
	return &service{store: store}
}

func (s *service) HasItem(uid int64, itemID int64) bool {
	items, err := s.store.ListItems(context.Background(), uid, 0)
	if err != nil {
		return false
	}
	for _, it := range items {
		if it.ID == itemID {
			return true
		}
	}
	return false
}

func (s *service) GetItemCount(uid int64, itemID int64) int32 {
	items, err := s.store.ListItems(context.Background(), uid, 0)
	if err != nil {
		return 0
	}
	for _, it := range items {
		if it.ID == itemID {
			return it.Count
		}
	}
	return 0
}
