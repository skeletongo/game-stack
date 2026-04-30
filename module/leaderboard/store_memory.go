package leaderboard

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

type memoryStore struct {
	mu     sync.RWMutex
	boards map[string]map[int64]*Entry
}

func newMemoryStore() *memoryStore {
	return &memoryStore{boards: make(map[string]map[int64]*Entry)}
}

func (s *memoryStore) ensure(boardName string) {
	if _, ok := s.boards[boardName]; !ok {
		s.boards[boardName] = make(map[int64]*Entry)
	}
}

func (s *memoryStore) UpdateScore(_ context.Context, boardName string, uid int64, nickname string, score int64, level int32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensure(boardName)
	s.boards[boardName][uid] = &Entry{PlayerID: uid, Nickname: nickname, Score: score, Level: level}
	// re-rank
	var entries []*Entry
	for _, e := range s.boards[boardName] {
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Score > entries[j].Score })
	for i, e := range entries {
		e.Rank = int32(i + 1)
	}
	return nil
}

func (s *memoryStore) GetRank(_ context.Context, boardName string, uid int64) (*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.ensure(boardName)
	e, ok := s.boards[boardName][uid]
	if !ok {
		return nil, fmt.Errorf("not ranked")
	}
	return e, nil
}

func (s *memoryStore) GetTop(_ context.Context, boardName string, page, pageSize int32) ([]*Entry, int32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.ensure(boardName)
	var entries []*Entry
	for _, e := range s.boards[boardName] {
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Score > entries[j].Score })
	total := int32(len(entries))
	start := (page - 1) * pageSize
	if start >= total {
		return nil, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return entries[start:end], total, nil
}
