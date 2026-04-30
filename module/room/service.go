package room

import "context"

// Service 房间模块对外服务接口。
type Service interface {
	// GetRoom 获取房间信息。
	GetRoom(roomID int64) (*Room, error)
}

type service struct {
	store Store
}

func newService(store Store) *service {
	return &service{store: store}
}

func (s *service) GetRoom(roomID int64) (*Room, error) {
	return s.store.GetRoom(context.Background(), roomID)
}
