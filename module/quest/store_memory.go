package quest

import (
	"context"
	"fmt"
	"sync"
)

type memoryStore struct {
	mu           sync.RWMutex
	templates    map[int32]*Quest
	playerQuests map[int64]map[int32]*Quest
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		templates:    make(map[int32]*Quest),
		playerQuests: make(map[int64]map[int32]*Quest),
	}
}

func (s *memoryStore) GetAllQuests(_ context.Context) ([]*Quest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var quests []*Quest
	for _, q := range s.templates {
		quests = append(quests, q)
	}
	return quests, nil
}

func (s *memoryStore) GetPlayerQuest(_ context.Context, uid int64, questID int32) (*Quest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	pq, ok := s.playerQuests[uid]
	if !ok {
		return nil, fmt.Errorf("player %d quests not found", uid)
	}
	q, ok := pq[questID]
	if !ok {
		return nil, fmt.Errorf("quest %d not found", questID)
	}
	return q, nil
}

func (s *memoryStore) ListPlayerQuests(_ context.Context, uid int64, qType int32) ([]*Quest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	pq, ok := s.playerQuests[uid]
	if !ok {
		return nil, nil
	}
	var quests []*Quest
	for _, q := range pq {
		if qType == 0 || q.Type == qType {
			quests = append(quests, q)
		}
	}
	return quests, nil
}

func (s *memoryStore) AcceptQuest(_ context.Context, uid int64, questID int32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.playerQuests[uid]; !ok {
		s.playerQuests[uid] = make(map[int32]*Quest)
	}
	tpl, ok := s.templates[questID]
	if !ok {
		return fmt.Errorf("quest %d not found", questID)
	}
	q := *tpl
	q.Status = QuestDoing
	s.playerQuests[uid][questID] = &q
	return nil
}

func (s *memoryStore) SubmitQuest(_ context.Context, uid int64, questID int32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.playerQuests[uid]; !ok {
		return fmt.Errorf("player not found")
	}
	q, ok := s.playerQuests[uid][questID]
	if !ok {
		return fmt.Errorf("quest not found")
	}
	q.Status = QuestClaimed
	return nil
}

func (s *memoryStore) UpdateProgress(_ context.Context, uid int64, questID int32, progress int32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.playerQuests[uid]; !ok {
		return fmt.Errorf("player not found")
	}
	q, ok := s.playerQuests[uid][questID]
	if !ok {
		return fmt.Errorf("quest not found")
	}
	q.Progress = progress
	return nil
}
