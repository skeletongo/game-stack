package stack

import (
	"context"
	"sync"
	"time"

	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"
)

// PlayerDoneCleaner 管理玩家断线后的延迟清理。
//
// 设计原因：
//  1. Grace Period：玩家可能短暂断网后立即重连，延迟清理避免频繁创建/销毁
//  2. 防止消息队列竞态：延迟期间，旧节点的待处理消息可能仍在修改玩家数据。
//     真正清理时，节点检查确保只有"玩家已不在本节点"才执行清理。
type PlayerDoneCleaner struct {
	mu       sync.Mutex
	timers   map[int64]*time.Timer
	services []CleanableService
	proxy    *node.Proxy
	delay    time.Duration
}

// NewPlayerDoneCleaner 创建玩家延迟清理器。
// delay 是断线等待时间（建议 30~60 秒）。
func NewPlayerDoneCleaner(proxy *node.Proxy, delay time.Duration) *PlayerDoneCleaner {
	return &PlayerDoneCleaner{
		timers: make(map[int64]*time.Timer),
		proxy:  proxy,
		delay:  delay,
	}
}

// Register 注册一个 CleanableService。模块可在 Init 中调用。
func (c *PlayerDoneCleaner) Register(svc CleanableService) {
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
}

// doCleanup 执行真正的清理。先检查玩家是否还在本节点，再清理内存。
func (c *PlayerDoneCleaner) doCleanup(uid int64) {
	c.mu.Lock()
	delete(c.timers, uid)
	c.mu.Unlock()

	// 再查一次：玩家是否已经重新登录到本节点了（在定时器触发前的一瞬间）
	// 如果是，则不清理
	_, ok, err := c.proxy.AskNode(context.Background(), uid, c.proxy.GetName(), c.proxy.GetID())
	if err == nil && ok {
		log.Infof("[cleaner] cleanup skipped, player still on this node: uid=%d", uid)
		return
	}

	log.Infof("[cleaner] cleaning player data: uid=%d", uid)
	for _, svc := range c.services {
		svc.CleanPlayerData(uid)
	}
}
