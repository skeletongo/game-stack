package stack

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"
)

// Middleware 定义中间件处理函数类型。
// 适合直接用于 due 框架的 AddRouteHandler 包装。
type Middleware func(next node.Context) node.Context

// Recovery 返回一个 panic 恢复中间件。
// 在路由处理器发生 panic 时捕获并返回内部错误响应，
// 防止单个请求的 panic 导致整个节点崩溃。
func Recovery() func(ctx node.Context) {
	return func(ctx node.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("panic recovered in route %d: %v", ctx.Route(), r)
				RespondError(ctx, ErrInternalError)
			}
		}()
		// 恢复后不做额外处理，让 due 框架继续执行下一个处理器
	}
}

// AuthRequired 返回一个认证检查中间件。
// 当 UID 为 0 时拒绝请求（表示未登录或网关未绑定用户）。
// 应注册在需要登录态的接口上。
func AuthRequired() func(ctx node.Context) {
	return func(ctx node.Context) {
		if ctx.UID() == 0 {
			RespondError(ctx, ErrUnauthorized)
			return
		}
		// 通过认证，due 框架自动继续执行
	}
}
