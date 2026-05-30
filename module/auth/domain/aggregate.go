package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/skeletongo/game-stack/ddd"
	"github.com/skeletongo/game-stack/stack"
)

// Account 是用户账户聚合根。
//
// 聚合不变量：
//   - Token 非空 == 在线（OnlineGID 非空）
//   - 被封禁的账户不允许登录
//   - 密码哈希永不为空
type Account struct {
	id           UserID
	username     Username
	passwordHash PasswordHash
	nickname     Nickname
	bannedAt     int64 // 0 表示未封禁，>0 为封禁时间戳
	token        Token
	onlineGID    string // 非空表示在线，值为 Gate 节点 ID
	createdAt    int64
}

var _ ddd.Aggregate = (*Account)(nil)

// NewAccount 创建新账户（工厂方法）。
// 所有值对象在构造时校验，密码自动哈希。
func NewAccount(id int64, username, password, nickname string) (*Account, error) {
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
		username:     un,
		passwordHash: hashPassword(password),
		nickname:     nn,
		createdAt:    time.Now().Unix(),
	}, nil
}

// ReconstructAccount 从持久化数据重建账户（仓储专用，跳过校验）。
func ReconstructAccount(id int64, username, passwordHash, nickname string, token string, onlineGID string, bannedAt, createdAt int64) *Account {
	return &Account{
		id:           UserID(id),
		username:     Username(username),
		passwordHash: PasswordHash(passwordHash),
		nickname:     Nickname(nickname),
		bannedAt:     bannedAt,
		token:        Token(token),
		onlineGID:    onlineGID,
		createdAt:    createdAt,
	}
}

// ID 返回聚合唯一标识。
func (a *Account) ID() int64 { return a.id.Int64() }

// 查询方法

func (a *Account) Username() Username { return a.username }
func (a *Account) Nickname() Nickname { return a.nickname }
func (a *Account) Token() Token       { return a.token }
func (a *Account) OnlineGID() string  { return a.onlineGID }
func (a *Account) BannedAt() int64    { return a.bannedAt }
func (a *Account) CreatedAt() int64   { return a.createdAt }

// IsBanned 账户是否被封禁。
func (a *Account) IsBanned() bool { return a.bannedAt > 0 }

// IsOnline 账户是否在线（有活跃的 Gate 连接）。
func (a *Account) IsOnline() bool { return a.onlineGID != "" }

// 行为方法

// VerifyPassword 验证密码是否匹配。
func (a *Account) VerifyPassword(password string) bool {
	return a.passwordHash == hashPassword(password)
}

// Login 登录：设置令牌和在线状态。
// 前置条件：账户未被封禁。
func (a *Account) Login(token Token, gid string) error {
	if a.IsBanned() {
		return stack.ErrAccountBanned
	}
	a.token = token
	a.onlineGID = gid
	return nil
}

// Logout 登出：清除令牌和在线状态。
func (a *Account) Logout() {
	a.token = ""
	a.onlineGID = ""
}

// RefreshToken 刷新令牌。
func (a *Account) RefreshToken(newToken Token) {
	a.token = newToken
}

// Ban 封禁账户，同时强制登出。
func (a *Account) Ban() {
	a.bannedAt = time.Now().Unix()
	a.token = ""
	a.onlineGID = ""
}

// Unban 解封账户。
func (a *Account) Unban() {
	a.bannedAt = 0
}

// SetOnline 设置在线 Gate 节点（Connect 事件触发）。
func (a *Account) SetOnline(gid string) {
	a.onlineGID = gid
}

// SetOffline 设置离线状态（Disconnect 事件触发）。
func (a *Account) SetOffline() {
	a.onlineGID = ""
}

// hashPassword 对密码做 SHA256 哈希。
func hashPassword(password string) PasswordHash {
	h := sha256.Sum256([]byte(password))
	return PasswordHash(hex.EncodeToString(h[:]))
}
