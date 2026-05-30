// Package domain 定义 player 限界上下文的领域模型。
//
// Player 聚合是核心：封装玩家属性、货币和等级，
// 所有状态变更都通过行为方法执行并维护不变量。
package domain

import (
	"errors"

	"github.com/skeletongo/game-stack/ddd"
)

// 领域错误
var (
	ErrInvalidNickname = errors.New("nickname must be 1-16 characters")
	ErrInvalidLevel    = errors.New("level must be >= 1")
	ErrNegativeGold    = errors.New("gold must be >= 0")
	ErrNegativeDiamond = errors.New("diamond must be >= 0")
	ErrNegativeExp     = errors.New("exp must be >= 0")
)

// PlayerID 玩家唯一标识值对象。
// 强类型防止与 UserID 等其他 ID 混淆。
type PlayerID int64

func (id PlayerID) Int64() int64 { return int64(id) }
func (id PlayerID) Equals(other ddd.ValueObject) bool {
	o, ok := other.(PlayerID)
	return ok && id == o
}
func (id PlayerID) Scalar() any { return int64(id) }

// Nickname 玩家昵称值对象（1-16 字符）。
type Nickname string

// NewNickname 创建昵称，校验长度。
func NewNickname(s string) (Nickname, error) {
	if len(s) == 0 || len(s) > 16 {
		return "", ErrInvalidNickname
	}
	return Nickname(s), nil
}

func (n Nickname) String() string { return string(n) }
func (n Nickname) Equals(other ddd.ValueObject) bool {
	o, ok := other.(Nickname)
	return ok && n == o
}
func (n Nickname) Scalar() any { return string(n) }

// Level 玩家等级值对象（>= 1）。
type Level int32

// NewLevel 创建等级，校验 >= 1。
func NewLevel(v int32) (Level, error) {
	if v < 1 {
		return 0, ErrInvalidLevel
	}
	return Level(v), nil
}

func (l Level) Int32() int32 { return int32(l) }
func (l Level) Equals(other ddd.ValueObject) bool {
	o, ok := other.(Level)
	return ok && l == o
}
func (l Level) Scalar() any { return int32(l) }

// Gold 金币值对象（>= 0）。
// 封装 Add/Subtract 操作，确保不变量。
type Gold int32

// NewGold 创建金币，校验 >= 0。
func NewGold(v int32) (Gold, error) {
	if v < 0 {
		return 0, ErrNegativeGold
	}
	return Gold(v), nil
}

func (g Gold) Int32() int32 { return int32(g) }

// Add 增加金币，返回新的 Gold 值对象。
func (g Gold) Add(delta int32) (Gold, error) { return NewGold(int32(g) + delta) }

// Subtract 扣除金币，返回新的 Gold 值对象。余额不足时返回错误。
func (g Gold) Subtract(delta int32) (Gold, error) { return NewGold(int32(g) - delta) }

func (g Gold) Equals(other ddd.ValueObject) bool {
	o, ok := other.(Gold)
	return ok && g == o
}
func (g Gold) Scalar() any { return int32(g) }

// Diamond 钻石值对象（>= 0）。
type Diamond int32

// NewDiamond 创建钻石，校验 >= 0。
func NewDiamond(v int32) (Diamond, error) {
	if v < 0 {
		return 0, ErrNegativeDiamond
	}
	return Diamond(v), nil
}

func (d Diamond) Int32() int32 { return int32(d) }

// Add 增加钻石，返回新的 Diamond 值对象。
func (d Diamond) Add(delta int32) (Diamond, error) { return NewDiamond(int32(d) + delta) }

// Subtract 扣除钻石，返回新的 Diamond 值对象。余额不足时返回错误。
func (d Diamond) Subtract(delta int32) (Diamond, error) { return NewDiamond(int32(d) - delta) }

func (d Diamond) Equals(other ddd.ValueObject) bool {
	o, ok := other.(Diamond)
	return ok && d == o
}
func (d Diamond) Scalar() any { return int32(d) }

// Exp 经验值值对象（>= 0）。
type Exp int64

// NewExp 创建经验值，校验 >= 0。
func NewExp(v int64) (Exp, error) {
	if v < 0 {
		return 0, ErrNegativeExp
	}
	return Exp(v), nil
}

func (e Exp) Int64() int64 { return int64(e) }

// Add 增加经验值，返回新的 Exp 值对象。
func (e Exp) Add(delta int64) (Exp, error) { return NewExp(int64(e) + delta) }

func (e Exp) Equals(other ddd.ValueObject) bool {
	o, ok := other.(Exp)
	return ok && e == o
}
func (e Exp) Scalar() any { return int64(e) }

// Avatar 头像 URL 值对象。
type Avatar string

func (a Avatar) String() string { return string(a) }
func (a Avatar) Equals(other ddd.ValueObject) bool {
	o, ok := other.(Avatar)
	return ok && a == o
}
func (a Avatar) Scalar() any { return string(a) }
