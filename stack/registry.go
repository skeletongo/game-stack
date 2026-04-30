package stack

import (
	"fmt"
	"sync"
)

var serviceRegistry = struct {
	mu       sync.RWMutex
	services map[string]any
}{services: make(map[string]any)}

// RegisterService 注册模块服务。
// 模块在 Init 方法中调用此函数注册其 Service 接口，
// 供其他模块通过 GetService 获取。
func RegisterService(name string, svc any) {
	serviceRegistry.mu.Lock()
	defer serviceRegistry.mu.Unlock()

	if _, ok := serviceRegistry.services[name]; ok {
		panic(fmt.Sprintf("service %s already registered", name))
	}
	serviceRegistry.services[name] = svc
}

// GetService 获取已注册的模块服务。
// 使用方需要自行做类型断言。
func GetService(name string) any {
	serviceRegistry.mu.RLock()
	defer serviceRegistry.mu.RUnlock()
	return serviceRegistry.services[name]
}
