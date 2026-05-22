package player

import "context"

// Service 玩家模块对外的服务接口。
type Service interface {
	// GetPlayer 获取玩家信息。
	GetPlayer(id int64) (*Player, error)
	// AddExp 增加经验值，返回新等级和是否升级。
	AddExp(id int64, exp int64) (newLevel int32, leveledUp bool, err error)
	// AddGold 增加金币。
	AddGold(id int64, gold int32) error
	// AddDiamond 增加钻石。
	AddDiamond(id int64, diamond int32) error
	// DeductGold 扣除金币（返回错误如果不足）。
	DeductGold(id int64, gold int32) error
	// DeductDiamond 扣除钻石（返回错误如果不足）。
	DeductDiamond(id int64, diamond int32) error
}

type service struct {
	store Store
}

func newService(store Store) *service {
	return &service{store: store}
}

func (s *service) GetPlayer(id int64) (*Player, error) {
	return s.store.GetPlayer(context.Background(), id)
}

func (s *service) AddExp(id int64, exp int64) (int32, bool, error) {
	p, err := s.store.GetPlayer(context.Background(), id)
	if err != nil {
		return 0, false, err
	}

	p.Exp += exp
	oldLevel := p.Level
	newLevel := calcLevel(p.Exp)

	if newLevel > oldLevel {
		p.Level = newLevel
		_ = s.store.UpdatePlayer(context.Background(), p)
		return newLevel, true, nil
	}

	_ = s.store.UpdatePlayer(context.Background(), p)
	return oldLevel, false, nil
}

func (s *service) AddGold(id int64, gold int32) error {
	p, err := s.store.GetPlayer(context.Background(), id)
	if err != nil {
		return err
	}
	p.Gold += gold
	return s.store.UpdatePlayer(context.Background(), p)
}

func (s *service) AddDiamond(id int64, diamond int32) error {
	p, err := s.store.GetPlayer(context.Background(), id)
	if err != nil {
		return err
	}
	p.Diamond += diamond
	return s.store.UpdatePlayer(context.Background(), p)
}

func (s *service) DeductGold(id int64, gold int32) error {
	p, err := s.store.GetPlayer(context.Background(), id)
	if err != nil {
		return err
	}
	if p.Gold < gold {
		return ErrNotEnoughGold
	}
	p.Gold -= gold
	return s.store.UpdatePlayer(context.Background(), p)
}

func (s *service) DeductDiamond(id int64, diamond int32) error {
	p, err := s.store.GetPlayer(context.Background(), id)
	if err != nil {
		return err
	}
	if p.Diamond < diamond {
		return ErrNotEnoughDiamond
	}
	p.Diamond -= diamond
	return s.store.UpdatePlayer(context.Background(), p)
}

// CleanPlayerData 清理玩家内存数据（断线时调用）。
func (s *service) CleanPlayerData(uid int64) error {
	return s.store.RemovePlayer(context.Background(), uid)
}

// calcLevel 根据总经验值计算等级。
func calcLevel(exp int64) int32 {
	level := int32(1)
	needed := int64(100)
	for exp >= needed {
		exp -= needed
		level++
		needed = int64(level) * 100
	}
	return level
}
