package shop

import (
	"context"
	"fmt"
	"sync"
)

type memoryStore struct {
	mu    sync.RWMutex
	items map[int64]*ShopItem
}

func newMemoryStore() *memoryStore {
	ms := &memoryStore{items: make(map[int64]*ShopItem)}
	// 预设一些道具
	for i := int64(1); i <= 10; i++ {
		ms.items[i] = &ShopItem{ID: i, ItemID: int32(100 + i), ItemName: fmt.Sprintf("Item_%d", i), Price: int32(i * 10), CurrencyType: 1, Discount: 100, LimitCount: -1, TabIndex: int32(i%3 + 1)}
	}
	return ms
}

func (s *memoryStore) ListItems(_ context.Context, tabIndex int32) ([]*ShopItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var items []*ShopItem
	for _, it := range s.items {
		if tabIndex <= 0 || it.TabIndex == tabIndex {
			items = append(items, it)
		}
	}
	return items, nil
}

func (s *memoryStore) BuyItem(_ context.Context, uid int64, shopItemID int64, count int32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	it, ok := s.items[shopItemID]
	if !ok {
		return fmt.Errorf("item not found")
	}
	if it.LimitCount > 0 && it.SoldCount+count > it.LimitCount {
		return fmt.Errorf("sold out")
	}
	it.SoldCount += count
	return nil
}
