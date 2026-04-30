// Package auth 定义登录认证相关消息类型。
package auth

// LoginRequest 登录请求。
type LoginRequest struct {
	Username string `json:"username" msgpack:"username"`
	Password string `json:"password" msgpack:"password"`
	Platform string `json:"platform" msgpack:"platform"`
	DeviceID string `json:"deviceId" msgpack:"deviceId"`
}

// LoginResponse 登录响应。
type LoginResponse struct {
	Token       string `json:"token" msgpack:"token"`
	PlayerID    int64  `json:"playerId" msgpack:"playerId"`
	ExpiresAt   int64  `json:"expiresAt" msgpack:"expiresAt"`
	IsNewPlayer bool   `json:"isNewPlayer" msgpack:"isNewPlayer"`
}

// RegisterRequest 注册请求。
type RegisterRequest struct {
	Username string `json:"username" msgpack:"username"`
	Password string `json:"password" msgpack:"password"`
	Nickname string `json:"nickname" msgpack:"nickname"`
}

// RegisterResponse 注册响应。
type RegisterResponse struct {
	Token    string `json:"token" msgpack:"token"`
	PlayerID int64  `json:"playerId" msgpack:"playerId"`
}

// LogoutRequest 登出请求。
type LogoutRequest struct{}

// TokenRefreshRequest Token 刷新请求。
type TokenRefreshRequest struct {
	Token string `json:"token" msgpack:"token"`
}

// TokenRefreshResponse Token 刷新响应。
type TokenRefreshResponse struct {
	Token     string `json:"token" msgpack:"token"`
	ExpiresAt int64  `json:"expiresAt" msgpack:"expiresAt"`
}
