// Package infrastructure 提供 player 模块的基础设施层实现。
package infrastructure

import (
	"context"
	"fmt"
	"sync"

	"github.com/skeletongo/game-stack/module/player/internal/domain"
)

var _ domain.PlayerRepository = (*MemoryRepo)(nil)

// MemoryRepo 是 PlayerRepository 的内存实现，用于开发环境。
type MemoryRepo struct {
	mu       sync.RWMutex
	players  map[int64]*domain.Player
	nickname map[string]int64 // nickname → playerID
}

// NewMemoryRepo 创建内存仓储。
func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{
		players:  make(map[int64]*domain.Player),
		nickname: make(map[string]int64),
	}
}

// Load 按 ID 加载玩家。
func (r *MemoryRepo) Load(_ context.Context, id int64) (*domain.Player, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.players[id]
	if !ok {
		return nil, fmt.Errorf("player %d not found", id)
	}
	return p, nil
}

// Save 持久化玩家。
func (r *MemoryRepo) Save(_ context.Context, p *domain.Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.players[p.ID()] = p
	r.nickname[p.Nickname().String()] = p.ID()
	return nil
}

// Delete 删除玩家（幂等：已删除的玩家重复删除不报错）。
func (r *MemoryRepo) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.players[id]
	if !ok {
		return nil
	}
	delete(r.nickname, p.Nickname().String())
	delete(r.players, id)
	return nil
}

// FindByNickname 按昵称查找玩家。
func (r *MemoryRepo) FindByNickname(_ context.Context, nickname string) (*domain.Player, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.nickname[nickname]
	if !ok {
		return nil, fmt.Errorf("player with nickname %s not found", nickname)
	}
	return r.players[id], nil
}
