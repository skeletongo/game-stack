// Package debug 提供开发调试用的 HTTP 服务。
//
// 端点：
//   - GET  /debug/modules           列出所有模块
//   - GET  /debug/module/:name       列出模块能力
//   - POST /debug/query              查询聚合快照
//   - POST /debug/command            执行命令
//   - POST /debug/patch              直接修改聚合字段
//
// 启用：stack.WithDebug(":6060")
// 模块注册：debug.Register[*Player]("player", repo, cmdBus)
package debug

import (
	"context"
	"sync"

	"github.com/skeletongo/game-stack/ddd"
)

// Module 存储模块的 debug 能力。
type Module struct {
	Name   string
	Load   func(ctx context.Context, id int64) (ddd.Aggregate, error)
	Save   func(ctx context.Context, agg ddd.Aggregate) error
	CmdBus *ddd.CommandBus
}

var registry = struct {
	mu      sync.RWMutex
	modules map[string]*Module
}{modules: make(map[string]*Module)}

// Register 注册模块到 debug 服务。
//
// T 由 repo 参数自动推导。Register 后模块的所有命令和聚合查询自动可用。
func Register[T ddd.Aggregate](name string, repo ddd.Repository[T], cmdBus *ddd.CommandBus) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	registry.modules[name] = &Module{
		Name: name,
		Load: func(ctx context.Context, id int64) (ddd.Aggregate, error) {
			return repo.Load(ctx, id)
		},
		Save: func(ctx context.Context, agg ddd.Aggregate) error {
			return repo.Save(ctx, agg.(T))
		},
		CmdBus: cmdBus,
	}
}

// get 按名称获取模块。
func get(name string) (*Module, bool) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	m, ok := registry.modules[name]
	return m, ok
}

// names 返回所有模块名称。
func names() []string {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	ns := make([]string, 0, len(registry.modules))
	for name := range registry.modules {
		ns = append(ns, name)
	}
	return ns
}
