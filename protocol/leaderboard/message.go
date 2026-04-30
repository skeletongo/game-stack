// Package leaderboard 定义排行榜相关消息类型。
package leaderboard

// Entry 排行榜条目。
type Entry struct {
	Rank     int32  `json:"rank" msgpack:"rank"`
	PlayerID int64  `json:"playerId" msgpack:"playerId"`
	Nickname string `json:"nickname" msgpack:"nickname"`
	Score    int64  `json:"score" msgpack:"score"`
	Level    int32  `json:"level" msgpack:"level"`
}

// GetRequest 获取排行榜请求。
type GetRequest struct {
	BoardName string `json:"boardName" msgpack:"boardName"` // level, pvp, wealth, guild
	Page      int32  `json:"page" msgpack:"page"`
	PageSize  int32  `json:"pageSize" msgpack:"pageSize"`
}

// GetResponse 获取排行榜响应。
type GetResponse struct {
	Entries   []*Entry `json:"entries" msgpack:"entries"`
	BoardName string   `json:"boardName" msgpack:"boardName"`
	Total     int32    `json:"total" msgpack:"total"`
}

// RankRequest 查询我的排名请求。
type RankRequest struct {
	BoardName string `json:"boardName" msgpack:"boardName"`
}

// RankResponse 查询我的排名响应。
type RankResponse struct {
	Entry *Entry `json:"entry" msgpack:"entry"`
}

// UpdateEvent 排名更新事件（服务器推送）。
type UpdateEvent struct {
	BoardName string `json:"boardName" msgpack:"boardName"`
	OldRank   int32  `json:"oldRank" msgpack:"oldRank"`
	NewRank   int32  `json:"newRank" msgpack:"newRank"`
	NewScore  int64  `json:"newScore" msgpack:"newScore"`
}
