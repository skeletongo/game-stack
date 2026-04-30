// Package social 定义社交/好友相关消息类型。
package social

// FriendInfo 好友信息。
type FriendInfo struct {
	PlayerID  int64  `json:"playerId" msgpack:"playerId"`
	Nickname  string `json:"nickname" msgpack:"nickname"`
	Level     int32  `json:"level" msgpack:"level"`
	Avatar    string `json:"avatar" msgpack:"avatar"`
	Online    bool   `json:"online" msgpack:"online"`
	GuildName string `json:"guildName" msgpack:"guildName"`
	Intimacy  int32  `json:"intimacy" msgpack:"intimacy"`
}

// FriendListRequest 获取好友列表请求。
type FriendListRequest struct {
	Page     int32 `json:"page" msgpack:"page"`
	PageSize int32 `json:"pageSize" msgpack:"pageSize"`
}

// FriendListResponse 获取好友列表响应。
type FriendListResponse struct {
	Friends  []*FriendInfo `json:"friends" msgpack:"friends"`
	Total    int32         `json:"total" msgpack:"total"`
	MaxCount int32         `json:"maxCount" msgpack:"maxCount"`
}

// AddFriendRequest 添加好友请求。
type AddFriendRequest struct {
	PlayerID int64  `json:"playerId" msgpack:"playerId"`
	Message  string `json:"message" msgpack:"message"`
}

// RemoveFriendRequest 删除好友请求。
type RemoveFriendRequest struct {
	PlayerID int64 `json:"playerId" msgpack:"playerId"`
}

// BlockRequest 拉黑请求。
type BlockRequest struct {
	PlayerID int64 `json:"playerId" msgpack:"playerId"`
}

// UnblockRequest 解除拉黑请求。
type UnblockRequest struct {
	PlayerID int64 `json:"playerId" msgpack:"playerId"`
}

// BlacklistRequest 获取黑名单请求。
type BlacklistRequest struct {
	Page     int32 `json:"page" msgpack:"page"`
	PageSize int32 `json:"pageSize" msgpack:"pageSize"`
}

// BlacklistResponse 获取黑名单响应。
type BlacklistResponse struct {
	Blacklist []*FriendInfo `json:"blacklist" msgpack:"blacklist"`
	Total     int32         `json:"total" msgpack:"total"`
}

// FriendAddEvent 好友添加事件（服务器推送）。
type FriendAddEvent struct {
	Friend *FriendInfo `json:"friend" msgpack:"friend"`
}
