package playerlife

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/module/actor"
)

// Manager 管理玩家断线 Grace Period、Actor 销毁和最终节点解绑。
type Manager struct {
	mu         sync.Mutex
	timers     map[int64]*time.Timer
	retries    map[int64]int
	proxy      *node.Proxy
	delay      time.Duration
	maxRetries int
}

// NewManager 创建玩家生命周期管理器。
//
// maxRetries 控制定位失败后的重试行为：
//   - -1 不重试，直接强制解绑
//   - 0  一直重试直到定位成功
//   - >0 最多重试 maxRetries 次
func NewManager(proxy *node.Proxy, delay time.Duration, maxRetries int) *Manager {
	return &Manager{
		timers:     make(map[int64]*time.Timer),
		retries:    make(map[int64]int),
		proxy:      proxy,
		delay:      delay,
		maxRetries: maxRetries,
	}
}

// onDisconnect 为断线玩家启动延迟解绑。
func (m *Manager) onDisconnect(ctx context.Context, uid int64) {
	m.mu.Lock()
	if t, ok := m.timers[uid]; ok {
		t.Stop()
	}
	delete(m.retries, uid)
	m.timers[uid] = time.AfterFunc(m.delay, func() {
		m.cleanup(context.Background(), uid)
	})
	m.mu.Unlock()
}

// OnLogin 取消待执行的清理，并在归属节点恢复玩家 Actor。
func (m *Manager) OnLogin(ctx context.Context, uid int64) {
	m.mu.Lock()
	if t, ok := m.timers[uid]; ok {
		t.Stop()
		delete(m.timers, uid)
		log.Infof("[playerlife] cleanup cancelled after relogin: uid=%d", uid)
	}
	delete(m.retries, uid)
	m.mu.Unlock()

	owns, err := m.ownsPlayer(ctx, uid)
	if err != nil {
		return
	}
	if owns {
		if _, err := actor.SpawnPlayer(m.proxy, uid); err != nil {
			log.Warnf("[playerlife] spawn player actor after login failed: uid=%d err=%v", uid, err)
		}
	}
}

func (m *Manager) cleanup(ctx context.Context, uid int64) {
	m.mu.Lock()
	retries := m.retries[uid]
	delete(m.timers, uid)
	m.mu.Unlock()

	gid, err := m.proxy.LocateGate(ctx, uid)
	if err != nil {
		m.scheduleRetry(uid, retries, fmt.Errorf("locate gate: %w", err))
		return
	}
	if gid != "" {
		log.Infof("[playerlife] cleanup skipped, player reconnected: uid=%d gid=%s", uid, gid)
		m.clearRetry(uid)
		return
	}

	owns, err := m.ownsPlayer(ctx, uid)
	if err != nil {
		m.scheduleRetry(uid, retries, fmt.Errorf("locate node: %w", err))
		return
	}
	if !owns {
		log.Infof("[playerlife] cleanup skipped, ownership lost: uid=%d", uid)
		m.clearRetry(uid)
		return
	}

	m.clearRetry(uid)
	log.Infof("[playerlife] grace period expired, unbinding node: uid=%d", uid)
	if err := m.proxy.UnbindNode(ctx, uid); err != nil {
		log.Warnf("[playerlife] unbind node failed: uid=%d err=%v", uid, err)
	}
}

func (m *Manager) scheduleRetry(uid int64, retries int, err error) {
	switch {
	case m.maxRetries < 0:
		m.clearRetry(uid)
		log.Warnf("[playerlife] delayed unbind check failed, force unbind (no retry): uid=%d err=%v", uid, err)
		_ = m.proxy.UnbindNode(context.Background(), uid)

	case m.maxRetries == 0 || retries < m.maxRetries:
		m.mu.Lock()
		m.retries[uid] = retries + 1
		m.timers[uid] = time.AfterFunc(m.delay, func() {
			m.cleanup(context.Background(), uid)
		})
		m.mu.Unlock()
		log.Warnf("[playerlife] delayed unbind check failed, scheduling retry: uid=%d attempt=%d maxRetries=%d err=%v",
			uid, retries+1, m.maxRetries, err)

	default:
		m.clearRetry(uid)
		log.Errorf("[playerlife] delayed unbind check failed after %d retries, force unbind: uid=%d err=%v", retries, uid, err)
		_ = m.proxy.UnbindNode(context.Background(), uid)
	}
}

func (m *Manager) clearRetry(uid int64) {
	m.mu.Lock()
	delete(m.retries, uid)
	m.mu.Unlock()
}

func (m *Manager) ownsPlayer(ctx context.Context, uid int64) (bool, error) {
	nid, err := m.proxy.LocateNode(ctx, uid, m.proxy.GetName())
	if err != nil {
		log.Warnf("[playerlife] locate node failed: uid=%d err=%v", uid, err)
		return false, err
	}
	return nid == m.proxy.GetID(), nil
}

// handleConnect 处理集群 Connect 事件。
func (m *Manager) handleConnect(ctx node.Context) {
	uid := ctx.UID()
	if uid == 0 {
		return
	}

	log.Infof("[playerlife] player connected: uid=%d cid=%d gid=%s", uid, ctx.CID(), ctx.GID())
	m.OnLogin(ctx.Context(), uid)
}

// handleDisconnect 处理集群 Disconnect 事件。
func (m *Manager) handleDisconnect(ctx node.Context) {
	uid := ctx.UID()
	if uid == 0 {
		log.Infof("[playerlife] player disconnected: uid=%d cid=%d", uid, ctx.CID())
		return
	}
	gid, err := m.proxy.LocateGate(ctx.Context(), uid)
	if err != nil {
		log.Warnf("[playerlife] disconnect gate check failed: uid=%d err=%v", uid, err)
		return
	}
	if gid != "" && gid != ctx.GID() {
		log.Debugf("[playerlife] stale disconnect ignored: uid=%d current_gid=%s event_gid=%s", uid, gid, ctx.GID())
		return
	}

	owns, err := m.ownsPlayer(ctx.Context(), uid)
	if err != nil {
		return
	}
	if !owns {
		log.Debugf("[playerlife] disconnect ignored, ownership lost: uid=%d", uid)
		return
	}

	actor.KillPlayer(m.proxy, uid)
	m.onDisconnect(ctx.Context(), uid)
	log.Infof("[playerlife] player disconnected: uid=%d cid=%d", uid, ctx.CID())
}
