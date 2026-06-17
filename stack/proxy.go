package stack

import (
	"github.com/dobyte/due/v2/cluster"
	"github.com/dobyte/due/v2/cluster/node"
)

type myproxy struct {
	proxy  *node.Proxy
	events map[cluster.Event][]node.EventHandler
}

var global = &myproxy{
	proxy:  nil,
	events: make(map[cluster.Event][]node.EventHandler),
}

// Proxy 获取节点代理
func Proxy() *node.Proxy {
	return global.proxy
}

// AddEventHandler 添加事件处理器
func AddEventHandler(event cluster.Event, handler node.EventHandler) {
	global.events[event] = append(global.events[event], handler)
}

func Set(proxy *node.Proxy) {
	global.proxy = proxy
}

func Init() {
	if global.proxy == nil {
		panic("stack proxy not set!!!")
	}
	for k, v := range global.events {
		global.proxy.AddEventHandler(k, func(ctx node.Context) {
			for _, v := range v {
				v(ctx)
			}
		})
	}
}
