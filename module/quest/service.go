package quest

import "context"

type Service interface {
	CheckLevelQuests(uid int64) error
}

type service struct {
	store Store
}

func newService(store Store) *service {
	return &service{store: store}
}

func (s *service) CheckLevelQuests(uid int64) error {
	return nil
}

func (s *service) CleanPlayerData(uid int64) error {
	return s.store.RemovePlayerQuests(context.Background(), uid)
}

// nextQuestID for simple ID generation.
var nextQuestID int32

func init() { nextQuestID = 1000 }

func newQuestID() int32 {
	nextQuestID++
	return nextQuestID
}
