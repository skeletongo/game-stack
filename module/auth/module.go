// Package auth 是认证限界上下文。
//
// 职责：用户注册、登录/登出、令牌管理、连接生命周期事件处理。
//
// 战略分类：通用域（Generic）—— 标准化认证能力，可替换为第三方方案。
//
// 聚合根：Account（用户凭证 + 会话状态）。
package auth

import (
	"github.com/dobyte/due/v2/cluster"
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/ddd"
	"github.com/skeletongo/game-stack/module/auth/internal/application"
	"github.com/skeletongo/game-stack/module/auth/internal/domain"
	interfaces "github.com/skeletongo/game-stack/module/auth/internal/interface"
	svcserver "github.com/skeletongo/game-stack/module/auth/internal/svc"
	"github.com/skeletongo/game-stack/stack"
	"github.com/skeletongo/game-stack/stack/debug"
)

const name = "auth"

// Module 创建 Auth 模块（DDD 四层架构）。
//
// 使用方式：
//
//	stack.WithModules(auth.Module(auth.WithRepository(myRedisRepo)))
func Module(opts ...Option) stack.Module {
	return &authModule{opts: opts}
}

type authModule struct {
	opts []Option
}

func (m *authModule) Name() string { return name }

// Init 初始化模块：装配四层依赖、注册路由和事件处理器。
func (m *authModule) Init(proxy *node.Proxy) error {
	o := defaultOptions()
	for _, opt := range m.opts {
		opt(o)
	}

	// ---- 基础设施层 ----
	repo := o.repo

	// ---- 领域层 ----
	eventBus := ddd.NewEventBus()
	cmdBus := ddd.NewCommandBus()

	// ---- 应用层：命令处理器 ----
	ddd.Register(cmdBus, application.CmdRegister, &application.RegisterHandler{Repo: repo, EventBus: eventBus})
	ddd.Register(cmdBus, application.CmdLogin, &application.LoginHandler{Repo: repo, EventBus: eventBus})
	ddd.Register(cmdBus, application.CmdMarkOnline, &application.MarkOnlineHandler{Repo: repo, EventBus: eventBus})
	ddd.Register(cmdBus, application.CmdLogout, &application.LogoutHandler{Repo: repo, EventBus: eventBus})
	ddd.Register(cmdBus, application.CmdRefreshToken, &application.RefreshTokenHandler{Repo: repo})

	// ---- 接口层：路由 + 事件处理器 ----
	routes := interfaces.NewHandlers(proxy, cmdBus)

	// 无状态路由（无需认证）
	proxy.AddRouteHandler(stack.RouteAuthLogin, routes.HandleLogin)
	proxy.AddRouteHandler(stack.RouteAuthRegister, routes.HandleRegister)

	// 有状态 + 认证路由
	proxy.AddRouteHandler(stack.RouteAuthLogout, routes.HandleLogout, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteAuthTokenRefresh, routes.HandleRefresh, stack.StatefulAuthorizedRoute)

	// 注册玩家生命周期回调（断线标记离线）
	stack.AddEventHandler(cluster.Disconnect, func(ctx node.Context) {
		uid := ctx.UID()

		gid, err := ctx.Proxy().LocateGate(ctx.Context(), uid)
		if err != nil {
			log.Warnf("[auth] disconnect gate check failed: uid=%d err=%v", uid, err)
			return
		}
		if gid != "" && gid != ctx.GID() {
			log.Debugf("[auth] stale disconnect ignored: uid=%d current_gid=%s event_gid=%s", uid, gid, ctx.GID())
			return
		}

		nid, err := ctx.Proxy().LocateNode(ctx.Context(), uid, ctx.Proxy().GetName())
		if err != nil {
			log.Warnf("[auth] disconnect ownership check failed: uid=%d err=%v", uid, err)
			return
		}

		if nid != ctx.Proxy().GetID() {
			return
		}

		account, err := repo.Load(ctx.Context(), uid)
		if err != nil {
			return
		}
		account.SetOffline()
		if err = repo.Save(ctx.Context(), account); err != nil {
			return
		}

		eventBus.Publish(domain.NewAccountDisconnect(uid))
	})

	// 注册跨模块 Service（供其他模块通过 stack.GetService("auth") 获取）
	stack.RegisterService(name, svcserver.New(repo))

	// 注册到 debug 服务（运行时查询/修改数据）
	debug.Register[*domain.Account](name, repo, cmdBus)

	log.Infof("[auth] module initialized (DDD)")
	return nil
}
