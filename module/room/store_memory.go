package room

import (
	"context"
	"fmt"
	"sync"
)

type memoryStore struct {
	mu    sync.RWMutex
	rooms map[int64]*Room
}

func newMemoryStore() *memoryStore {
	return &memoryStore{rooms: make(map[int64]*Room)}
}

func (s *memoryStore) CreateRoom(_ context.Context, room *Room) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rooms[room.ID] = room
	return nil
}

func (s *memoryStore) GetRoom(_ context.Context, roomID int64) (*Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.rooms[roomID]
	if !ok {
		return nil, fmt.Errorf("room %d not found", roomID)
	}
	return r, nil
}

func (s *memoryStore) ListRooms(_ context.Context, sceneID string, page, pageSize int32) ([]*Room, int32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var rooms []*Room
	for _, r := range s.rooms {
		if sceneID == "" || r.SceneID == sceneID {
			rooms = append(rooms, r)
		}
	}

	total := int32(len(rooms))
	start := (page - 1) * pageSize
	if start >= total {
		return nil, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	return rooms[start:end], total, nil
}

func (s *memoryStore) DeleteRoom(_ context.Context, roomID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.rooms, roomID)
	return nil
}

func (s *memoryStore) AddPlayer(_ context.Context, roomID int64, player *RoomPlayer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.rooms[roomID]
	if !ok {
		return fmt.Errorf("room %d not found", roomID)
	}
	for _, p := range r.Players {
		if p.PlayerID == player.PlayerID {
			return fmt.Errorf("player %d already in room", player.PlayerID)
		}
	}
	r.Players = append(r.Players, player)
	r.CurPlayers = int32(len(r.Players))
	return nil
}

func (s *memoryStore) RemovePlayer(_ context.Context, roomID int64, playerID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.rooms[roomID]
	if !ok {
		return fmt.Errorf("room %d not found", roomID)
	}
	for i, p := range r.Players {
		if p.PlayerID == playerID {
			r.Players = append(r.Players[:i], r.Players[i+1:]...)
			r.CurPlayers = int32(len(r.Players))
			if r.CurPlayers == 0 {
				delete(s.rooms, roomID)
			}
			return nil
		}
	}
	return fmt.Errorf("player %d not in room", playerID)
}

func (s *memoryStore) SetReady(_ context.Context, roomID int64, playerID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.rooms[roomID]
	if !ok {
		return fmt.Errorf("room %d not found", roomID)
	}
	for _, p := range r.Players {
		if p.PlayerID == playerID {
			p.Ready = true
			return nil
		}
	}
	return fmt.Errorf("player %d not in room", playerID)
}
