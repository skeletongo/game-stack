package domain

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateToken 生成 32 字节随机令牌（hex 编码）。
// 用于登录和令牌刷新。
func GenerateToken() Token {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return Token(hex.EncodeToString(b))
}
