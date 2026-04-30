package mail

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

type memoryStore struct {
	mu    sync.RWMutex
	mails map[int64]*Mail
}

func newMemoryStore() *memoryStore {
	return &memoryStore{mails: make(map[int64]*Mail)}
}

func (s *memoryStore) CreateMail(_ context.Context, mail *Mail) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mails[mail.ID] = mail
	return nil
}

func (s *memoryStore) GetMail(_ context.Context, mailID int64) (*Mail, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.mails[mailID]
	if !ok {
		return nil, fmt.Errorf("mail %d not found", mailID)
	}
	return m, nil
}

func (s *memoryStore) ListMail(_ context.Context, uid int64, page, pageSize int32) ([]*Mail, int32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var userMails []*Mail
	for _, m := range s.mails {
		if m.ReceiverID == uid {
			userMails = append(userMails, m)
		}
	}

	sort.Slice(userMails, func(i, j int) bool {
		return userMails[i].SentAt > userMails[j].SentAt
	})

	total := int32(len(userMails))
	start := (page - 1) * pageSize
	if start >= total {
		return nil, total, nil
	}

	end := start + pageSize
	if end > total {
		end = total
	}

	return userMails[start:end], total, nil
}

func (s *memoryStore) MarkRead(_ context.Context, mailID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.mails[mailID]
	if !ok {
		return fmt.Errorf("mail %d not found", mailID)
	}
	m.Read = true
	return nil
}

func (s *memoryStore) MarkAttachmentReceived(_ context.Context, mailID int64, itemID int32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.mails[mailID]
	if !ok {
		return fmt.Errorf("mail %d not found", mailID)
	}
	for _, a := range m.Attachments {
		if a.ItemID == itemID {
			a.Received = true
			return nil
		}
	}
	return fmt.Errorf("attachment %d not found in mail %d", itemID, mailID)
}

func (s *memoryStore) DeleteMail(_ context.Context, mailIDs []int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range mailIDs {
		delete(s.mails, id)
	}
	return nil
}
