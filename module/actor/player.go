// Package actor 封装 due Actor 模型，为玩家状态管理提供串行化保障。
//
// 核心机制：
//  1. 每个在线玩家对应一个 PlayerActor。登录时 Spawn，断线时 Kill。
//  2. Kill 时 mailbox 和 fnChan 关闭，所有排队消息被丢弃 → 消除断线幽灵消息。
//  3. 模块通过 registerPlayerRouteInitializer 注册 Actor 上的路由处理器。
//     Spawn 时自动应用所有已注册的初始化器。
//
// 三种使用模式：
//
//	模式1 (fire-and-forget):
//	  actor.InvokePlayer(proxy, uid, func() { _ = playerSvc.DeductGold(ctx, uid, 100) })
//
//	模式2 (同步执行并返回结果):
//	  newExp, err := actor.InvokePlayerSync[int64](ctx, proxy, uid, func(ctx context.Context) (int64, error) {
//	      return ddd.Dispatch[int64](ctx, cmdBus, application.AddExpCmd{...})
//	  })
//
//	模式3 (同步请求-响应，通过 Actor mailbox):
//	  模块 Init 时：
//	    ① proxy.AddRouteHandler(route, actor.RouteToActor(actor.KindPlayer), opts)
//	    ② actor.RegisterRouteInitializer(func(act) { act.AddRouteHandler(route, handler) })
//	  消息处理：
//	    Node router → RouteToActor → Actor mailbox → Actor.handler → ctx.Response()
package actor

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/errors"
	"github.com/dobyte/due/v2/log"
)

const KindPlayer = "player"

// playerRouteInitializer 在 Actor 创建后调用，用于注册路由和事件处理器。
// 此回调在 Actor 的 Init 阶段、Start 之前执行。
type playerRouteInitializer func(actor *node.Actor)

var (
	mu                sync.Mutex
	routeInitializers []playerRouteInitializer
)

// registerPlayerRouteInitializer 注册一个 Actor 路由初始化器。
// 在模块的 Init() 中调用。SpawnPlayer 时会应用所有已注册的初始化器。
func registerPlayerRouteInitializer(fn playerRouteInitializer) {
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
			initializers := make([]playerRouteInitializer, len(routeInitializers))
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
		if errors.Is(err, errors.ErrActorExists) {
			act, b := proxy.Actor(KindPlayer, id)
			if !b {
				log.Errorf("[actor] spawn player actor find failed: uid=%d", uid)
				return nil, err
			}
			return act, nil
		}
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
		var err error
		act, err = SpawnPlayer(proxy, uid)
		if err != nil {
			return nil
		}
	}
	return act
}

// InvokePlayer 向玩家 Actor 投递函数，在 Actor goroutine 中串行执行。
// Actor 不存在时自动创建。
// Actor 存在但归属权不属于本节点时，杀死残留 Actor 并丢弃。
// 注意：fire-and-forget，不等待执行结果。
func InvokePlayer(proxy *node.Proxy, uid int64, fn func()) {
	// 防御性检查：玩家是否仍绑定在本节点
	if nid, err := proxy.LocateNode(context.Background(), uid, proxy.GetName()); err == nil && nid != "" && nid != proxy.GetID() {
		log.Warnf("[actor] invoke dropped, ownership lost: uid=%d bound_nid=%s my_nid=%s", uid, nid, proxy.GetID())
		KillPlayer(proxy, uid)
		return
	}

	if act := GetPlayer(proxy, uid); act != nil {
		act.Invoke(fn)
		return
	}

	log.Warnf("[actor] player actor not found uid=%v", uid)
}

// InvokePlayerSync 向玩家 Actor 同步投递函数，在 Actor goroutine 中串行执行并返回结果。
//
// Actor 不存在时自动创建。
// ctx 用于取消等待：调用方取消 ctx 后不再等待 Actor 执行结果，直接返回 ctx.Err()。
// 归属权不属于本节点时返回错误。
func InvokePlayerSync[T any](ctx context.Context, proxy *node.Proxy, uid int64, fn func(context.Context) (T, error)) (T, error) {
	var zero T

	// 防御性检查：归属权
	if nid, err := proxy.LocateNode(context.Background(), uid, proxy.GetName()); err == nil && nid != "" && nid != proxy.GetID() {
		log.Warnf("[actor] invoke sync dropped, ownership lost: uid=%d bound_nid=%s my_nid=%s", uid, nid, proxy.GetID())
		KillPlayer(proxy, uid)
		return zero, fmt.Errorf("player actor ownership lost: uid=%d", uid)
	}

	act := GetPlayer(proxy, uid)
	if act == nil {
		return zero, fmt.Errorf("player actor not found: uid=%d", uid)
	}

	type result struct {
		val T
		err error
	}
	ch := make(chan result, 1)

	act.Invoke(func() {
		val, err := fn(ctx)
		ch <- result{val, err}
	})

	select {
	case r := <-ch:
		return r.val, r.err
	case <-ctx.Done():
		return zero, ctx.Err()
	}
}

// routeToPlayerActor 返回一个 Node 路由处理器，将消息投递到 Actor mailbox 串行处理。
//
// 归属权由 due 框架的 StatefulRoute 保证——gate 通过 Locator 定位玩家
// 绑定节点后才投递，消息不会到达旧节点。
//
// 配合 StatefulAuthorizedRoute 使用。使用 routeToPlayerActor 时，必须同时调用
// registerPlayerRouteInitializer 为 Actor 注册实际的路由处理器。
func routeToPlayerActor(kind string) node.RouteHandler {
	return func(ctx node.Context) {
		uid := ctx.UID()
		if uid == 0 {
			log.Warnf("[actor] uid == 0")
			return
		}

		// 归属权由 due 的 StatefulRoute 保证——gate 通过 Locator 定位
		// 玩家绑定节点后才投递消息，消息不会到达旧节点。
		proxy := ctx.Proxy()

		act := GetPlayer(proxy, uid)
		if act == nil {
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

// AddPlayerRouteHandler 给玩家Actor注册路由
func AddPlayerRouteHandler(proxy *node.Proxy, route int32, handler node.RouteHandler, opts ...node.RouteOptions) {
	proxy.AddRouteHandler(route, routeToPlayerActor(KindPlayer), opts...)
	registerPlayerRouteInitializer(func(act *node.Actor) {
		act.AddRouteHandler(route, handler)
	})
}
