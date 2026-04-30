// Package chat 定义聊天相关消息类型。
package chat

// ChatMessage 聊天消息。
type ChatMessage struct {
	ID         int64  `json:"id" msgpack:"id"`
	SenderID   int64  `json:"senderId" msgpack:"senderId"`
	SenderName string `json:"senderName" msgpack:"senderName"`
	Content    string `json:"content" msgpack:"content"`
	Channel    string `json:"channel" msgpack:"channel"` // world, guild, private
	TargetID   int64  `json:"targetId" msgpack:"targetId"`
	SentAt     int64  `json:"sentAt" msgpack:"sentAt"`
}

// SendRequest 发送消息请求。
type SendRequest struct {
	Channel  string `json:"channel" msgpack:"channel"`
	Content  string `json:"content" msgpack:"content"`
	TargetID int64  `json:"targetId" msgpack:"targetId"`
}

// HistoryRequest 获取聊天记录请求。
type HistoryRequest struct {
	Channel  string `json:"channel" msgpack:"channel"`
	TargetID int64  `json:"targetId" msgpack:"targetId"`
	Page     int32  `json:"page" msgpack:"page"`
	PageSize int32  `json:"pageSize" msgpack:"pageSize"`
}

// HistoryResponse 获取聊天记录响应。
type HistoryResponse struct {
	Messages []*ChatMessage `json:"messages" msgpack:"messages"`
	Total    int32          `json:"total" msgpack:"total"`
}

// NewMessageEvent 新消息事件（服务器推送）。
type NewMessageEvent struct {
	Message *ChatMessage `json:"message" msgpack:"message"`
}
