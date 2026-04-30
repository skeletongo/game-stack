package mail

import (
	"context"
	"time"
)

// Service 邮件模块对外的服务接口。
// 其他模块可以通过此接口发送系统邮件。
type Service interface {
	// SendSystemMail 发送系统邮件。
	SendSystemMail(receiverID int64, title, content string, attachments []*Attachment, expireDays int32) (int64, error)
	// SendMail 发送邮件（玩家间）。
	SendMail(senderID int64, senderName string, receiverID int64, title, content string, attachments []*Attachment, expireDays int32) (int64, error)
}

type service struct {
	store Store
}

func newService(store Store) *service {
	return &service{store: store}
}

func (s *service) SendSystemMail(receiverID int64, title, content string, attachments []*Attachment, expireDays int32) (int64, error) {
	return s.SendMail(0, "System", receiverID, title, content, attachments, expireDays)
}

func (s *service) SendMail(senderID int64, senderName string, receiverID int64, title, content string, attachments []*Attachment, expireDays int32) (int64, error) {
	now := time.Now().Unix()
	mail := &Mail{
		ID:          now * 1000, // simple ID generation
		SenderID:    senderID,
		SenderName:  senderName,
		ReceiverID:  receiverID,
		Title:       title,
		Content:     content,
		Attachments: attachments,
		Read:        false,
		SentAt:      now,
		ExpireAt:    now + int64(expireDays)*86400,
	}

	if err := s.store.CreateMail(context.Background(), mail); err != nil {
		return 0, err
	}

	return mail.ID, nil
}
