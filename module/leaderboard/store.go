package leaderboard

import "context"

type Entry struct {
	Rank     int32
	PlayerID int64
	Nickname string
	Score    int64
	Level    int32
}

type Store interface {
	UpdateScore(ctx context.Context, boardName string, uid int64, nickname string, score int64, level int32) error
	GetRank(ctx context.Context, boardName string, uid int64) (*Entry, error)
	GetTop(ctx context.Context, boardName string, page, pageSize int32) ([]*Entry, int32, error)
}
