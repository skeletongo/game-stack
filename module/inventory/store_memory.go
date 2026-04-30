package inventory

import (
	"context"
	"fmt"
	"sync"
)

type memoryStore struct {
	mu   sync.RWMutex
	bags map[int64][]*Item
}

func newMemoryStore() *memoryStore {
	return &memoryStore{bags: make(map[int64][]*Item)}
}

func (s *memoryStore) ensureBag(uid int64) {
	if _, ok := s.bags[uid]; !ok {
		s.bags[uid] = make([]*Item, 0)
	}
}

func (s *memoryStore) ListItems(_ context.Context, uid int64, bagType int32) ([]*Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.ensureBag(uid)
	items := s.bags[uid]
	if bagType == 0 {
		return items, nil
	}

	var filtered []*Item
	for _, it := range items {
		if it.Type == bagType || bagType == 3 {
			filtered = append(filtered, it)
		}
	}
	return filtered, nil
}

func (s *memoryStore) AddItem(_ context.Context, uid int64, item *Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ensureBag(uid)
	s.bags[uid] = append(s.bags[uid], item)
	return nil
}

func (s *memoryStore) RemoveItem(_ context.Context, uid int64, itemID int64, count int32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ensureBag(uid)
	for _, it := range s.bags[uid] {
		if it.ID == itemID {
			if it.Count < count {
				return fmt.Errorf("not enough items")
			}
			it.Count -= count
			return nil
		}
	}
	return fmt.Errorf("item %d not found", itemID)
}

func (s *memoryStore) UseItem(_ context.Context, uid int64, itemID int64, count int32) error {
	return s.RemoveItem(context.Background(), uid, itemID, count)
}

func (s *memoryStore) EquipItem(_ context.Context, uid int64, itemID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ensureBag(uid)
	for _, it := range s.bags[uid] {
		if it.ID == itemID {
			it.Equipped = true
			return nil
		}
	}
	return fmt.Errorf("item %d not found", itemID)
}

func (s *memoryStore) UnequipItem(_ context.Context, uid int64, itemID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ensureBag(uid)
	for _, it := range s.bags[uid] {
		if it.ID == itemID {
			it.Equipped = false
			return nil
		}
	}
	return fmt.Errorf("item %d not found", itemID)
}

func (s *memoryStore) GetBagSize(_ context.Context, uid int64) (int32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.ensureBag(uid)
	return int32(len(s.bags[uid])), nil
}

func (s *memoryStore) RemovePlayerBag(_ context.Context, uid int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.bags, uid)
	return nil
}
