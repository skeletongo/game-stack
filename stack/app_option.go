package stack

import (
	"time"

	"github.com/dobyte/due/v2/crypto"
	"github.com/dobyte/due/v2/encoding"
	"github.com/dobyte/due/v2/locate"
	"github.com/dobyte/due/v2/registry"
	"github.com/dobyte/due/v2/transport"
)

// appOptions 应用配置选项。
type appOptions struct {
	id          string
	name        string
	locator     locate.Locator
	registry    registry.Registry
	encryptor   crypto.Encryptor
	transporter transport.Transporter
	codec       encoding.Codec
	addr        string
	connNum     int
	callTimeout time.Duration
	dialTimeout time.Duration
	weight      int
	metadata    map[string]string
	modules     []Module
	debugAddr   string // 不为空时启动 debug HTTP 服务
}

func defaultAppOptions() *appOptions {
	return &appOptions{
		connNum:     5,
		callTimeout: 3 * time.Second,
		dialTimeout: 3 * time.Second,
		weight:      100,
	}
}

// AppOption 是 Application 的函数式配置选项。
type AppOption func(o *appOptions)

// WithID 设置实例 ID。
func WithID(id string) AppOption {
	return func(o *appOptions) { o.id = id }
}

// WithName 设置实例名称。
func WithName(name string) AppOption {
	return func(o *appOptions) { o.name = name }
}

// WithLocator 设置用户定位器（如 Redis）。
func WithLocator(l locate.Locator) AppOption {
	return func(o *appOptions) { o.locator = l }
}

// WithRegistry 设置服务注册器（如 etcd）。
func WithRegistry(r registry.Registry) AppOption {
	return func(o *appOptions) { o.registry = r }
}

// WithEncryptor 设置加密器。
func WithEncryptor(e crypto.Encryptor) AppOption {
	return func(o *appOptions) { o.encryptor = e }
}

// WithTransporter 设置 RPC 传输器（如 gRPC）。
func WithTransporter(t transport.Transporter) AppOption {
	return func(o *appOptions) { o.transporter = t }
}

// WithCodec 设置消息编解码器。
func WithCodec(c encoding.Codec) AppOption {
	return func(o *appOptions) { o.codec = c }
}

// WithAddr 设置内部通信监听地址。
func WithAddr(addr string) AppOption {
	return func(o *appOptions) { o.addr = addr }
}

// WithConnNum 设置最大连接数。
func WithConnNum(n int) AppOption {
	return func(o *appOptions) { o.connNum = n }
}

// WithCallTimeout 设置 RPC 调用超时时间。
func WithCallTimeout(d time.Duration) AppOption {
	return func(o *appOptions) { o.callTimeout = d }
}

// WithDialTimeout 设置拨号超时时间。
func WithDialTimeout(d time.Duration) AppOption {
	return func(o *appOptions) { o.dialTimeout = d }
}

// WithWeight 设置负载均衡权重。
func WithWeight(w int) AppOption {
	return func(o *appOptions) { o.weight = w }
}

// WithMetadata 设置元数据。
func WithMetadata(m map[string]string) AppOption {
	return func(o *appOptions) { o.metadata = m }
}

// WithModules 设置要加载的游戏模块列表。
func WithModules(modules ...Module) AppOption {
	return func(o *appOptions) { o.modules = modules }
}

// WithDebug 启用 debug HTTP 服务（开发调试用）。
// addr 如 "127.0.0.1:6060"。不设置则不启动。
func WithDebug(addr string) AppOption {
	return func(o *appOptions) { o.debugAddr = addr }
}
