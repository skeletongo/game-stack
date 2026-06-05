package domain

import (
	"time"

	"github.com/skeletongo/game-stack/ddd"
)

// Player 是玩家聚合根。所有属性修改都通过行为方法，不暴露 setter。
//
// 聚合不变量：
//   - Gold >= 0, Diamond >= 0, Exp >= 0
//   - Level = CalcLevel(Exp)（等级由经验值推导，不可独立修改）
//   - Nickname 1-16 字符
type Player struct {
	id        PlayerID
	nickname  Nickname
	level     Level
	exp       Exp
	gold      Gold
	diamond   Diamond
	avatar    Avatar
	createdAt int64
	updatedAt int64
}

var _ ddd.Aggregate = (*Player)(nil)

// NewPlayer 创建新玩家聚合（工厂方法）。
// 所有值对象在构造时校验，确保聚合始终处于有效状态。
func NewPlayer(id int64, nickname string, level, gold, diamond int32, exp int64, avatar string) (*Player, error) {
	nn, err := NewNickname(nickname)
	if err != nil {
		return nil, err
	}
	lv, err := NewLevel(level)
	if err != nil {
		return nil, err
	}
	g, err := NewGold(gold)
	if err != nil {
		return nil, err
	}
	d, err := NewDiamond(diamond)
	if err != nil {
		return nil, err
	}
	e, err := NewExp(exp)
	if err != nil {
		return nil, err
	}
	return &Player{
		id:        PlayerID(id),
		nickname:  nn,
		level:     lv,
		exp:       e,
		gold:      g,
		diamond:   d,
		avatar:    Avatar(avatar),
		createdAt: time.Now().Unix(),
		updatedAt: time.Now().Unix(),
	}, nil
}

// ReconstructPlayer 从持久化数据重建玩家（仓储专用，跳过校验）。
func ReconstructPlayer(id int64, nickname string, level, gold, diamond int32, exp int64, avatar string, createdAt, updatedAt int64) *Player {
	return &Player{
		id:        PlayerID(id),
		nickname:  Nickname(nickname),
		level:     Level(level),
		exp:       Exp(exp),
		gold:      Gold(gold),
		diamond:   Diamond(diamond),
		avatar:    Avatar(avatar),
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

// ID 返回聚合唯一标识。
func (p *Player) ID() int64 { return p.id.Int64() }

// 查询方法

func (p *Player) Nickname() Nickname { return p.nickname }
func (p *Player) Level() Level       { return p.level }
func (p *Player) Exp() Exp           { return p.exp }
func (p *Player) Gold() Gold         { return p.gold }
func (p *Player) Diamond() Diamond   { return p.diamond }
func (p *Player) Avatar() Avatar     { return p.avatar }
func (p *Player) CreatedAt() int64   { return p.createdAt }
func (p *Player) UpdatedAt() int64   { return p.updatedAt }

// 行为方法

// UpdateProfile 更新昵称和头像。
func (p *Player) UpdateProfile(nickname Nickname, avatar Avatar) {
	p.nickname = nickname
	p.avatar = avatar
	p.updatedAt = time.Now().Unix()
}

// AddExp 增加经验值。返回是否升级（调用方据此发布 PlayerLeveledUp 事件）。
func (p *Player) AddExp(amount int64) bool {
	p.exp, _ = p.exp.Add(amount)
	newLevel := CalcLevel(p.exp)
	if newLevel != p.level {
		p.level = newLevel
		p.updatedAt = time.Now().Unix()
		return true
	}
	p.updatedAt = time.Now().Unix()
	return false
}

// AddGold 增加金币。返回错误如果增量会导致溢出。
func (p *Player) AddGold(amount int32) error {
	ng, err := p.gold.Add(amount)
	if err != nil {
		return err
	}
	p.gold = ng
	p.updatedAt = time.Now().Unix()
	return nil
}

// DeductGold 扣除金币。余额不足时返回 stack.ErrInsufficientGold。
func (p *Player) DeductGold(amount int32) error {
	ng, err := p.gold.Subtract(amount)
	if err != nil {
		return err
	}
	p.gold = ng
	p.updatedAt = time.Now().Unix()
	return nil
}

// AddDiamond 增加钻石。
func (p *Player) AddDiamond(amount int32) error {
	nd, err := p.diamond.Add(amount)
	if err != nil {
		return err
	}
	p.diamond = nd
	p.updatedAt = time.Now().Unix()
	return nil
}

// DeductDiamond 扣除钻石。余额不足时返回 stack.ErrInsufficientDiamond。
func (p *Player) DeductDiamond(amount int32) error {
	nd, err := p.diamond.Subtract(amount)
	if err != nil {
		return err
	}
	p.diamond = nd
	p.updatedAt = time.Now().Unix()
	return nil
}
