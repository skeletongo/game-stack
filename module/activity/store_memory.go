package activity

import (
	"context"
	"fmt"
	"sync"
)

type memoryStore struct {
	mu         sync.RWMutex
	activities map[int32]*ActivityInfo
	claims     map[string]bool // "uid_activityID" -> claimed
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		activities: make(map[int32]*ActivityInfo),
		claims:     make(map[string]bool),
	}
}

func (s *memoryStore) GetActivity(_ context.Context, activityID int32) (*ActivityInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.activities[activityID]
	if !ok {
		return nil, fmt.Errorf("activity not found")
	}
	return a, nil
}

func (s *memoryStore) ListActivities(_ context.Context, aType int32) ([]*ActivityInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []*ActivityInfo
	for _, a := range s.activities {
		if aType <= 0 || a.Type == aType {
			list = append(list, a)
		}
	}
	return list, nil
}

func (s *memoryStore) ClaimReward(_ context.Context, uid int64, activityID int32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := fmt.Sprintf("%d_%d", uid, activityID)
	if s.claims[key] {
		return fmt.Errorf("already claimed")
	}
	s.claims[key] = true
	return nil
}

func (s *memoryStore) UpdateProgress(_ context.Context, uid int64, activityID int32, progress int32) error {
	return nil
}

func (s *memoryStore) RemovePlayerClaims(_ context.Context, uid int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	prefix := fmt.Sprintf("%d_", uid)
	for k := range s.claims {
		if len(k) > len(prefix) && k[:len(prefix)] == prefix {
			delete(s.claims, k)
		}
	}
	return nil
}
