package main

import (
	"github.com/dobyte/due/component/http/v2"
	"github.com/dobyte/due/locate/redis/v2"
	"github.com/dobyte/due/registry/etcd/v2"
	"github.com/dobyte/due/transport/grpc/v2"
	"github.com/dobyte/due/v2"
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/config"
	"github.com/dobyte/due/v2/config/file"

	"github.com/skeletongo/game-stack/cmd/frontend/app"
)

func init() {
	// 设置全局配置器
	config.SetConfigurator(config.NewConfigurator(config.WithSources(file.NewSource())))
}

// @title API服文档
// @version 1.0
// @host localhost:8080
// @BasePath /
// main 启动 frontend HTTP API 服务。
func main() {
	// 创建容器
	container := due.NewContainer()
	// 创建服务注册发现
	registry := etcd.NewRegistry()
	// 创建RPC传输器
	transporter := grpc.NewTransporter()
	// 创建定位器
	locator := redis.NewLocator()

	// 创建HTTP组件
	compHttp := http.NewServer(
		http.WithRegistry(registry),
		http.WithTransporter(transporter),
	)

	// 创建节点组件
	compNode := node.NewNode(
		node.WithLocator(locator),
		node.WithRegistry(registry),
		node.WithTransporter(transporter),
	)

	// 初始化应用
	app.Init(compHttp.Proxy(), compNode.Proxy())
	// 添加NODE组件
	container.Add(compHttp, compNode)
	// 启动容器
	container.Serve()
}
