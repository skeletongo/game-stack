package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// validModuleName 模块名仅允许小写字母+数字，以字母开头。
var validModuleName = regexp.MustCompile(`^[a-z][a-z0-9]*$`)

// moduleNew 创建完整模块骨架。
func moduleNew(args []string) {
	var name string
	var modNum int

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-n":
			i++
			if i >= len(args) {
				fatalf("-n 需要一个模块号\n")
			}
			fmt.Sscanf(args[i], "%d", &modNum)
		default:
			name = args[i]
		}
	}

	if name == "" {
		fmt.Fprintln(os.Stderr, "用法: gsc module new <模块名> [-n 模块号]")
		os.Exit(1)
	}
	if !validModuleName.MatchString(name) {
		fatalf("无效的模块名: %s（小写字母+数字，以字母开头）\n", name)
	}

	if modNum == 0 {
		modules, _ := scanModules()
		modNum = nextModuleNumber(modules)
	}

	routeStart := int32(modNum * 1000)
	nameTitle := strings.ToUpper(name[:1]) + name[1:]

	if _, err := os.Stat(fmt.Sprintf("module/%s", name)); err == nil {
		fatalf("模块 %s 已存在（module/%s）\n", name, name)
	}
	if _, err := os.Stat(fmt.Sprintf("proto/%s", name)); err == nil {
		fatalf("模块 %s 已存在（proto/%s）\n", name, name)
	}

	fmt.Printf("创建模块: %s (模块号 %d, 区间 %d-%d)\n", nameTitle, modNum, routeStart, routeStart+999)

	createProtoFile(name, nameTitle, routeStart)
	createModuleFiles(name, nameTitle)
	addRouteSection(name, nameTitle, routeStart)
	addErrCodeSection(name, nameTitle, modNum)
	addModuleToMain(name)

	fmt.Println("完成。")
	fmt.Printf("  添加路由: gsc route add -m %s <操作名>\n", name)
	fmt.Printf("  添加错误码: gsc errcode add -m %s <名称> <描述>\n", name)
}

// createProtoFile 创建 proto/<name>/<name>.proto，含 Route/Error 枚举和消息占位。
func createProtoFile(name, nameTitle string, routeStart int32) {
	dir := filepath.Join("proto", name)
	os.MkdirAll(dir, 0755)

	upper := strings.ToUpper(name)
	content := fmt.Sprintf(`syntax = "proto3";

package %[1]s;

option go_package = "github.com/skeletongo/game-stack/proto/%[1]s";

// %[2]sRoute %[2]s模块路由编号（%[3]d-%[4]d）。
enum %[2]sRoute {
    %[5]s_ROUTE_UNSPECIFIED = 0;
    // TODO: 添加路由
}

// %[2]sError %[2]s模块错误码（%[3]d-%[4]d）。
enum %[2]sError {
    %[5]s_ERROR_UNSPECIFIED = 0;
    // TODO: 添加错误码
}

// ==== 消息 ====
// TODO: 添加客户端消息
`, name, nameTitle, routeStart, routeStart+999, upper)

	path := filepath.Join(dir, name+".proto")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fatalf("创建 %s 失败: %v\n", path, err)
	}
	fmt.Printf("  创建 %s\n", path)
}

// createModuleFiles 创建 module/<name>/ 下 6 文件模板。
func createModuleFiles(name, nameTitle string) {
	dir := filepath.Join("module", name)
	os.MkdirAll(dir, 0755)

	templates := map[string]string{
		"store.go":        storeGo(name),
		"store_memory.go": storeMemoryGo(name),
		"option.go":       optionGo(name),
		"service.go":      serviceGo(name, nameTitle),
		"impl.go":         implGo(name),
		"module.go":       moduleGo(name, nameTitle),
	}

	for fn, content := range templates {
		path := filepath.Join(dir, fn)
		os.WriteFile(path, []byte(content), 0644)
		fmt.Printf("  创建 %s\n", path)
	}
}

// addRouteSection 在 stack/route.go 末尾追加模块路由段注释。
func addRouteSection(name, nameTitle string, routeStart int32) {
	data, _ := os.ReadFile("stack/route.go")
	content := string(data)
	last := strings.LastIndex(content, ")")
	sec := fmt.Sprintf("\n\t// %s 模块 (%d-%d)\n\t// TODO: add route constants for %s\n",
		nameTitle, routeStart, routeStart+999, name)
	os.WriteFile("stack/route.go", []byte(content[:last]+sec+content[last:]), 0644)
	fmt.Println("  更新 stack/route.go")
}

// addErrCodeSection 在 stack/errcode.go 末尾追加模块错误码注释段。
func addErrCodeSection(name, nameTitle string, modNum int) {
	data, _ := os.ReadFile("stack/errcode.go")
	content := string(data)
	last := strings.LastIndex(content, ")")
	sec := fmt.Sprintf("\n// %s 模块错误 (%d-%d)\nvar (\n\t// TODO: add error codes for %s\n)\n",
		nameTitle, modNum*1000, modNum*1000+999, name)
	os.WriteFile("stack/errcode.go", []byte(content[:last+1]+"\n"+sec), 0644)
	fmt.Println("  更新 stack/errcode.go")
}

// addModuleToMain 在 cmd/node/main.go 中添加模块 import 和 WithModules 条目。
func addModuleToMain(name string) {
	data, _ := os.ReadFile("cmd/node/main.go")
	content := string(data)

	ip := fmt.Sprintf(`"github.com/skeletongo/game-stack/module/%s"`, name)
	if !strings.Contains(content, ip) {
		marker := `"github.com/skeletongo/game-stack/stack"`
		content = strings.Replace(content, marker, marker+"\n\t"+ip, 1)
	}

	idx := strings.Index(content, "stack.WithModules(")
	close := strings.Index(content[idx:], "\n\t)") + idx
	line := fmt.Sprintf("\t\t%s.Module(),", name)
	content = content[:close] + "\n" + line + content[close:]

	os.WriteFile("cmd/node/main.go", []byte(content), 0644)
	fmt.Println("  更新 cmd/node/main.go")
}

// ==== 6 文件模板 ====

func storeGo(name string) string {
	return fmt.Sprintf(`package %[1]s

import "context"

// Store 定义%[1]s模块的数据存储接口。
type Store interface {
	// TODO: 添加数据存取方法
}
`, name)
}

func storeMemoryGo(name string) string {
	return fmt.Sprintf(`package %[1]s

import (
	"context"
	"sync"
)

type memoryStore struct {
	mu sync.RWMutex
	// TODO: 添加内存存储字段
}

func newMemoryStore() *memoryStore {
	return &memoryStore{}
}
`, name)
}

func optionGo(name string) string {
	return fmt.Sprintf(`package %[1]s

type options struct {
	store Store
}

func defaultOptions() *options {
	return &options{store: newMemoryStore()}
}

type Option func(o *options)

func WithStore(s Store) Option {
	return func(o *options) { o.store = s }
}
`, name)
}

func serviceGo(name, nameTitle string) string {
	return fmt.Sprintf(`package %[1]s

import "context"

// Service %[2]s模块对外的服务接口。
type Service interface {
	// TODO: 添加服务方法
}

type service struct {
	store Store
}

func newService(store Store) *service {
	return &service{store: store}
}

// CleanPlayerData 实现 stack.CleanableService，断线时清理玩家内存数据。
func (s *service) CleanPlayerData(uid int64) error {
	// TODO: 实现清理逻辑
	return nil
}
`, name, nameTitle)
}

func implGo(name string) string {
	return fmt.Sprintf(`package %[1]s

import (
	"github.com/dobyte/due/v2/cluster/node"

	proto "github.com/skeletongo/game-stack/proto/%[1]s"
	"github.com/skeletongo/game-stack/stack"
)

type impl struct {
	svc *service
}

func newImpl(store Store) *impl {
	return &impl{svc: newService(store)}
}
`, name)
}

func moduleGo(name, nameTitle string) string {
	return fmt.Sprintf(`package %[1]s

import (
	"github.com/dobyte/due/v2/cluster/node"
	"github.com/dobyte/due/v2/log"

	"github.com/skeletongo/game-stack/stack"
)

const name = "%[1]s"

// Module 创建%[2]s模块。
func Module(opts ...Option) stack.Module {
	return &%[1]sModule{opts: opts}
}

type %[1]sModule struct {
	opts []Option
}

func (m *%[1]sModule) Name() string { return name }

func (m *%[1]sModule) Init(proxy *node.Proxy) error {
	o := defaultOptions()
	for _, opt := range m.opts {
		opt(o)
	}

	impl := newImpl(o.store)

	// TODO: 注册路由

	// 注册玩家数据清理方法
	if c, ok := stack.GetService("cleaner").(*stack.PlayerDoneCleaner); ok {
		c.Register(impl.svc)
	}

	// 注册内部服务接口，供其它模块使用
	stack.RegisterService(name, impl.svc)

	log.Infof("[%%s] module initialized", name)
	return nil
}
`, name, nameTitle)
}
