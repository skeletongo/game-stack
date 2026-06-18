// Package infrastructure 提供 auth 模块的基础设施层实现。
package infrastructure

import (
	"context"
	"fmt"
	"sync"

	"github.com/skeletongo/game-stack/module/auth/internal/domain"
)

var _ domain.AccountRepository = (*MemoryRepo)(nil)

// MemoryRepo 是 AccountRepository 的内存实现，用于开发环境。
type MemoryRepo struct {
	mu       sync.RWMutex
	accounts map[int64]*domain.Account
	byName   map[string]int64 // username → userID
}

// NewMemoryRepo 创建内存仓储。
func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{
		accounts: make(map[int64]*domain.Account),
		byName:   make(map[string]int64),
	}
}

// Load 按 ID 加载账户。
func (r *MemoryRepo) Load(_ context.Context, id int64) (*domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.accounts[id]
	if !ok {
		return nil, fmt.Errorf("account %d not found", id)
	}
	return a, nil
}

// Save 持久化账户。
func (r *MemoryRepo) Save(_ context.Context, a *domain.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.accounts[a.ID()] = a
	r.byName[a.Username().String()] = a.ID()
	return nil
}

// Delete 删除账户（幂等：已删除的账户重复删除不报错）。
func (r *MemoryRepo) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.accounts[id]
	if !ok {
		return nil
	}
	delete(r.byName, a.Username().String())
	delete(r.accounts, id)
	return nil
}

// FindByUsername 按用户名查找账户。
func (r *MemoryRepo) FindByUsername(_ context.Context, username string) (*domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byName[username]
	if !ok {
		return nil, fmt.Errorf("account %s not found", username)
	}
	return r.accounts[id], nil
}
