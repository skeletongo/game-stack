package svc

import "context"

// IAuth 是 auth 模块对外暴露的跨模块服务接口。
//
// 其他模块通过 stack.GetService("auth") 获取，类型断言为 IAuth 后调用。
type IAuth interface {
	// Authenticate 验证令牌有效性，返回对应的用户 ID。
	// 令牌无效或过期时返回 error。
	Authenticate(ctx context.Context, token string) (int64, error)
	// IsOnline 检查用户是否在线（有活跃的 Gate 连接）。
	IsOnline(ctx context.Context, uid int64) bool
}
