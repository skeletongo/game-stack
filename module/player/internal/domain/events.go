package domain

import "time"

// 领域事件名称常量。
const (
	EventPlayerLeveledUp      = "player.leveled_up"
	EventPlayerProfileUpdated = "player.profile_updated"
	EventGoldChanged          = "player.gold_changed"
	EventDiamondChanged       = "player.diamond_changed"
)

// PlayerLeveledUp 玩家升级事件。
type PlayerLeveledUp struct {
	playerID   int64
	oldLevel   int32
	newLevel   int32
	occurredAt time.Time
}

func NewPlayerLeveledUp(playerID int64, oldLevel, newLevel int32) PlayerLeveledUp {
	return PlayerLeveledUp{playerID: playerID, oldLevel: oldLevel, newLevel: newLevel, occurredAt: time.Now()}
}
func (e PlayerLeveledUp) AggregateID() int64    { return e.playerID }
func (e PlayerLeveledUp) EventName() string     { return EventPlayerLeveledUp }
func (e PlayerLeveledUp) OccurredAt() time.Time { return e.occurredAt }
func (e PlayerLeveledUp) OldLevel() int32       { return e.oldLevel }
func (e PlayerLeveledUp) NewLevel() int32       { return e.newLevel }

// PlayerProfileUpdated 玩家资料更新事件。
type PlayerProfileUpdated struct {
	playerID   int64
	occurredAt time.Time
}

func NewPlayerProfileUpdated(playerID int64) PlayerProfileUpdated {
	return PlayerProfileUpdated{playerID: playerID, occurredAt: time.Now()}
}
func (e PlayerProfileUpdated) AggregateID() int64    { return e.playerID }
func (e PlayerProfileUpdated) EventName() string     { return EventPlayerProfileUpdated }
func (e PlayerProfileUpdated) OccurredAt() time.Time { return e.occurredAt }

// GoldChanged 金币变动事件。
type GoldChanged struct {
	playerID   int64
	delta      int32
	after      int32
	occurredAt time.Time
}

func NewGoldChanged(playerID int64, delta, after int32) GoldChanged {
	return GoldChanged{playerID: playerID, delta: delta, after: after, occurredAt: time.Now()}
}
func (e GoldChanged) AggregateID() int64    { return e.playerID }
func (e GoldChanged) EventName() string     { return EventGoldChanged }
func (e GoldChanged) OccurredAt() time.Time { return e.occurredAt }
func (e GoldChanged) Delta() int32          { return e.delta }
func (e GoldChanged) After() int32          { return e.after }

// DiamondChanged 钻石变动事件。
type DiamondChanged struct {
	playerID   int64
	delta      int32
	after      int32
	occurredAt time.Time
}

func NewDiamondChanged(playerID int64, delta, after int32) DiamondChanged {
	return DiamondChanged{playerID: playerID, delta: delta, after: after, occurredAt: time.Now()}
}
func (e DiamondChanged) AggregateID() int64    { return e.playerID }
func (e DiamondChanged) EventName() string     { return EventDiamondChanged }
func (e DiamondChanged) OccurredAt() time.Time { return e.occurredAt }
func (e DiamondChanged) Delta() int32          { return e.delta }
func (e DiamondChanged) After() int32          { return e.after }
