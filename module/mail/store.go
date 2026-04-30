package mail

import "context"

// Attachment 邮件附件。
type Attachment struct {
	ItemID   int32
	ItemName string
	Count    int32
	Received bool
}

// Mail 邮件数据。
type Mail struct {
	ID          int64
	SenderID    int64
	SenderName  string
	ReceiverID  int64
	Title       string
	Content     string
	Attachments []*Attachment
	Read        bool
	SentAt      int64
	ExpireAt    int64
}

// Store 定义邮件模块的数据存储接口。
type Store interface {
	CreateMail(ctx context.Context, mail *Mail) error
	GetMail(ctx context.Context, mailID int64) (*Mail, error)
	ListMail(ctx context.Context, uid int64, page, pageSize int32) ([]*Mail, int32, error)
	MarkRead(ctx context.Context, mailID int64) error
	MarkAttachmentReceived(ctx context.Context, mailID int64, itemID int32) error
	DeleteMail(ctx context.Context, mailIDs []int64) error
}
