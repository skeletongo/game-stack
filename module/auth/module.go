// Package auth 是认证限界上下文。
//
// 职责：用户注册、登录/登出、令牌管理、连接生命周期事件处理。
//
// 战略分类：通用域（Generic）—— 标准化认证能力，可替换为第三方方案。
//
// 聚合根：Account（用户凭证 + 会话状态）。
package auth

import (
	"context"
	"time"

	"github.com/dobyte/due/v2/cluster"
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/ddd"
	"github.com/skeletongo/game-stack/module/auth/application"
	"github.com/skeletongo/game-stack/module/auth/domain"
	interfaces "github.com/skeletongo/game-stack/module/auth/interface"
	"github.com/skeletongo/game-stack/stack"
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

	// ---- 清理器（跨模块共享）----
	cleaner := stack.NewPlayerDoneCleaner(proxy, 30*time.Second, 3)
	stack.RegisterService("cleaner", cleaner)

	// ---- 应用层：命令处理器 ----
	registerHandler := &application.RegisterHandler{Repo: repo, EventBus: eventBus}
	loginHandler := &application.LoginHandler{Repo: repo, EventBus: eventBus}
	logoutHandler := &application.LogoutHandler{Repo: repo, EventBus: eventBus}
	refreshHandler := &application.RefreshTokenHandler{Repo: repo}

	// ---- 接口层：路由 + 事件处理器 ----
	routes := interfaces.NewHandlers(cleaner, proxy, registerHandler, loginHandler, logoutHandler, refreshHandler)

	// 无状态路由（无需认证）
	proxy.AddRouteHandler(stack.RouteAuthLogin, routes.HandleLogin)
	proxy.AddRouteHandler(stack.RouteAuthRegister, routes.HandleRegister)

	// 有状态 + 认证路由
	proxy.AddRouteHandler(stack.RouteAuthLogout, routes.HandleLogout, stack.StatefulAuthorizedRoute)
	proxy.AddRouteHandler(stack.RouteAuthTokenRefresh, routes.HandleRefresh, stack.StatefulAuthorizedRoute)

	// 连接/断开事件（全集群广播）
	proxy.AddEventHandler(cluster.Connect, routes.HandleConnect)
	proxy.AddEventHandler(cluster.Disconnect, routes.HandleDisconnect)

	// 注册清理回调（Grace Period 到期后清除 token）
	cleaner.Register(&cleanableAdapter{repo: repo})

	// 注册跨模块 Service（供其他模块通过 stack.GetService("auth") 获取）
	stack.RegisterService(name, &svcAdapter{repo: repo})

	log.Infof("[auth] module initialized (DDD)")
	return nil
}

// ---- 适配器 ----

// cleanableAdapter 适配 AccountRepository → stack.CleanableService。
type cleanableAdapter struct {
	repo domain.AccountRepository
}

func (a *cleanableAdapter) CleanPlayerData(uid int64) error {
	return a.repo.Delete(context.Background(), uid)
}

// svcAdapter 是 auth 模块对外的服务适配器。
// 其他模块通过 stack.GetService("auth") 获取，类型断言为 *svcAdapter。
//
// 提供的能力：
//   - Authenticate(token) → 验证令牌并返回 userID
//   - IsOnline(uid) → 检查用户是否在线
type svcAdapter struct {
	repo domain.AccountRepository
}

// Authenticate 验证令牌有效性，返回对应的用户 ID。
func (s *svcAdapter) Authenticate(token string) (int64, error) {
	acc, err := s.repo.FindByToken(context.Background(), token)
	if err != nil {
		return 0, err
	}
	return acc.ID(), nil
}

// IsOnline 检查用户是否在线（有活跃的 Gate 连接）。
func (s *svcAdapter) IsOnline(uid int64) bool {
	acc, err := s.repo.Load(context.Background(), uid)
	if err != nil {
		return false
	}
	return acc.IsOnline()
}
