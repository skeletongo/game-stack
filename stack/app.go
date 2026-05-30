package stack

import (
	"github.com/dobyte/due/v2"
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/stack/debug"
)

// Application 是 game-stack 框架的应用入口。
// 封装 due Container，提供简化的启动引导和模块管理。
type Application struct {
	opts *appOptions
}

// NewApplication 创建一个 game-stack 应用。
//
// 用法示例：
//
//	app := stack.NewApplication(
//	    stack.WithName("game-server"),
//	    stack.WithLocator(redis.NewLocator()),
//	    stack.WithRegistry(etcd.NewRegistry()),
//	    stack.WithTransporter(grpc.NewTransporter()),
//	    stack.WithModules(
//	        auth.Module(auth.WithSecretKey("my-secret")),
//	        player.Module(),
//	    ),
//	)
//	app.Run()
func NewApplication(opts ...AppOption) *Application {
	o := defaultAppOptions()
	for _, opt := range opts {
		opt(o)
	}
	return &Application{opts: o}
}

// Run 启动应用。创建 due Container、初始化模块、注册组件并开始服务。
// 此方法会阻塞直到收到系统信号（SIGINT/SIGTERM），然后执行优雅关闭。
func (a *Application) Run() {
	container := due.NewContainer()

	n := a.buildNode()

	proxy := n.Proxy()

	for _, m := range a.opts.modules {
		log.Infof("initializing module: %s", m.Name())
		if err := m.Init(proxy); err != nil {
			log.Fatalf("module %s init failed: %v", m.Name(), err)
		}
	}

	// 启动 debug HTTP 服务（可选）
	if a.opts.debugAddr != "" {
		debug.NewServer(a.opts.debugAddr).StartAsync()
	}

	container.Add(n)
	container.Serve()
}

// buildNode 根据选项构建 due Node 组件。
func (a *Application) buildNode() *node.Node {
	var opts []node.Option

	if a.opts.name != "" {
		opts = append(opts, node.WithName(a.opts.name))
	}
	if a.opts.id != "" {
		opts = append(opts, node.WithID(a.opts.id))
	}
	if a.opts.locator != nil {
		opts = append(opts, node.WithLocator(a.opts.locator))
	}
	if a.opts.registry != nil {
		opts = append(opts, node.WithRegistry(a.opts.registry))
	}
	if a.opts.encryptor != nil {
		opts = append(opts, node.WithEncryptor(a.opts.encryptor))
	}
	if a.opts.transporter != nil {
		opts = append(opts, node.WithTransporter(a.opts.transporter))
	}
	if a.opts.codec != nil {
		opts = append(opts, node.WithCodec(a.opts.codec))
	}
	if a.opts.addr != "" {
		opts = append(opts, node.WithAddr(a.opts.addr))
	}
	if a.opts.connNum > 0 {
		opts = append(opts, node.WithConnNum(a.opts.connNum))
	}
	if a.opts.callTimeout > 0 {
		opts = append(opts, node.WithCallTimeout(a.opts.callTimeout))
	}
	if a.opts.dialTimeout > 0 {
		opts = append(opts, node.WithDialTimeout(a.opts.dialTimeout))
	}
	if a.opts.weight > 0 {
		opts = append(opts, node.WithWeight(a.opts.weight))
	}
	if len(a.opts.metadata) > 0 {
		opts = append(opts, node.WithMetadata(a.opts.metadata))
	}

	return node.NewNode(opts...)
}
