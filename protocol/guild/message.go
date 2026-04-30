// Package guild 定义公会相关消息类型。
package guild

// GuildInfo 公会信息。
type GuildInfo struct {
	ID          int64  `json:"id" msgpack:"id"`
	Name        string `json:"name" msgpack:"name"`
	Level       int32  `json:"level" msgpack:"level"`
	Exp         int64  `json:"exp" msgpack:"exp"`
	OwnerID     int64  `json:"ownerId" msgpack:"ownerId"`
	OwnerName   string `json:"ownerName" msgpack:"ownerName"`
	MemberCount int32  `json:"memberCount" msgpack:"memberCount"`
	MaxMembers  int32  `json:"maxMembers" msgpack:"maxMembers"`
	Notice      string `json:"notice" msgpack:"notice"`
	Gold        int32  `json:"gold" msgpack:"gold"`
	CreatedAt   int64  `json:"createdAt" msgpack:"createdAt"`
}

// Member 公会成员信息。
type Member struct {
	PlayerID int64  `json:"playerId" msgpack:"playerId"`
	Nickname string `json:"nickname" msgpack:"nickname"`
	Level    int32  `json:"level" msgpack:"level"`
	Position int32  `json:"position" msgpack:"position"` // 0:成员 1:长老 2:副会 3:会长
	Donate   int64  `json:"donate" msgpack:"donate"`
	JoinedAt int64  `json:"joinedAt" msgpack:"joinedAt"`
}

// CreateRequest 创建公会请求。
type CreateRequest struct {
	Name string `json:"name" msgpack:"name"`
}

// CreateResponse 创建公会响应。
type CreateResponse struct {
	Guild *GuildInfo `json:"guild" msgpack:"guild"`
}

// JoinRequest 加入公会请求。
type JoinRequest struct {
	GuildID int64 `json:"guildId" msgpack:"guildId"`
}

// LeaveRequest 退出公会请求。
type LeaveRequest struct {
	GuildID int64 `json:"guildId" msgpack:"guildId"`
}

// KickRequest 踢出成员请求。
type KickRequest struct {
	GuildID  int64 `json:"guildId" msgpack:"guildId"`
	PlayerID int64 `json:"playerId" msgpack:"playerId"`
}

// ListRequest 公会列表请求。
type ListRequest struct {
	Page     int32 `json:"page" msgpack:"page"`
	PageSize int32 `json:"pageSize" msgpack:"pageSize"`
}

// ListResponse 公会列表响应。
type ListResponse struct {
	Guilds []*GuildInfo `json:"guilds" msgpack:"guilds"`
	Total  int32        `json:"total" msgpack:"total"`
}

// InfoRequest 公会详情请求。
type InfoRequest struct {
	GuildID int64 `json:"guildId" msgpack:"guildId"`
}

// InfoResponse 公会详情响应。
type InfoResponse struct {
	Guild   *GuildInfo `json:"guild" msgpack:"guild"`
	Members []*Member  `json:"members" msgpack:"members"`
}

// DonateRequest 公会捐献请求。
type DonateRequest struct {
	GuildID int64 `json:"guildId" msgpack:"guildId"`
	Gold    int32 `json:"gold" msgpack:"gold"`
}

// DonateResponse 公会捐献响应。
type DonateResponse struct {
	GuildExp   int64 `json:"guildExp" msgpack:"guildExp"`
	Contribute int64 `json:"contribute" msgpack:"contribute"`
}
