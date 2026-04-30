// Package player 定义玩家相关消息类型。
package player

// PlayerInfo 玩家基础信息。
type PlayerInfo struct {
	ID        int64  `json:"id" msgpack:"id"`
	Nickname  string `json:"nickname" msgpack:"nickname"`
	Level     int32  `json:"level" msgpack:"level"`
	Exp       int64  `json:"exp" msgpack:"exp"`
	Avatar    string `json:"avatar" msgpack:"avatar"`
	Gold      int32  `json:"gold" msgpack:"gold"`
	Diamond   int32  `json:"diamond" msgpack:"diamond"`
	CreatedAt int64  `json:"createdAt" msgpack:"createdAt"`
}

// GetInfoRequest 获取玩家信息请求。
type GetInfoRequest struct {
	PlayerID int64 `json:"playerId" msgpack:"playerId"`
}

// GetInfoResponse 获取玩家信息响应。
type GetInfoResponse struct {
	Player *PlayerInfo `json:"player" msgpack:"player"`
}

// UpdateProfileRequest 更新玩家资料请求。
type UpdateProfileRequest struct {
	Nickname string `json:"nickname" msgpack:"nickname"`
	Avatar   string `json:"avatar" msgpack:"avatar"`
}

// LevelUpEvent 升级事件（服务器推送）。
type LevelUpEvent struct {
	OldLevel int32 `json:"oldLevel" msgpack:"oldLevel"`
	NewLevel int32 `json:"newLevel" msgpack:"newLevel"`
	ExpGain  int64 `json:"expGain" msgpack:"expGain"`
}
