package leaderboard

import "context"

type Service interface {
	UpdateScore(boardName string, uid int64, nickname string, score int64, level int32) error
	GetRank(boardName string, uid int64) (*Entry, error)
}

type service struct{ store Store }

func newService(store Store) *service { return &service{store: store} }

func (s *service) UpdateScore(boardName string, uid int64, nickname string, score int64, level int32) error {
	return s.store.UpdateScore(context.Background(), boardName, uid, nickname, score, level)
}

func (s *service) GetRank(boardName string, uid int64) (*Entry, error) {
	return s.store.GetRank(context.Background(), boardName, uid)
}
