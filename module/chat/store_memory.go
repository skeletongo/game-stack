package chat

import (
	"context"
	"sync"
	"time"
)

type memoryStore struct {
	mu       sync.RWMutex
	messages []*Message
}

func newMemoryStore() *memoryStore {
	return &memoryStore{messages: make([]*Message, 0)}
}

func (s *memoryStore) SaveMessage(_ context.Context, msg *Message) error {
	msg.SentAt = time.Now().Unix()
	s.mu.Lock()
	s.messages = append(s.messages, msg)
	// 限制内存消息数
	if len(s.messages) > 10000 {
		s.messages = s.messages[len(s.messages)-5000:]
	}
	s.mu.Unlock()
	return nil
}

func (s *memoryStore) GetHistory(_ context.Context, channel string, targetID int64, page, pageSize int32) ([]*Message, int32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []*Message
	for i := len(s.messages) - 1; i >= 0; i-- {
		m := s.messages[i]
		if m.Channel == channel || channel == "" {
			if targetID == 0 || m.TargetID == targetID {
				filtered = append(filtered, m)
			}
		}
	}

	total := int32(len(filtered))
	start := (page - 1) * pageSize
	if start >= total {
		return nil, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}
