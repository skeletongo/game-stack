// Package combat 定义战斗/技能相关消息类型。
package combat

// Skill 技能信息。
type Skill struct {
	ID          int32  `json:"id" msgpack:"id"`
	Name        string `json:"name" msgpack:"name"`
	Level       int32  `json:"level" msgpack:"level"`
	Cooldown    int32  `json:"cooldown" msgpack:"cooldown"` // 冷却时间(ms)
	ManaCost    int32  `json:"manaCost" msgpack:"manaCost"`
	Damage      int32  `json:"damage" msgpack:"damage"`
	Description string `json:"description" msgpack:"description"`
}

// Buff 增益/减益效果。
type Buff struct {
	ID       int32  `json:"id" msgpack:"id"`
	Name     string `json:"name" msgpack:"name"`
	Duration int32  `json:"duration" msgpack:"duration"` // 持续时间(ms)
	Type     int32  `json:"type" msgpack:"type"`         // 1:增益 2:减益
	Value    int32  `json:"value" msgpack:"value"`
}

// CastRequest 释放技能请求。
type CastRequest struct {
	SkillID  int32 `json:"skillId" msgpack:"skillId"`
	TargetID int64 `json:"targetId" msgpack:"targetId"`
}

// CastResponse 释放技能响应。
type CastResponse struct {
	SkillID  int32 `json:"skillId" msgpack:"skillId"`
	CasterID int64 `json:"casterId" msgpack:"casterId"`
	TargetID int64 `json:"targetId" msgpack:"targetId"`
	Damage   int32 `json:"damage" msgpack:"damage"`
	Critical bool  `json:"critical" msgpack:"critical"`
}

// MoveRequest 移动请求。
type MoveRequest struct {
	X float32 `json:"x" msgpack:"x"`
	Y float32 `json:"y" msgpack:"y"`
	Z float32 `json:"z" msgpack:"z"`
	R float32 `json:"r" msgpack:"r"` // 朝向
}

// DamageEvent 伤害事件（服务器推送）。
type DamageEvent struct {
	AttackerID int64 `json:"attackerId" msgpack:"attackerId"`
	TargetID   int64 `json:"targetId" msgpack:"targetId"`
	SkillID    int32 `json:"skillId" msgpack:"skillId"`
	Damage     int32 `json:"damage" msgpack:"damage"`
	Critical   bool  `json:"critical" msgpack:"critical"`
	HPRemain   int32 `json:"hpRemain" msgpack:"hpRemain"`
}
