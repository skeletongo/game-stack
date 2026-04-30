package chat

import "context"

// Service 聊天模块对外服务接口。
type Service interface {
	// SendSystemMessage 发送系统消息到指定频道。
	SendSystemMessage(channel string, content string, targetID int64) error
}

type service struct {
	store Store
	opts  *options
}

func newService(store Store, opts *options) *service {
	return &service{store: store, opts: opts}
}

func (s *service) SendSystemMessage(channel string, content string, targetID int64) error {
	return s.store.SaveMessage(context.Background(), &Message{
		ID:         nextID(),
		SenderID:   0,
		SenderName: "System",
		Content:    content,
		Channel:    channel,
		TargetID:   targetID,
	})
}

var lastID int64

func nextID() int64 {
	lastID++
	return lastID
}
