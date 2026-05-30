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
//	  actor.InvokePlayer(proxy, uid, func() { playerSvc.DeductGold(uid, 100) })
//
//	模式2 (同步请求-响应，通过 Actor mailbox):
//	  模块 Init 时：
//	    ① proxy.AddRouteHandler(route, actor.RouteToActor(actor.KindPlayer), opts)
//	    ② actor.RegisterRouteInitializer(func(act) { act.AddRouteHandler(route, handler) })
//	  消息处理：
//	    Node router → RouteToActor → Actor mailbox → Actor.handler → ctx.Response()
package actor

import (
	"context"
	"strconv"
	"sync"

	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"
)

const KindPlayer = "player"

// RouteInitializer 在 Actor 创建后调用，用于注册路由和事件处理器。
// 此回调在 Actor 的 Init 阶段、Start 之前执行。
type RouteInitializer func(actor *node.Actor)

var (
	mu                sync.Mutex
	routeInitializers []RouteInitializer
)

// RegisterRouteInitializer 注册一个 Actor 路由初始化器。
// 在模块的 Init() 中调用。SpawnPlayer 时会应用所有已注册的初始化器。
func RegisterRouteInitializer(fn RouteInitializer) {
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
			initializers := make([]RouteInitializer, len(routeInitializers))
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

// GetPlayer 获取玩家的 Actor。不存在返回 nil。
func GetPlayer(proxy *node.Proxy, uid int64) *node.Actor {
	act, ok := proxy.Actor(KindPlayer, strconv.FormatInt(uid, 10))
	if !ok {
		return nil
	}
	return act
}

// InvokePlayer 向玩家 Actor 投递函数，在 Actor goroutine 中串行执行。
// Actor 不存在时静默丢弃（玩家已断线）。
// Actor 存在但归属权不属于本节点时，杀死残留 Actor 并丢弃。
func InvokePlayer(proxy *node.Proxy, uid int64, fn func()) {
	// 防御性检查：玩家是否仍绑定在本节点
	if nid, err := proxy.LocateNode(context.Background(), uid, proxy.GetName()); err == nil && nid != "" && nid != proxy.GetID() {
		log.Warnf("[actor] invoke dropped, ownership lost: uid=%d bound_nid=%s my_nid=%s", uid, nid, proxy.GetID())
		KillPlayer(proxy, uid)
		return
	}

	if act := GetPlayer(proxy, uid); act != nil {
		act.Invoke(fn)
	}
}

// RouteToActor 返回一个 Node 路由处理器，将消息投递到 Actor mailbox 串行处理。
//
// 归属权由 due 框架的 StatefulRoute 保证——gate 通过 Locator 定位玩家
// 绑定节点后才投递，消息不会到达旧节点。
//
// 配合 StatefulAuthorizedRoute 使用。使用 RouteToActor 时，必须同时调用
// RegisterRouteInitializer 为 Actor 注册实际的路由处理器。
func RouteToActor(kind string) node.RouteHandler {
	return func(ctx node.Context) {
		uid := ctx.UID()
		if uid == 0 {
			log.Warnf("[actor] uid == 0")
			return
		}

		// 归属权由 due 的 StatefulRoute 保证——gate 通过 Locator 定位
		// 玩家绑定节点后才投递消息，消息不会到达旧节点。
		proxy := ctx.Proxy()
		id := strconv.FormatInt(uid, 10)

		act, ok := proxy.Actor(kind, id)
		if !ok {
			log.Warnf("[actor] actor %s not found uid: %d", kind, uid)
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
