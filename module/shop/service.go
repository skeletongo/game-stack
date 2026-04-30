package shop

import "context"

type Service interface {
	ListItems(tabIndex int32) ([]*ShopItem, error)
}

type service struct{ store Store }

func newService(store Store) *service { return &service{store: store} }

func (s *service) ListItems(tabIndex int32) ([]*ShopItem, error) {
	return s.store.ListItems(context.Background(), tabIndex)
}
