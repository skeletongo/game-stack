package chat

import "context"

// Message 聊天消息。
type Message struct {
	ID         int64
	SenderID   int64
	SenderName string
	Content    string
	Channel    string
	TargetID   int64
	SentAt     int64
}

// Store 聊天模块数据存储接口。
type Store interface {
	SaveMessage(ctx context.Context, msg *Message) error
	GetHistory(ctx context.Context, channel string, targetID int64, page, pageSize int32) ([]*Message, int32, error)
}
