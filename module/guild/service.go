package guild

import "context"

type Service interface {
	GetGuild(guildID int64) (*Guild, error)
}

type service struct{ store Store }

func newService(store Store) *service { return &service{store: store} }

func (s *service) GetGuild(guildID int64) (*Guild, error) {
	return s.store.GetGuild(context.Background(), guildID)
}
