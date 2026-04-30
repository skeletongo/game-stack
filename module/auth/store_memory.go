package auth

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type memoryStore struct {
	mu       sync.RWMutex
	users    map[int64]*User  // id -> user
	username map[string]int64 // username -> id
	tokens   map[int64]string // uid -> token
	tokenRev map[string]int64 // token -> uid
	online   map[int64]string // uid -> gid
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		users:    make(map[int64]*User),
		username: make(map[string]int64),
		tokens:   make(map[int64]string),
		tokenRev: make(map[string]int64),
		online:   make(map[int64]string),
	}
}

func (s *memoryStore) CreateUser(_ context.Context, user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[user.ID] = user
	s.username[user.Username] = user.ID
	return nil
}

func (s *memoryStore) GetUserByID(_ context.Context, id int64) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.users[id]
	if !ok {
		return nil, fmt.Errorf("user %d not found", id)
	}
	return u, nil
}

func (s *memoryStore) GetUserByUsername(_ context.Context, username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.username[username]
	if !ok {
		return nil, fmt.Errorf("user %s not found", username)
	}
	return s.users[id], nil
}

func (s *memoryStore) UpdateUser(_ context.Context, user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[user.ID] = user
	return nil
}

func (s *memoryStore) BanUser(_ context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.users[id]
	if !ok {
		return fmt.Errorf("user %d not found", id)
	}
	u.BannedAt = 1
	return nil
}

func (s *memoryStore) UnbanUser(_ context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.users[id]
	if !ok {
		return fmt.Errorf("user %d not found", id)
	}
	u.BannedAt = 0
	return nil
}

func (s *memoryStore) SetToken(_ context.Context, uid int64, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tokens[uid] = token
	s.tokenRev[token] = uid
	return nil
}

func (s *memoryStore) GetToken(_ context.Context, uid int64) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, ok := s.tokens[uid]
	if !ok {
		return "", errors.New("token not found")
	}
	return t, nil
}

func (s *memoryStore) DeleteToken(_ context.Context, uid int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if t, ok := s.tokens[uid]; ok {
		delete(s.tokenRev, t)
	}
	delete(s.tokens, uid)
	return nil
}

func (s *memoryStore) GetTokenByValue(_ context.Context, token string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	uid, ok := s.tokenRev[token]
	if !ok {
		return 0, errors.New("token not found")
	}
	return uid, nil
}

func (s *memoryStore) SetOnline(_ context.Context, uid int64, gid string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.online[uid] = gid
	return nil
}

func (s *memoryStore) SetOffline(_ context.Context, uid int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.online, uid)
	return nil
}

func (s *memoryStore) IsOnline(_ context.Context, uid int64) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.online[uid]
	return ok, nil
}

func (s *memoryStore) OnlineCount(_ context.Context) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return int64(len(s.online)), nil
}
