// Package room 定义房间/场景相关消息类型。
package room

// RoomInfo 房间信息。
type RoomInfo struct {
	ID         int64  `json:"id" msgpack:"id"`
	Name       string `json:"name" msgpack:"name"`
	OwnerID    int64  `json:"ownerId" msgpack:"ownerId"`
	MaxPlayers int32  `json:"maxPlayers" msgpack:"maxPlayers"`
	CurPlayers int32  `json:"curPlayers" msgpack:"curPlayers"`
	SceneID    string `json:"sceneId" msgpack:"sceneId"`
	Locked     bool   `json:"locked" msgpack:"locked"`
	Password   string `json:"password,omitempty" msgpack:"password,omitempty"`
}

// CreateRequest 创建房间请求。
type CreateRequest struct {
	Name       string `json:"name" msgpack:"name"`
	MaxPlayers int32  `json:"maxPlayers" msgpack:"maxPlayers"`
	SceneID    string `json:"sceneId" msgpack:"sceneId"`
	Password   string `json:"password" msgpack:"password"`
}

// CreateResponse 创建房间响应。
type CreateResponse struct {
	Room *RoomInfo `json:"room" msgpack:"room"`
}

// JoinRequest 加入房间请求。
type JoinRequest struct {
	RoomID   int64  `json:"roomId" msgpack:"roomId"`
	Password string `json:"password" msgpack:"password"`
}

// LeaveRequest 离开房间请求。
type LeaveRequest struct {
	RoomID int64 `json:"roomId" msgpack:"roomId"`
}

// ListRequest 房间列表请求。
type ListRequest struct {
	SceneID  string `json:"sceneId" msgpack:"sceneId"`
	Page     int32  `json:"page" msgpack:"page"`
	PageSize int32  `json:"pageSize" msgpack:"pageSize"`
}

// ListResponse 房间列表响应。
type ListResponse struct {
	Rooms []*RoomInfo `json:"rooms" msgpack:"rooms"`
	Total int32       `json:"total" msgpack:"total"`
}

// KickRequest 踢人请求。
type KickRequest struct {
	RoomID   int64 `json:"roomId" msgpack:"roomId"`
	PlayerID int64 `json:"playerId" msgpack:"playerId"`
}

// PlayerJoinedEvent 玩家加入事件（服务器推送）。
type PlayerJoinedEvent struct {
	RoomID   int64  `json:"roomId" msgpack:"roomId"`
	PlayerID int64  `json:"playerId" msgpack:"playerId"`
	Nickname string `json:"nickname" msgpack:"nickname"`
}

// PlayerLeftEvent 玩家离开事件（服务器推送）。
type PlayerLeftEvent struct {
	RoomID   int64 `json:"roomId" msgpack:"roomId"`
	PlayerID int64 `json:"playerId" msgpack:"playerId"`
}
