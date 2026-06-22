// Package actor 封装 due Actor 模型，为玩家状态管理提供串行化执行保障。
//
// 核心机制：
//  1. 玩家 Actor 跟随玩家节点归属，不跟随普通连接生命周期。
//     登录流程不主动创建 Actor；有状态路由或 InvokePlayer* 在缺失时按需创建。
//  2. Actor 访问会刷新空闲计时；玩家离线且空闲超时后只释放本地 Actor。
//     后续消息仍会投递到玩家绑定节点，并在 Actor 缺失时重新创建。
//  3. InvokePlayer* 会校验玩家节点归属；发现本地 Actor 失去归属时丢弃回调并清理残留 Actor。
//  4. 模块通过 AddPlayerRouteHandler 注册 Actor 上的路由处理器。
//     Spawn 时自动应用所有已注册的初始化器。
//
// 三种使用模式：
//
//	模式1 (fire-and-forget):
//	  actor.InvokePlayer(ctx, proxy, uid, func(ctx context.Context) {
//	      _ = playerSvc.DeductGold(ctx, uid, 100)
//	  })
//
//	模式2 (同步执行并返回结果):
//	  newExp, err := actor.InvokePlayerSync[int64](ctx, proxy, uid, func(ctx context.Context) (int64, error) {
//	      return ddd.Dispatch[int64](ctx, cmdBus, application.AddExpCmd{...})
//	  })
//
//	模式3 (同步请求-响应，通过 Actor mailbox):
//	  模块 Init 时：
//	    actor.AddPlayerRouteHandler(proxy, route, handler, opts)
//	  消息处理：
//	    Node router → routeToPlayerActor → Actor mailbox → Actor.handler → ctx.Response()
package actor

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/errors"
	"github.com/dobyte/due/v2/log"
)

const (
	KindPlayer    = "player"
	playerIdleTTL = 30 * time.Minute
)

type playerRouteInitializer func(actor *node.Actor)

var (
	mu                sync.Mutex
	routeInitializers []playerRouteInitializer

	idleMu     sync.Mutex
	idleTimers = make(map[int64]*time.Timer)
)

func registerPlayerRouteInitializer(fn playerRouteInitializer) {
	mu.Lock()
	defer mu.Unlock()
	routeInitializers = append(routeInitializers, fn)
}

// SpawnPlayer 创建或返回当前节点上的玩家 Actor。
// 登录流程不会主动调用该方法；有状态路由在玩家 Actor 缺失时按需创建。
func SpawnPlayer(proxy *node.Proxy, uid int64) (*node.Actor, error) {
	id := strconv.FormatInt(uid, 10)

	act, err := proxy.Spawn(
		func(actor *node.Actor, args ...any) node.Processor {
			proc := &playerProc{actor: actor, uid: uid}

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
			idleMu.Lock()
			act, ok := proxy.Actor(KindPlayer, id)
			if !ok {
				idleMu.Unlock()
				log.Errorf("[actor] spawn player actor find failed: uid=%d", uid)
				return nil, err
			}
			touchPlayerIdleLocked(proxy, uid)
			idleMu.Unlock()
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

	touchPlayerIdle(proxy, uid)

	log.Infof("[actor] spawned: uid=%d pid=%s routes=%d", uid, act.PID(), len(routeInitializers))
	return act, nil
}

// KillPlayer 只释放当前节点上的玩家 Actor，不解绑玩家节点归属。
func KillPlayer(proxy *node.Proxy, uid int64) {
	idleMu.Lock()
	stopPlayerIdleLocked(uid)
	killed := killPlayerLocked(proxy, uid)
	idleMu.Unlock()

	if killed {
		log.Infof("[actor] killed: uid=%d", uid)
	}
}

func touchPlayerIdle(proxy *node.Proxy, uid int64) {
	idleMu.Lock()
	touchPlayerIdleLocked(proxy, uid)
	idleMu.Unlock()
}

func touchPlayerIdleLocked(proxy *node.Proxy, uid int64) {
	var timer *time.Timer
	timer = time.AfterFunc(playerIdleTTL, func() {
		reapIdlePlayer(context.Background(), proxy, uid, timer)
	})

	if old := idleTimers[uid]; old != nil {
		old.Stop()
	}
	idleTimers[uid] = timer
}

// stopPlayerIdle 停止玩家 Actor 的空闲回收计时。
func stopPlayerIdle(uid int64) {
	idleMu.Lock()
	stopPlayerIdleLocked(uid)
	idleMu.Unlock()
}

func stopPlayerIdleLocked(uid int64) {
	if timer := idleTimers[uid]; timer != nil {
		timer.Stop()
		delete(idleTimers, uid)
	}
}

func killPlayerLocked(proxy *node.Proxy, uid int64) bool {
	id := strconv.FormatInt(uid, 10)
	proxy.UnbindActor(uid, KindPlayer)
	return proxy.Kill(KindPlayer, id)
}

func checkPlayerOwnership(ctx context.Context, proxy *node.Proxy, uid int64) (bool, error) {
	nid, err := proxy.LocateNode(ctx, uid, proxy.GetName())
	if err != nil {
		if errors.Is(err, errors.ErrNotFoundUserLocation) {
			return true, fmt.Errorf("player node binding not found: uid=%d", uid)
		}
		return false, fmt.Errorf("locate player node failed: uid=%d: %w", uid, err)
	}
	if nid == "" {
		return true, fmt.Errorf("player node binding not found: uid=%d", uid)
	}
	if nid != proxy.GetID() {
		return true, fmt.Errorf("player actor ownership lost: uid=%d bound_nid=%s my_nid=%s", uid, nid, proxy.GetID())
	}
	return false, nil
}

// reapIdlePlayer 检查玩家是否仍在线；离线且本地 Actor 仍有效时释放 Actor。
func reapIdlePlayer(ctx context.Context, proxy *node.Proxy, uid int64, timer *time.Timer) {
	idleMu.Lock()
	if idleTimers[uid] != timer {
		idleMu.Unlock()
		return
	}
	idleMu.Unlock()

	gid, err := proxy.LocateGate(ctx, uid)
	if err != nil && !errors.Is(err, errors.ErrNotFoundUserLocation) {
		log.Warnf("[actor] idle check locate gate failed: uid=%d err=%v", uid, err)
		touchPlayerIdle(proxy, uid)
		return
	}
	if gid != "" {
		touchPlayerIdle(proxy, uid)
		return
	}

	nid, err := proxy.LocateNode(ctx, uid, proxy.GetName())
	if err != nil && !errors.Is(err, errors.ErrNotFoundUserLocation) {
		log.Warnf("[actor] idle check locate node failed: uid=%d err=%v", uid, err)
		touchPlayerIdle(proxy, uid)
		return
	}

	idleMu.Lock()
	if idleTimers[uid] != timer {
		idleMu.Unlock()
		return
	}
	delete(idleTimers, uid)
	killed := killPlayerLocked(proxy, uid)
	idleMu.Unlock()
	if !killed {
		return
	}

	if nid != "" && nid != proxy.GetID() {
		log.Infof("[actor] idle actor ownership lost, killing local actor: uid=%d bound_nid=%s my_nid=%s", uid, nid, proxy.GetID())
		return
	}

	log.Infof("[actor] idle timeout, killing local actor: uid=%d", uid)
}

// GetPlayer 获取当前节点上的玩家 Actor；不存在时按需创建。
func GetPlayer(proxy *node.Proxy, uid int64) *node.Actor {
	idleMu.Lock()
	act, ok := proxy.Actor(KindPlayer, strconv.FormatInt(uid, 10))
	if ok {
		touchPlayerIdleLocked(proxy, uid)
		idleMu.Unlock()
		return act
	}
	idleMu.Unlock()

	var err error
	act, err = SpawnPlayer(proxy, uid)
	if err != nil {
		return nil
	}
	return act
}

// InvokePlayer 将函数投递到玩家 Actor 中串行执行，不等待执行结果。
func InvokePlayer(ctx context.Context, proxy *node.Proxy, uid int64, fn func(context.Context)) {
	if cleanup, err := checkPlayerOwnership(ctx, proxy, uid); err != nil {
		log.Warnf("[actor] invoke dropped: %v", err)
		if cleanup {
			KillPlayer(proxy, uid)
		}
		return
	}

	if act := GetPlayer(proxy, uid); act != nil {
		act.Invoke(func() {
			if cleanup, err := checkPlayerOwnership(ctx, proxy, uid); err != nil {
				log.Warnf("[actor] invoke dropped: %v", err)
				if cleanup {
					go KillPlayer(proxy, uid)
				}
				return
			}
			fn(ctx)
		})
		return
	}

	log.Warnf("[actor] player actor not found uid=%v", uid)
}

// InvokePlayerSync 将函数投递到玩家 Actor 中串行执行，并等待执行结果。
func InvokePlayerSync[T any](ctx context.Context, proxy *node.Proxy, uid int64, fn func(context.Context) (T, error)) (T, error) {
	var zero T

	if cleanup, err := checkPlayerOwnership(ctx, proxy, uid); err != nil {
		log.Warnf("[actor] invoke sync dropped: %v", err)
		if cleanup {
			KillPlayer(proxy, uid)
		}
		return zero, err
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
		if cleanup, err := checkPlayerOwnership(ctx, proxy, uid); err != nil {
			log.Warnf("[actor] invoke sync dropped: %v", err)
			ch <- result{err: err}
			if cleanup {
				go KillPlayer(proxy, uid)
			}
			return
		}
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

func routeToPlayerActor(kind string) node.RouteHandler {
	return func(ctx node.Context) {
		uid := ctx.UID()
		if uid == 0 {
			log.Warnf("[actor] uid == 0")
			return
		}

		act := GetPlayer(ctx.Proxy(), uid)
		if act == nil {
			log.Warnf("[actor] actor %s not found uid: %d", kind, uid)
			return
		}
		act.Next(ctx)
	}
}

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
