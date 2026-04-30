package social

import (
	"context"
	"fmt"
	"sync"
)

const maxFriends = 200

type memoryStore struct {
	mu        sync.RWMutex
	friends   map[int64]map[int64]*FriendInfo // uid -> friendID -> FriendInfo
	blacklist map[int64]map[int64]bool        // uid -> blockedID -> true
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		friends:   make(map[int64]map[int64]*FriendInfo),
		blacklist: make(map[int64]map[int64]bool),
	}
}

func (s *memoryStore) ensureFriends(uid int64) {
	if _, ok := s.friends[uid]; !ok {
		s.friends[uid] = make(map[int64]*FriendInfo)
	}
}

func (s *memoryStore) ensureBlacklist(uid int64) {
	if _, ok := s.blacklist[uid]; !ok {
		s.blacklist[uid] = make(map[int64]bool)
	}
}

func (s *memoryStore) AddFriend(_ context.Context, uid int64, friend *FriendInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureFriends(uid)
	if _, ok := s.friends[uid][friend.PlayerID]; ok {
		return fmt.Errorf("already friends")
	}
	if int32(len(s.friends[uid])) >= maxFriends {
		return fmt.Errorf("friend list full")
	}
	s.friends[uid][friend.PlayerID] = friend
	return nil
}

func (s *memoryStore) RemoveFriend(_ context.Context, uid int64, friendID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureFriends(uid)
	delete(s.friends[uid], friendID)
	return nil
}

func (s *memoryStore) ListFriends(_ context.Context, uid int64, page, pageSize int32) ([]*FriendInfo, int32, int32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.ensureFriends(uid)
	var list []*FriendInfo
	for _, f := range s.friends[uid] {
		list = append(list, f)
	}
	total := int32(len(list))
	start := (page - 1) * pageSize
	if start >= total {
		return nil, total, maxFriends, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return list[start:end], total, maxFriends, nil
}

func (s *memoryStore) BlockUser(_ context.Context, uid int64, targetID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureBlacklist(uid)
	s.blacklist[uid][targetID] = true
	return nil
}

func (s *memoryStore) UnblockUser(_ context.Context, uid int64, targetID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureBlacklist(uid)
	delete(s.blacklist[uid], targetID)
	return nil
}

func (s *memoryStore) RemovePlayerData(_ context.Context, uid int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.friends, uid)
	delete(s.blacklist, uid)
	return nil
}

func (s *memoryStore) ListBlacklist(_ context.Context, uid int64, page, pageSize int32) ([]*FriendInfo, int32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.ensureBlacklist(uid)
	var list []*FriendInfo
	for id := range s.blacklist[uid] {
		list = append(list, &FriendInfo{PlayerID: id})
	}
	total := int32(len(list))
	start := (page - 1) * pageSize
	if start >= total {
		return nil, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return list[start:end], total, nil
}
