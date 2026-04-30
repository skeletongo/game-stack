package room

import "context"

// Room 房间数据。
type Room struct {
	ID         int64
	Name       string
	OwnerID    int64
	MaxPlayers int32
	CurPlayers int32
	SceneID    string
	Locked     bool
	Password   string
	Players    []*RoomPlayer
	CreatedAt  int64
}

// RoomPlayer 房间内的玩家。
type RoomPlayer struct {
	PlayerID int64
	Nickname string
	Ready    bool
	JoinedAt int64
}

// Store 房间模块数据存储接口。
type Store interface {
	CreateRoom(ctx context.Context, room *Room) error
	GetRoom(ctx context.Context, roomID int64) (*Room, error)
	ListRooms(ctx context.Context, sceneID string, page, pageSize int32) ([]*Room, int32, error)
	DeleteRoom(ctx context.Context, roomID int64) error
	AddPlayer(ctx context.Context, roomID int64, player *RoomPlayer) error
	RemovePlayer(ctx context.Context, roomID int64, playerID int64) error
	SetReady(ctx context.Context, roomID int64, playerID int64) error
}
