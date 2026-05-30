// Package domain 定义 auth 限界上下文的领域模型。
//
// Account 聚合是核心：封装用户凭证、令牌、在线状态和封禁状态，
// 所有状态变更都通过行为方法执行并维护不变量。
package domain

import (
	"errors"

	"github.com/skeletongo/game-stack/ddd"
)

// 领域错误
var (
	ErrInvalidUsername = errors.New("username must be 1-32 characters")
	ErrInvalidPassword = errors.New("password must be 6-128 characters")
	ErrInvalidNickname = errors.New("nickname must be 1-16 characters")
	ErrInvalidToken    = errors.New("invalid token")

	ErrAccountBanned = errors.New("account is banned")
	ErrWrongPassword = errors.New("wrong password")
	ErrAccountExists = errors.New("account already exists")
)

// UserID 用户唯一标识值对象。
// 强类型防止与 PlayerID 等其他 ID 混淆。
type UserID int64

func (id UserID) Int64() int64 { return int64(id) }
func (id UserID) Equals(other ddd.ValueObject) bool {
	o, ok := other.(UserID)
	return ok && id == o
}
func (id UserID) Scalar() any { return int64(id) }

// Username 用户名值对象（1-32 字符）。
type Username string

// NewUsername 创建用户名，校验长度。
func NewUsername(s string) (Username, error) {
	if len(s) == 0 || len(s) > 32 {
		return "", ErrInvalidUsername
	}
	return Username(s), nil
}

func (u Username) String() string { return string(u) }
func (u Username) Equals(other ddd.ValueObject) bool {
	o, ok := other.(Username)
	return ok && u == o
}
func (u Username) Scalar() any { return string(u) }

// PasswordHash 密码 SHA256 哈希值对象。
// 原始密码从不存储，只保留哈希。
type PasswordHash string

func (p PasswordHash) String() string { return string(p) }
func (p PasswordHash) Equals(other ddd.ValueObject) bool {
	o, ok := other.(PasswordHash)
	return ok && p == o
}
func (p PasswordHash) Scalar() any { return string(p) }

// Nickname 昵称值对象（1-16 字符）。
type Nickname string

// NewNickname 创建昵称，校验长度。
func NewNickname(s string) (Nickname, error) {
	if len(s) == 0 || len(s) > 16 {
		return "", ErrInvalidNickname
	}
	return Nickname(s), nil
}

func (n Nickname) String() string { return string(n) }
func (n Nickname) Equals(other ddd.ValueObject) bool {
	o, ok := other.(Nickname)
	return ok && n == o
}
func (n Nickname) Scalar() any { return string(n) }

// Token 认证令牌值对象（32 字节随机 hex）。
type Token string

func (t Token) String() string { return string(t) }

// IsEmpty 令牌是否为空（离线状态）。
func (t Token) IsEmpty() bool { return t == "" }

func (t Token) Equals(other ddd.ValueObject) bool {
	o, ok := other.(Token)
	return ok && t == o
}
func (t Token) Scalar() any { return string(t) }
