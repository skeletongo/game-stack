package activity

import "context"

type Service interface {
	ListActivities(aType int32) ([]*ActivityInfo, error)
}

type service struct{ store Store }

func newService(store Store) *service { return &service{store: store} }

func (s *service) ListActivities(aType int32) ([]*ActivityInfo, error) {
	return s.store.ListActivities(context.Background(), aType)
}
