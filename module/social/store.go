package social

import "context"

type FriendInfo struct {
	PlayerID  int64
	Nickname  string
	Level     int32
	Avatar    string
	Online    bool
	GuildName string
	Intimacy  int32
}

type Store interface {
	AddFriend(ctx context.Context, uid int64, friend *FriendInfo) error
	RemoveFriend(ctx context.Context, uid int64, friendID int64) error
	ListFriends(ctx context.Context, uid int64, page, pageSize int32) ([]*FriendInfo, int32, int32, error)
	BlockUser(ctx context.Context, uid int64, targetID int64) error
	UnblockUser(ctx context.Context, uid int64, targetID int64) error
	ListBlacklist(ctx context.Context, uid int64, page, pageSize int32) ([]*FriendInfo, int32, error)
}
