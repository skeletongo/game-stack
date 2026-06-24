package logic

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"
)

// OnServerInit 处理节点初始化事件。
func OnServerInit(proxy *node.Proxy) {
	log.Infof("服务初始化")
}

// OnServerStart 处理节点启动事件。
func OnServerStart(proxy *node.Proxy) {
	log.Infof("服务启动")
}

// OnServerClose 处理节点关闭事件。
func OnServerClose(proxy *node.Proxy) {
	log.Infof("服务关闭")
}

// OnServerDestroy 处理节点销毁事件。
func OnServerDestroy(proxy *node.Proxy) {
	log.Infof("服务销毁")
}
