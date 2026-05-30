package tool

import (
	"fmt"
	"os"
	"sort"
)

// Registry 是 gsc 核心数据结构，聚合所有模块和路由信息。
// 每次运行时从源码文件动态解析构建。
type Registry struct {
	root     string                  // 项目根目录
	Sections []ModuleSection         // route.go 中的所有模块段（按出现顺序）
	Modules  []*ModuleInfo           // 所有模块（合并 route.go + 磁盘目录）
	ByName   map[string]*ModuleInfo  // 模块名 → 模块信息
	ByNumber map[int]*ModuleInfo     // 模块号 → 模块信息
	ByRoute  map[int32]*RouteInfo    // 路由号 → 路由信息
	AllRoutes []*RouteInfo           // 所有路由（按源文件顺序）
}

// BuildRegistry 从项目根目录构建注册表。
func BuildRegistry(root string) (*Registry, error) {
	r := &Registry{
		root:    root,
		ByName:  make(map[string]*ModuleInfo),
		ByNumber: make(map[int]*ModuleInfo),
		ByRoute:  make(map[int32]*RouteInfo),
	}

	// 1. 解析 route.go
	sections, err := ParseRouteFile(RouteGo(root))
	if err != nil {
		return nil, fmt.Errorf("解析路由文件失败: %w", err)
	}
	r.Sections = sections

	// 2. 构建初始模块信息（来自 route.go）
	for _, sec := range sections {
		info := &ModuleInfo{
			Name:   sec.ModuleName,
			Number: sec.ModuleNumber,
			Routes: make([]RouteInfo, len(sec.Routes)),
		}
		copy(info.Routes, sec.Routes)
		r.Modules = append(r.Modules, info)
		r.ByName[info.Name] = info
		r.ByNumber[info.Number] = info

		for i := range sec.Routes {
			route := &sec.Routes[i]
			r.AllRoutes = append(r.AllRoutes, route)
			r.ByRoute[route.Number] = route
		}
	}

	// 3. 检查 module/ 目录，发现磁盘上的模块（可能未在 route.go 中注册）
	modDir := ModuleDir(root, "")
	entries, err := os.ReadDir(modDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			// 跳过非模块目录（如 actor）
			if name == "actor" {
				continue
			}
			if _, ok := r.ByName[name]; !ok {
				// 磁盘上有但 route.go 中无 → 添加占位模块信息
				info := &ModuleInfo{
					Name:     name,
					Number:   0, // 未分配模块号
					DirExist: true,
				}
				r.Modules = append(r.Modules, info)
				r.ByName[name] = info
			} else {
				r.ByName[name].DirExist = true
			}
		}
	}

	// 4. 解析错误码文件，关联到模块
	errRanges, err := ParseErrCodeFile(ErrCodeGo(root))
	if err == nil {
		for i := range errRanges {
			er := &errRanges[i]
			if info, ok := r.ByName[er.Module]; ok {
				info.ErrRange = er
			}
		}
	}

	// 5. 按模块号排序
	sort.Slice(r.Modules, func(i, j int) bool {
		return r.Modules[i].Number < r.Modules[j].Number
	})

	return r, nil
}

// NextSubProtocol 返回指定模块的下一个可用子协议号（从 1 开始，最大 999）。
func (r *Registry) NextSubProtocol(name string) int32 {
	info := r.ByName[name]
	if info == nil {
		return 1
	}
	max := int32(0)
	for _, route := range info.Routes {
		sub := route.Number - int32(info.Number*1000)
		if sub > max {
			max = sub
		}
	}
	return max + 1
}

// NextModuleNumber 返回下一个可用模块号。
func (r *Registry) NextModuleNumber() int {
	max := 0
	for _, info := range r.Modules {
		if info.Number > max {
			max = info.Number
		}
	}
	return max + 1
}

// IsRouteNameExists 检查路由常量名是否已存在。
func (r *Registry) IsRouteNameExists(name string) bool {
	for _, route := range r.AllRoutes {
		if route.Name == name {
			return true
		}
	}
	return false
}

// IsRouteNumberExists 检查路由号是否已存在。
func (r *Registry) IsRouteNumberExists(num int32) bool {
	_, ok := r.ByRoute[num]
	return ok
}
