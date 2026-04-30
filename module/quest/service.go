package quest

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

// nextQuestID for simple ID generation.
var nextQuestID int32

func init() { nextQuestID = 1000 }

func newQuestID() int32 {
	nextQuestID++
	return nextQuestID
}
