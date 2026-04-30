package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

// Service 认证模块对外的服务接口。
// 其他模块通过此接口进行认证相关的操作。
type Service interface {
	// Authenticate 验证 token 并返回玩家 ID。
	Authenticate(token string) (int64, error)
	// OnlineCount 返回当前在线玩家数。
	OnlineCount() int64
	// IsOnline 检查玩家是否在线。
	IsOnline(uid int64) bool
	// KickPlayer 强制踢下线。
	KickPlayer(uid int64) error
}

type service struct {
	store Store
	mu    sync.RWMutex
}

func newService(store Store) *service {
	return &service{store: store}
}

func (s *service) Authenticate(token string) (int64, error) {
	uid, err := s.store.GetTokenByValue(context.Background(), token)
	if err != nil {
		return 0, err
	}
	return uid, nil
}

func (s *service) OnlineCount() int64 {
	count, err := s.store.OnlineCount(context.Background())
	if err != nil {
		return 0
	}
	return count
}

func (s *service) IsOnline(uid int64) bool {
	online, err := s.store.IsOnline(context.Background(), uid)
	if err != nil {
		return false
	}
	return online
}

func (s *service) KickPlayer(uid int64) error {
	return s.store.DeleteToken(context.Background(), uid)
}

// generateToken 生成随机 token。
func generateToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// hashPassword 对密码做 SHA256 哈希。
func hashPassword(password string) string {
	h := sha256.Sum256([]byte(password))
	return hex.EncodeToString(h[:])
}
