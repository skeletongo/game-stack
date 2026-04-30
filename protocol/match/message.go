// Package match 定义匹配系统相关消息类型。
package match

// JoinRequest 加入匹配请求。
type JoinRequest struct {
	MatchType string `json:"matchType" msgpack:"matchType"` // pvp, pve, custom
	Mode      string `json:"mode" msgpack:"mode"`           // 1v1, 3v3, 5v5
	MaxRank   int32  `json:"maxRank" msgpack:"maxRank"`
	MinRank   int32  `json:"minRank" msgpack:"minRank"`
}

// JoinResponse 加入匹配响应。
type JoinResponse struct {
	QueueID   string `json:"queueId" msgpack:"queueId"`
	QueueSize int32  `json:"queueSize" msgpack:"queueSize"`
}

// CancelRequest 取消匹配请求。
type CancelRequest struct {
	QueueID string `json:"queueId" msgpack:"queueId"`
}

// MatchResult 匹配结果（服务器推送）。
type MatchResult struct {
	MatchID   string  `json:"matchId" msgpack:"matchId"`
	MatchType string  `json:"matchType" msgpack:"matchType"`
	PlayerIDs []int64 `json:"playerIds" msgpack:"playerIds"`
	RoomAddr  string  `json:"roomAddr" msgpack:"roomAddr"`
}

// StatusRequest 查询匹配状态请求。
type StatusRequest struct {
	QueueID string `json:"queueId" msgpack:"queueId"`
}

// StatusResponse 查询匹配状态响应。
type StatusResponse struct {
	InQueue   bool  `json:"inQueue" msgpack:"inQueue"`
	QueueSize int32 `json:"queueSize" msgpack:"queueSize"`
	WaitTime  int64 `json:"waitTime" msgpack:"waitTime"`
}
