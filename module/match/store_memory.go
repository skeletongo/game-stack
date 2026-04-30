package match

import (
	"context"
	"errors"
	"sync"
	"time"
)

type memoryStore struct {
	mu    sync.RWMutex
	queue map[int64]*MatchInfo
}

func newMemoryStore() *memoryStore {
	return &memoryStore{queue: make(map[int64]*MatchInfo)}
}

func (s *memoryStore) Push(_ context.Context, uid int64, info *MatchInfo) error {
	info.JoinedAt = time.Now().Unix()

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.queue[uid]; ok {
		return errors.New("already in queue")
	}
	s.queue[uid] = info
	return nil
}

func (s *memoryStore) Pop(_ context.Context, uid int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.queue[uid]; !ok {
		return errors.New("not in queue")
	}
	delete(s.queue, uid)
	return nil
}

func (s *memoryStore) FindMatch(_ context.Context, info *MatchInfo) ([]int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var candidates []*MatchInfo
	for _, m := range s.queue {
		candidates = append(candidates, m)
	}

	if m := winMatch(info, candidates); m != nil {
		delete(s.queue, info.UID)
		delete(s.queue, m.UID)
		return []int64{info.UID, m.UID}, nil
	}
	return nil, errors.New("no match found")
}

func (s *memoryStore) GetStatus(_ context.Context, uid int64) (*MatchInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, ok := s.queue[uid]
	if !ok {
		return nil, errors.New("not in queue")
	}
	return info, nil
}
