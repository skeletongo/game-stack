package combat

import "context"

type Skill struct {
	ID          int32
	Name        string
	Level       int32
	Cooldown    int32
	ManaCost    int32
	Damage      int32
	Description string
}

type Buff struct {
	ID       int32
	Name     string
	Duration int32
	Type     int32
	Value    int32
}

type CombatState struct {
	PlayerID int64
	HP       int32
	MaxHP    int32
	MP       int32
	MaxMP    int32
	Skills   []*Skill
	Buffs    []*Buff
}

type Store interface {
	GetCombatState(ctx context.Context, uid int64) (*CombatState, error)
	UpdateHP(ctx context.Context, uid int64, delta int32) (int32, error)
	UpdateMP(ctx context.Context, uid int64, delta int32) (int32, error)
	AddBuff(ctx context.Context, uid int64, buff *Buff) error
	RemoveBuff(ctx context.Context, uid int64, buffID int32) error
	// RemoveCombatState 删除玩家战斗状态（断线清理用）。
	RemoveCombatState(ctx context.Context, uid int64) error
}
