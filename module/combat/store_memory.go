package combat

import (
	"context"
	"fmt"
	"sync"
)

type memoryStore struct {
	mu     sync.RWMutex
	states map[int64]*CombatState
}

func newMemoryStore() *memoryStore {
	return &memoryStore{states: make(map[int64]*CombatState)}
}

func (s *memoryStore) ensure(uid int64) {
	if _, ok := s.states[uid]; !ok {
		s.states[uid] = &CombatState{
			PlayerID: uid,
			HP:       1000,
			MaxHP:    1000,
			MP:       500,
			MaxMP:    500,
			Skills:   make([]*Skill, 0),
			Buffs:    make([]*Buff, 0),
		}
	}
}

func (s *memoryStore) GetCombatState(_ context.Context, uid int64) (*CombatState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.ensure(uid)
	return s.states[uid], nil
}

func (s *memoryStore) UpdateHP(_ context.Context, uid int64, delta int32) (int32, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensure(uid)
	s.states[uid].HP += delta
	if s.states[uid].HP < 0 {
		s.states[uid].HP = 0
	}
	if s.states[uid].HP > s.states[uid].MaxHP {
		s.states[uid].HP = s.states[uid].MaxHP
	}
	return s.states[uid].HP, nil
}

func (s *memoryStore) UpdateMP(_ context.Context, uid int64, delta int32) (int32, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensure(uid)
	s.states[uid].MP += delta
	return s.states[uid].MP, nil
}

func (s *memoryStore) AddBuff(_ context.Context, uid int64, buff *Buff) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensure(uid)
	s.states[uid].Buffs = append(s.states[uid].Buffs, buff)
	return nil
}

func (s *memoryStore) RemoveCombatState(_ context.Context, uid int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.states, uid)
	return nil
}

func (s *memoryStore) RemoveBuff(_ context.Context, uid int64, buffID int32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensure(uid)
	for i, b := range s.states[uid].Buffs {
		if b.ID == buffID {
			s.states[uid].Buffs = append(s.states[uid].Buffs[:i], s.states[uid].Buffs[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("buff not found")
}
