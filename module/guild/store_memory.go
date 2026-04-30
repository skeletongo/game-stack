package guild

import (
	"context"
	"fmt"
	"sync"
)

type memoryStore struct {
	mu     sync.RWMutex
	guilds map[int64]*Guild
}

func newMemoryStore() *memoryStore { return &memoryStore{guilds: make(map[int64]*Guild)} }

func (s *memoryStore) CreateGuild(_ context.Context, guild *Guild) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, g := range s.guilds {
		if g.Name == guild.Name {
			return fmt.Errorf("name exists")
		}
	}
	s.guilds[guild.ID] = guild
	return nil
}

func (s *memoryStore) GetGuild(_ context.Context, guildID int64) (*Guild, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	g, ok := s.guilds[guildID]
	if !ok {
		return nil, fmt.Errorf("guild not found")
	}
	return g, nil
}

func (s *memoryStore) ListGuilds(_ context.Context, page, pageSize int32) ([]*Guild, int32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []*Guild
	for _, g := range s.guilds {
		list = append(list, g)
	}
	total := int32(len(list))
	start := (page - 1) * pageSize
	if start >= total {
		return nil, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return list[start:end], total, nil
}

func (s *memoryStore) DeleteGuild(_ context.Context, guildID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.guilds, guildID)
	return nil
}

func (s *memoryStore) AddMember(_ context.Context, guildID int64, member *Member) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	g, ok := s.guilds[guildID]
	if !ok {
		return fmt.Errorf("guild not found")
	}
	if int32(len(g.Members)) >= g.MaxMembers {
		return fmt.Errorf("guild full")
	}
	g.Members = append(g.Members, member)
	g.MemberCount = int32(len(g.Members))
	return nil
}

func (s *memoryStore) RemoveMember(_ context.Context, guildID int64, playerID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	g, ok := s.guilds[guildID]
	if !ok {
		return fmt.Errorf("guild not found")
	}
	for i, m := range g.Members {
		if m.PlayerID == playerID {
			g.Members = append(g.Members[:i], g.Members[i+1:]...)
			g.MemberCount = int32(len(g.Members))
			return nil
		}
	}
	return fmt.Errorf("member not found")
}

func (s *memoryStore) UpdateMemberPosition(_ context.Context, guildID int64, playerID int64, position int32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	g, ok := s.guilds[guildID]
	if !ok {
		return fmt.Errorf("guild not found")
	}
	for _, m := range g.Members {
		if m.PlayerID == playerID {
			m.Position = position
			return nil
		}
	}
	return fmt.Errorf("member not found")
}

func (s *memoryStore) Donate(_ context.Context, guildID int64, playerID int64, gold int32) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	g, ok := s.guilds[guildID]
	if !ok {
		return 0, fmt.Errorf("guild not found")
	}
	g.Exp += int64(gold)
	g.Gold += gold
	for _, m := range g.Members {
		if m.PlayerID == playerID {
			m.Donate += int64(gold)
			break
		}
	}
	return g.Exp, nil
}
