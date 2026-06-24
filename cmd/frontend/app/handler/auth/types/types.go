package types

// RegisterReq 注册请求参数。
type RegisterReq struct {
	Username string `json:"username"` // 账号名
	Password string `json:"password"` // 密码
	Nickname string `json:"nickname"` // 昵称
}

// RegisterResp 注册响应结果。
type RegisterResp struct {
	Code    int32  `json:"code"`              // 错误码，0 表示成功
	Message string `json:"message,omitempty"` // 错误描述
}

// LoginReq 登录请求参数。
type LoginReq struct {
	Username string `json:"username"` // 账号名
	Password string `json:"password"` // 密码
	GameID   int64  `json:"game_id"`  // 游戏id
}

// LoginResp 登录响应结果。
type LoginResp struct {
	Code      int32  `json:"code"`                 // 错误码，0 表示成功
	Message   string `json:"message,omitempty"`    // 错误描述
	Token     string `json:"token,omitempty"`      // 登录令牌
	ExpiresAt int64  `json:"expires_at,omitempty"` // token过期时间
	PlayerID  int64  `json:"player_id,omitempty"`  // 玩家id
	UnixNano  int64  `json:"unix_nano,omitempty"`  // 服务器时间戳，纳秒
}

// LogoutResp 登出响应结果。
type LogoutResp struct {
	Code    int32  `json:"code"`              // 错误码，0 表示成功
	Message string `json:"message,omitempty"` // 错误描述
}

// TokenRefreshResp 刷新 token 响应结果。
type TokenRefreshResp struct {
	Code      int32  `json:"code"`                 // 错误码，0 表示成功
	Message   string `json:"message,omitempty"`    // 错误描述
	Token     string `json:"token,omitempty"`      // 新登录令牌
	ExpiresAt int64  `json:"expires_at,omitempty"` // token过期时间
}
