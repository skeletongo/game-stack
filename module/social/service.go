package social

import "context"

type Service interface {
	IsFriend(uid, targetID int64) bool
}

type service struct{ store Store }

func newService(store Store) *service { return &service{store: store} }

func (s *service) IsFriend(uid, targetID int64) bool {
	friends, _, _, _ := s.store.ListFriends(context.Background(), uid, 1, 1000)
	if friends == nil {
		return false
	}
	for _, f := range friends {
		if f.PlayerID == targetID {
			return true
		}
	}
	return false
}

func (s *service) CleanPlayerData(uid int64) {
	_ = s.store.RemovePlayerData(context.Background(), uid)
}
