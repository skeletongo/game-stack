package clean

import (
	"context"
	"sync"
	"time"

	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"
	"github.com/skeletongo/game-stack/module/actor"
)

// CleanablePlayer 是模块 Service 可选实现的接口。
// 实现了此接口的模块，会在玩家断线时被调用以清理该玩家的内存数据。
//
// CleanPlayerData 会重试直到成功（最多 maxRetries 次），
// 全部成功后才会解除节点绑定，防止清理失败导致数据丢失。
type CleanablePlayer interface {
	CleanPlayerData(uid int64) error
}

// PlayerDoneCleaner 管理玩家断线后的延迟清理。
//
// 设计原因：
//  1. Grace Period：玩家可能短暂断网后立即重连，延迟清理避免频繁创建/销毁。
//  2. 断线瞬间杀 Actor：关闭 mailbox，丢弃排队消息，杜绝旧消息在 Grace Period
//     期间修改玩家数据。
//  3. 重连检测：doCleanup 时通过 LocateGate 检查玩家是否有活跃的 Gate 连接，
//     有则跳过清理（已重连），无则执行清理。
//  4. 失败重试：CleanPlayerData 失败时会延迟重试，全部成功后才会 UnbindNode，
//     防止清理失败导致数据丢失。
type PlayerDoneCleaner struct {
	mu         sync.Mutex
	timers     map[int64]*time.Timer
	retries    map[int64]int // per-UID retry count
	services   []CleanablePlayer
	proxy      *node.Proxy
	delay      time.Duration
	maxRetries int // -1: 不重试, 0: 一直重试, >0: 最大重试次数
}

// NewPlayerDoneCleaner 创建玩家延迟清理器。
//
// delay 是断线等待时间（建议 30~60 秒）。
// maxRetries 控制清理失败后的重试行为：
//   - -1  不重试，清理一次后直接解绑
//   - 0   一直重试直到成功
//   - >0  最多重试 maxRetries 次
func NewPlayerDoneCleaner(proxy *node.Proxy, delay time.Duration, maxRetries int) *PlayerDoneCleaner {
	return &PlayerDoneCleaner{
		timers:     make(map[int64]*time.Timer),
		retries:    make(map[int64]int),
		proxy:      proxy,
		delay:      delay,
		maxRetries: maxRetries,
	}
}

// Register 注册一个 CleanablePlayer。模块可在 Init 中调用。
func (c *PlayerDoneCleaner) Register(svc CleanablePlayer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services = append(c.services, svc)
}

// OnDisconnect 玩家断线时调用，启动延迟清理定时器。
func (c *PlayerDoneCleaner) OnDisconnect(uid int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果已有定时器，说明之前断线但还没清理，直接重置
	if t, ok := c.timers[uid]; ok {
		t.Stop()
	}

	// 重置重试计数
	delete(c.retries, uid)

	c.timers[uid] = time.AfterFunc(c.delay, func() {
		c.doCleanup(uid)
	})

	log.Debugf("[cleaner] disconnect cleanup scheduled: uid=%d delay=%v", uid, c.delay)
}

// OnLogin 玩家重新登录时调用，取消延迟清理。
func (c *PlayerDoneCleaner) OnLogin(uid int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if t, ok := c.timers[uid]; ok {
		t.Stop()
		delete(c.timers, uid)
		log.Infof("[cleaner] cleanup cancelled after relogin: uid=%d", uid)
	}

	delete(c.retries, uid)
}

// doCleanup 执行真正的清理。先检查玩家是否已重连，再清理内存。
// 如果 CleanPlayerData 失败则根据 maxRetries 决定是否重试。
func (c *PlayerDoneCleaner) doCleanup(uid int64) {
	c.mu.Lock()
	retries := c.retries[uid]
	delete(c.timers, uid)
	c.mu.Unlock()

	// 再查一次：玩家是否已经重连（有活跃的 Gate 绑定）
	gid, err := c.proxy.LocateGate(context.Background(), uid)
	if err == nil && gid != "" {
		log.Infof("[cleaner] cleanup skipped, player reconnected: uid=%d gid=%s", uid, gid)
		c.mu.Lock()
		delete(c.retries, uid)
		c.mu.Unlock()
		return
	}

	log.Infof("[cleaner] cleaning player data: uid=%d attempt=%d", uid, retries)

	var errs []error
	for _, svc := range c.services {
		if err := svc.CleanPlayerData(uid); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		// 全部成功 → 解绑
		c.mu.Lock()
		delete(c.retries, uid)
		c.mu.Unlock()

		log.Infof("[cleaner] cleanup done, unbinding: uid=%d", uid)
		_ = c.proxy.UnbindNode(context.Background(), uid)
		return
	}

	// 有失败，根据 maxRetries 决定行为
	switch {
	case c.maxRetries < 0:
		// 不重试，直接解绑
		c.mu.Lock()
		delete(c.retries, uid)
		c.mu.Unlock()

		log.Warnf("[cleaner] cleanup failed, force unbind (no retry): uid=%d errs=%v", uid, errs)
		_ = c.proxy.UnbindNode(context.Background(), uid)

	case c.maxRetries == 0 || retries < c.maxRetries:
		// 无限重试 or 未达上限 → 延迟重试
		c.mu.Lock()
		c.retries[uid] = retries + 1
		c.timers[uid] = time.AfterFunc(c.delay, func() {
			c.doCleanup(uid)
		})
		c.mu.Unlock()

		log.Warnf("[cleaner] cleanup failed, scheduling retry: uid=%d attempt=%d maxRetries=%d errs=%v",
			uid, retries+1, c.maxRetries, errs)

	default:
		// 已达重试上限 → 强制解绑
		c.mu.Lock()
		delete(c.retries, uid)
		c.mu.Unlock()

		log.Errorf("[cleaner] cleanup failed after %d retries, force unbind: uid=%d errs=%v", retries, uid, errs)
		_ = c.proxy.UnbindNode(context.Background(), uid)
	}
}

// HandleConnect 处理集群 Connect 事件。
//
// Connect 事件为全集群广播。旧节点收到后取消玩家的延迟清理定时器，
// 若 Actor 已被杀死则重新创建（断线重连场景）。
func (c *PlayerDoneCleaner) HandleConnect(ctx node.Context) {
	uid := ctx.UID()
	if uid == 0 {
		return
	}
	log.Infof("player connected: uid=%d cid=%d gid=%s", uid, ctx.CID(), ctx.GID())
	c.OnLogin(uid)
}

// HandleDisconnect 处理集群 Disconnect 事件。
//
// 两阶段清理：
//
//	立即：杀 Actor（关闭 mailbox，丢弃排队消息）
//	延迟 30s：Grace Period 到期后清理 token + 解绑节点
func (c *PlayerDoneCleaner) HandleDisconnect(ctx node.Context) {
	uid := ctx.UID()
	if uid != 0 {
		actor.KillPlayer(c.proxy, uid)
		c.OnDisconnect(uid)
	}
	log.Infof("player disconnected: uid=%d cid=%d", uid, ctx.CID())
}
