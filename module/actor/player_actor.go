// Package actor 封装 due Actor 模型，为玩家状态管理提供串行化保障。
//
// 核心机制：
//  1. 每个在线玩家对应一个 PlayerActor。登录时 Spawn，断线时 Kill。
//  2. Kill 时 mailbox 和 fnChan 关闭，所有排队消息被丢弃 → 消除断线幽灵消息。
//  3. 模块通过 RegisterRouteInitializer 注册 Actor 上的路由处理器。
//     Spawn 时自动应用所有已注册的初始化器。
//
// 两种使用模式：
//
//	模式1 (fire-and-forget):
//	  actor.Invoke(proxy, uid, func() { playerSvc.DeductGold(uid, 100) })
//
//	模式2 (同步请求-响应，通过 Actor mailbox):
//	  模块 Init 时：
//	    ① proxy.AddRouteHandler(route, actor.RouteToActor(actor.KindPlayer), opts)
//	    ② actor.RegisterRouteInitializer(func(act) { act.AddRouteHandler(route, handler) })
//	  消息处理：
//	    Node router → RouteToActor → Actor mailbox → Actor.handler → ctx.Response()
package actor

import (
	"strconv"
	"sync"

	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/stack"
)

const KindPlayer = "player"

// ActorRouteInitializer 在 Actor 创建后调用，用于注册路由和事件处理器。
// 此回调在 Actor 的 Init 阶段、Start 之前执行。
type ActorRouteInitializer func(actor *node.Actor)

var (
	mu                sync.Mutex
	routeInitializers []ActorRouteInitializer
)

// RegisterRouteInitializer 注册一个 Actor 路由初始化器。
// 在模块的 Init() 中调用。SpawnPlayer 时会应用所有已注册的初始化器。
func RegisterRouteInitializer(fn ActorRouteInitializer) {
	mu.Lock()
	defer mu.Unlock()
	routeInitializers = append(routeInitializers, fn)
}

// SpawnPlayer 为玩家创建并绑定 Actor。登录成功后调用。
func SpawnPlayer(proxy *node.Proxy, uid int64) (*node.Actor, error) {
	id := strconv.FormatInt(uid, 10)

	act, err := proxy.Spawn(
		func(actor *node.Actor, args ...any) node.Processor {
			proc := &playerProc{actor: actor, uid: uid}

			// 应用所有模块注册的 Actor 路由初始化器
			mu.Lock()
			initializers := make([]ActorRouteInitializer, len(routeInitializers))
			copy(initializers, routeInitializers)
			mu.Unlock()

			for _, fn := range initializers {
				fn(actor)
			}

			return proc
		},
		node.WithActorKind(KindPlayer),
		node.WithActorID(id),
	)
	if err != nil {
		log.Errorf("[actor] spawn player actor failed: uid=%d err=%v", uid, err)
		return nil, err
	}

	if err := proxy.BindActor(uid, KindPlayer, id); err != nil {
		log.Errorf("[actor] bind player actor failed: uid=%d err=%v", uid, err)
		act.Destroy()
		return nil, err
	}

	log.Infof("[actor] spawned: uid=%d pid=%s routes=%d", uid, act.PID(), len(routeInitializers))
	return act, nil
}

// KillPlayer 杀死玩家 Actor。断线/登出时调用。
// mailbox 和 fnChan 被关闭，所有排队消息/Invoke 被丢弃。
func KillPlayer(proxy *node.Proxy, uid int64) {
	id := strconv.FormatInt(uid, 10)

	proxy.UnbindActor(uid, KindPlayer)

	if proxy.Kill(KindPlayer, id) {
		log.Infof("[actor] killed: uid=%d", uid)
	}
}

// Get 获取玩家的 Actor。不存在返回 nil。
func Get(proxy *node.Proxy, uid int64) *node.Actor {
	act, ok := proxy.Actor(KindPlayer, strconv.FormatInt(uid, 10))
	if !ok {
		return nil
	}
	return act
}

// Invoke 向玩家 Actor 投递函数，在 Actor goroutine 中串行执行。
// Actor 不存在时静默丢弃（玩家已断线）。
func Invoke(proxy *node.Proxy, uid int64, fn func()) {
	if act := Get(proxy, uid); act != nil {
		act.Invoke(fn)
	}
}

// RouteToActor 返回一个 Node 路由处理器，将消息投递到 Actor mailbox 串行处理。
// 配合 StatefulAuthorizedRoute，消息先路由到玩家节点，再投递到 Actor。
// ctx.Response() 会在 Actor 的 dispatch goroutine 中执行。
//
// 使用 RouteToActor 时，必须同时调用 RegisterRouteInitializer
// 为 Actor 注册实际的路由处理器。
func RouteToActor(kind string) node.RouteHandler {
	return func(ctx node.Context) {
		uid := ctx.UID()
		if uid == 0 {
			stack.RespondError(ctx, stack.ErrUnauthorized)
			return
		}
		act, ok := ctx.Proxy().Actor(kind, strconv.FormatInt(uid, 10))
		if !ok {
			stack.RespondError(ctx, stack.ErrPlayerNotFound)
			return
		}
		act.Next(ctx)
	}
}

// playerProc 是 Actor 的轻量 Processor。
type playerProc struct {
	node.BaseProcessor
	actor *node.Actor
	uid   int64
}

func (p *playerProc) Init()  { log.Debugf("[actor] init: uid=%d", p.uid) }
func (p *playerProc) Start() { log.Debugf("[actor] start: uid=%d", p.uid) }
func (p *playerProc) Destroy() {
	log.Infof("[actor] destroyed: uid=%d", p.uid)
}
