package match

import "context"

// Store 匹配模块数据存储接口（队列状态）。
type Store interface {
	Push(ctx context.Context, uid int64, info *MatchInfo) error
	Pop(ctx context.Context, uid int64) error
	FindMatch(ctx context.Context, info *MatchInfo) ([]int64, error)
	GetStatus(ctx context.Context, uid int64) (*MatchInfo, error)
}

// MatchInfo 匹配队列信息。
type MatchInfo struct {
	UID       int64
	MatchType string
	Mode      string
	Rank      int32
	MaxRank   int32
	MinRank   int32
	JoinedAt  int64
}
