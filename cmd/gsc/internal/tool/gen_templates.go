package tool

import (
	"bytes"
	"strings"
	"text/template"
)

// 代码生成模板

// ModuleGoTmpl 模块 module.go 模板。
var ModuleGoTmpl = template.Must(template.New("module.go").Parse(`package {{.Name}}

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/stack"
)

const name = "{{.Name}}"

// Module 创建{{.Name}}模块。
func Module(opts ...Option) stack.Module {
	return &{{.Name}}Module{opts: opts}
}

type {{.Name}}Module struct {
	opts []Option
}

func (m *{{.Name}}Module) Name() string { return name }

func (m *{{.Name}}Module) Init(proxy *node.Proxy) error {
	o := defaultOptions()
	for _, opt := range m.opts {
		opt(o)
	}

	impl := newImpl(o.store)

	// TODO: 在此注册路由处理器
	// proxy.AddRouteHandler(stack.Route{{.NameTitle}}Xxx, impl.handleXxx, stack.StatefulAuthorizedRoute)

	stack.RegisterService(name, impl.svc)

	log.Infof("[{{.Name}}] module initialized")
	return nil
}
`))

// ImplGoTmpl 模块 impl.go 模板。
var ImplGoTmpl = template.Must(template.New("impl.go").Parse(`package {{.Name}}

import (
	"github.com/dobyte/due/v2/cluster/node"

	"github.com/skeletongo/game-stack/stack"
)

type impl struct {
	svc *service
}

func newImpl(store Store) *impl {
	return &impl{svc: newService(store)}
}

// TODO: 在此添加路由处理函数
`))

// StoreGoTmpl 模块 store.go 模板。
var StoreGoTmpl = template.Must(template.New("store.go").Parse(`package {{.Name}}

import "context"

// Store 定义{{.NameTitle}}模块的数据存储接口。
// 默认使用内存实现，生产环境可注入 Redis/MySQL 实现。
type Store interface {
	// TODO: 定义存储方法
	_ context.Context // 占位，确保 import 不被 gofmt 移除
}
`))

// StoreMemoryGoTmpl 模块 store_memory.go 模板。
var StoreMemoryGoTmpl = template.Must(template.New("store_memory.go").Parse(`package {{.Name}}

import (
	"sync"
)

type memoryStore struct {
	mu sync.RWMutex
	// TODO: 添加数据字段
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		// TODO: 初始化数据字段
	}
}
`))

// OptionGoTmpl 模块 option.go 模板。
var OptionGoTmpl = template.Must(template.New("option.go").Parse(`package {{.Name}}

type options struct {
	store Store
}

func defaultOptions() *options {
	return &options{store: newMemoryStore()}
}

// Option 模块配置选项。
type Option func(o *options)

// WithStore 注入自定义存储实现。
func WithStore(s Store) Option {
	return func(o *options) { o.store = s }
}
`))

// ServiceGoTmpl 模块 service.go 模板。
var ServiceGoTmpl = template.Must(template.New("service.go").Parse(`package {{.Name}}

// Service 定义{{.NameTitle}}模块对外暴露的服务接口。
type Service interface {
	// TODO: 定义服务方法
}

type service struct {
	store Store
}

func newService(store Store) *service {
	return &service{store: store}
}
`))

// MsgGoTmpl protocol 消息文件模板。
var MsgGoTmpl = template.Must(template.New("message.go").Parse(`// Package {{.Name}} 定义{{.NameTitle}}相关消息类型。
package {{.Name}}

// TODO: 添加 Request/Response 结构体
`))

// HandlerStubTmpl 路由处理函数桩模板。
var HandlerStubTmpl = template.Must(template.New("handler").Parse(
	strings.TrimSpace(`
func (i *impl) handle{{.Action}}(ctx node.Context) {
	req := &{{.PkgName}}.{{.Action}}Request{}
	if err := ctx.Parse(req); err != nil {
		stack.ProtoResponse(ctx, &{{.PkgName}}.{{.Action}}Response{Code: stack.ErrCode(err), Message: err.Error()})
		return
	}
	// TODO: 实现 {{.Action}} 业务逻辑
	stack.ProtoResponse(ctx, &{{.PkgName}}.{{.Action}}Response{Code: stack.CodeOK})
}
`) + "\n",
))

// HandlerStubActorTmpl Actor 模式处理函数桩模板。
var HandlerStubActorTmpl = template.Must(template.New("handler_actor").Parse(
	strings.TrimSpace(`
func (i *impl) handle{{.Action}}Actor(ctx node.Context) {
	req := &{{.PkgName}}.{{.Action}}Request{}
	if err := ctx.Parse(req); err != nil {
		stack.ProtoResponse(ctx, &{{.PkgName}}.{{.Action}}Response{Code: stack.ErrCode(err), Message: err.Error()})
		return
	}
	// TODO: 实现 {{.Action}} Actor 业务逻辑
	stack.ProtoResponse(ctx, &{{.PkgName}}.{{.Action}}Response{Code: stack.CodeOK})
}
`) + "\n",
))

// ProtoStructTmpl 请求/响应结构体模板。
var ProtoStructTmpl = template.Must(template.New("proto").Parse(
	strings.TrimSpace(`
// {{.Action}}Request {{.Action}}请求。
type {{.Action}}Request struct {
	// TODO: 添加字段
}

// {{.Action}}Response {{.Action}}响应。
type {{.Action}}Response struct {
	// TODO: 添加字段
}
`) + "\n",
))

// HandlerStubData 生成 handler 桩的模板数据。
type HandlerStubData struct {
	Action  string // PascalCase 操作名
	PkgName string // protocol 包别名（如 "pauth"）
}

// ProtoStructData 生成消息结构体的模板数据。
type ProtoStructData struct {
	Action string // PascalCase 操作名
}

// RenderTemplate 执行模板并返回结果字符串。
func RenderTemplate(tmpl *template.Template, data any) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
