// Package mail 定义邮件系统相关消息类型。
package mail

// Attachment 邮件附件。
type Attachment struct {
	ItemID   int32  `json:"itemId" msgpack:"itemId"`
	ItemName string `json:"itemName" msgpack:"itemName"`
	Count    int32  `json:"count" msgpack:"count"`
	Received bool   `json:"received" msgpack:"received"`
}

// Mail 邮件信息。
type Mail struct {
	ID          int64         `json:"id" msgpack:"id"`
	SenderID    int64         `json:"senderId" msgpack:"senderId"`
	SenderName  string        `json:"senderName" msgpack:"senderName"`
	ReceiverID  int64         `json:"receiverId" msgpack:"receiverId"`
	Title       string        `json:"title" msgpack:"title"`
	Content     string        `json:"content" msgpack:"content"`
	Attachments []*Attachment `json:"attachments" msgpack:"attachments"`
	Read        bool          `json:"read" msgpack:"read"`
	SentAt      int64         `json:"sentAt" msgpack:"sentAt"`
	ExpireAt    int64         `json:"expireAt" msgpack:"expireAt"`
}

// ListRequest 获取邮件列表请求。
type ListRequest struct {
	Page     int32 `json:"page" msgpack:"page"`
	PageSize int32 `json:"pageSize" msgpack:"pageSize"`
}

// ListResponse 获取邮件列表响应。
type ListResponse struct {
	Mails []*Mail `json:"mails" msgpack:"mails"`
	Total int32   `json:"total" msgpack:"total"`
}

// ReadRequest 阅读邮件请求。
type ReadRequest struct {
	MailID int64 `json:"mailId" msgpack:"mailId"`
}

// ReadResponse 阅读邮件响应。
type ReadResponse struct {
	Mail *Mail `json:"mail" msgpack:"mail"`
}

// ReceiveAttachmentRequest 领取附件请求。
type ReceiveAttachmentRequest struct {
	MailID int64 `json:"mailId" msgpack:"mailId"`
	ItemID int32 `json:"itemId" msgpack:"itemId"`
}

// DeleteRequest 删除邮件请求。
type DeleteRequest struct {
	MailIDs []int64 `json:"mailIds" msgpack:"mailIds"`
}

// SendRequest 发送邮件请求（系统邮件）。
type SendRequest struct {
	ReceiverID  int64         `json:"receiverId" msgpack:"receiverId"`
	Title       string        `json:"title" msgpack:"title"`
	Content     string        `json:"content" msgpack:"content"`
	Attachments []*Attachment `json:"attachments" msgpack:"attachments"`
	ExpireAt    int64         `json:"expireAt" msgpack:"expireAt"`
}

// NewMailEvent 新邮件事件（服务器推送）。
type NewMailEvent struct {
	MailID     int64  `json:"mailId" msgpack:"mailId"`
	Title      string `json:"title" msgpack:"title"`
	SenderName string `json:"senderName" msgpack:"senderName"`
}
