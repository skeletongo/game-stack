package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/skeletongo/game-stack/ddd"
)

// Account 是用户账户聚合根。
//
// 聚合不变量：
//   - 被封禁的账户不允许登录
//   - 密码哈希永不为空
type Account struct {
	id           UserID
	playerID     int64
	username     Username
	passwordHash PasswordHash
	nickname     Nickname
	bannedAt     int64 // 0 表示未封禁，>0 为封禁时间戳
	createdAt    int64
}

var _ ddd.Aggregate = (*Account)(nil)

// NewAccount 创建新账户（工厂方法）。
// 所有值对象在构造时校验，密码自动哈希。
func NewAccount(id int64, playerID int64, username, password, nickname string) (*Account, error) {
	un, err := NewUsername(username)
	if err != nil {
		return nil, err
	}
	nn, err := NewNickname(nickname)
	if err != nil {
		return nil, err
	}
	return &Account{
		id:           UserID(id),
		playerID:     playerID,
		username:     un,
		passwordHash: hashPassword(password),
		nickname:     nn,
		createdAt:    time.Now().Unix(),
	}, nil
}

// ReconstructAccount 从持久化数据重建账户（仓储专用，跳过校验）。
func ReconstructAccount(id int64, playerID int64, username, passwordHash, nickname string, bannedAt, createdAt int64) *Account {
	return &Account{
		id:           UserID(id),
		playerID:     playerID,
		username:     Username(username),
		passwordHash: PasswordHash(passwordHash),
		nickname:     Nickname(nickname),
		bannedAt:     bannedAt,
		createdAt:    createdAt,
	}
}

// ID 返回聚合唯一标识。
func (a *Account) ID() int64 { return a.id.Int64() }

// 查询方法

func (a *Account) Username() Username { return a.username }
func (a *Account) PlayerID() int64    { return a.playerID }
func (a *Account) Nickname() Nickname { return a.nickname }
func (a *Account) BannedAt() int64    { return a.bannedAt }
func (a *Account) CreatedAt() int64   { return a.createdAt }

// IsBanned 账户是否被封禁。
func (a *Account) IsBanned() bool { return a.bannedAt > 0 }

// 行为方法

// VerifyPassword 验证密码是否匹配。
func (a *Account) VerifyPassword(password string) bool {
	return a.passwordHash == hashPassword(password)
}

// Ban 封禁账户，同时强制登出。
func (a *Account) Ban() {
	a.bannedAt = time.Now().Unix()
}

// Unban 解封账户。
func (a *Account) Unban() {
	a.bannedAt = 0
}

// hashPassword 对密码做 SHA256 哈希。
func hashPassword(password string) PasswordHash {
	h := sha256.Sum256([]byte(password))
	return PasswordHash(hex.EncodeToString(h[:]))
}
