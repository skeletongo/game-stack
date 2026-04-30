package auth

import "context"

// User 用户账号数据。
type User struct {
	ID        int64
	Username  string
	Password  string
	Nickname  string
	BannedAt  int64
	CreatedAt int64
}

// Store 定义认证模块的数据存储接口。
// 默认使用内存实现，生产环境可注入 Redis/MySQL 实现。
type Store interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, id int64) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	BanUser(ctx context.Context, id int64) error
	UnbanUser(ctx context.Context, id int64) error
	SetToken(ctx context.Context, uid int64, token string) error
	GetToken(ctx context.Context, uid int64) (string, error)
	DeleteToken(ctx context.Context, uid int64) error
	GetTokenByValue(ctx context.Context, token string) (int64, error)
	SetOnline(ctx context.Context, uid int64, gid string) error
	SetOffline(ctx context.Context, uid int64) error
	IsOnline(ctx context.Context, uid int64) (bool, error)
	OnlineCount(ctx context.Context) (int64, error)
}
