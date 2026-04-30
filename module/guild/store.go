package guild

import "context"

type Guild struct {
	ID          int64
	Name        string
	Level       int32
	Exp         int64
	OwnerID     int64
	OwnerName   string
	MemberCount int32
	MaxMembers  int32
	Notice      string
	Gold        int32
	Members     []*Member
	CreatedAt   int64
}

type Member struct {
	PlayerID int64
	Nickname string
	Level    int32
	Position int32
	Donate   int64
	JoinedAt int64
}

type Store interface {
	CreateGuild(ctx context.Context, guild *Guild) error
	GetGuild(ctx context.Context, guildID int64) (*Guild, error)
	ListGuilds(ctx context.Context, page, pageSize int32) ([]*Guild, int32, error)
	DeleteGuild(ctx context.Context, guildID int64) error
	AddMember(ctx context.Context, guildID int64, member *Member) error
	RemoveMember(ctx context.Context, guildID int64, playerID int64) error
	UpdateMemberPosition(ctx context.Context, guildID int64, playerID int64, position int32) error
	Donate(ctx context.Context, guildID int64, playerID int64, gold int32) (int64, error)
}
