package app

import (
	"github.com/dobyte/due/component/http/v2"
	"github.com/dobyte/due/v2/cluster"
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/skeletongo/game-stack/cmd/frontend/app/context"
	"github.com/skeletongo/game-stack/cmd/frontend/app/handler/auth"
	"github.com/skeletongo/game-stack/cmd/frontend/app/logic"
)

// Init 初始化 frontend HTTP 和节点代理。
func Init(proxyHttp *http.Proxy, proxyNode *node.Proxy) {
	context.NewServiceContext(proxyHttp, proxyNode)

	// 初始化节点生命周期回调。
	initProxy(proxyNode)

	// 初始化 HTTP API 路由。
	initAPI()
}

// initProxy 注册节点生命周期钩子。
func initProxy(proxy *node.Proxy) {
	proxy.AddHookListener(cluster.Init, logic.OnServerInit)
	proxy.AddHookListener(cluster.Start, logic.OnServerStart)
	proxy.AddHookListener(cluster.Close, logic.OnServerClose)
	proxy.AddHookListener(cluster.Destroy, logic.OnServerDestroy)
}

// initAPI 注册 frontend 对外 HTTP 接口。
func initAPI() {
	auth.New().Init()
}
