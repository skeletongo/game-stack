package player

import (
	"context"
	"fmt"
	"sync"
)

type memoryStore struct {
	mu      sync.RWMutex
	players map[int64]*Player
}

func newMemoryStore() *memoryStore {
	return &memoryStore{players: make(map[int64]*Player)}
}

func (s *memoryStore) CreatePlayer(_ context.Context, player *Player) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.players[player.ID] = player
	return nil
}

func (s *memoryStore) GetPlayer(_ context.Context, id int64) (*Player, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.players[id]
	if !ok {
		return nil, fmt.Errorf("player %d not found", id)
	}
	return p, nil
}

func (s *memoryStore) UpdatePlayer(_ context.Context, player *Player) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.players[player.ID] = player
	return nil
}

func (s *memoryStore) GetPlayerByName(_ context.Context, name string) (*Player, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.players {
		if p.Nickname == name {
			return p, nil
		}
	}
	return nil, fmt.Errorf("player %s not found", name)
}

func (s *memoryStore) RemovePlayer(_ context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.players, id)
	return nil
}
