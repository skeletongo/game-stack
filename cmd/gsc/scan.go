package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// EnumEntry proto 枚举条目（名称 + 值）。
type EnumEntry struct {
	Name  string // 如 "LOGIN"
	Value int32  // 如 1001
}

// RouteDef 路由定义。
type RouteDef struct {
	Name   string // 枚举值名，如 "LOGIN"
	Code   int32  // 路由号
	Module string // 所属模块名
}

// ErrorDef 错误码定义。
type ErrorDef struct {
	Name    string // 枚举值名，如 "INVALID_TOKEN"
	Code    int32  // 错误码
	Message string
	Module  string
}

// ModuleDef 模块定义。
type ModuleDef struct {
	Name       string
	Number     int
	RouteStart int32
}

// scanProtoEnumEntries 扫描 proto 枚举条目（含名称+值）。
func scanProtoEnumEntries() (map[string]map[string][]EnumEntry, error) {
	result := make(map[string]map[string][]EnumEntry)

	entries, err := filepath.Glob("proto/*/*.proto")
	if err != nil {
		return nil, err
	}

	enumValueRe := regexp.MustCompile(`^\s*(\w+)\s*=\s*(\d+)\s*;`)

	for _, path := range entries {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		content := string(data)

		enums := make(map[string][]EnumEntry)
		var currentEnum string
		var currentEntries []EnumEntry
		enumStartRe := regexp.MustCompile(`^enum\s+(\w+)\s*\{`)

		flush := func() {
			if currentEnum != "" && len(currentEntries) > 0 {
				enums[currentEnum] = currentEntries
			}
			currentEnum = ""
		}

		for _, line := range strings.Split(content, "\n") {
			if m := enumStartRe.FindStringSubmatch(line); m != nil {
				flush()
				currentEnum = m[1]
				currentEntries = nil
				continue
			}
			if currentEnum != "" && strings.TrimSpace(line) == "}" {
				flush()
				continue
			}
			if currentEnum != "" {
				if m := enumValueRe.FindStringSubmatch(line); m != nil {
					val, _ := strconv.Atoi(m[2])
					currentEntries = append(currentEntries, EnumEntry{Name: m[1], Value: int32(val)})
				}
			}
		}
		flush()

		if len(enums) > 0 {
			result[path] = enums
		}
	}

	return result, nil
}

// scanAllProtoEnums 扫描 proto 枚举（仅值，向后兼容）。
func scanAllProtoEnums() (map[string]map[string][]int32, error) {
	raw, err := scanProtoEnumEntries()
	if err != nil {
		return nil, err
	}
	result := make(map[string]map[string][]int32)
	for path, enums := range raw {
		m := make(map[string][]int32)
		for enumType, entries := range enums {
			var vals []int32
			for _, e := range entries {
				vals = append(vals, e.Value)
			}
			m[enumType] = vals
		}
		result[path] = m
	}
	return result, nil
}

// scanModules 扫描已有模块信息。
func scanModules() ([]ModuleDef, error) {
	entries, err := scanProtoEnumEntries()
	if err != nil {
		return nil, err
	}

	var modules []ModuleDef
	for path, enums := range entries {
		name := fileBase(path)
		if name == "common" {
			continue
		}

		mod := ModuleDef{Name: name}
		for enumType, vals := range enums {
			if !isRouteEnum(enumType) {
				continue
			}
			for _, e := range vals {
				if e.Value >= 1000 {
					mod.Number = int(e.Value / 1000)
					mod.RouteStart = int32(mod.Number * 1000)
					break
				}
			}
			if mod.Number > 0 {
				break
			}
		}
		modules = append(modules, mod)
	}

	sort.Slice(modules, func(i, j int) bool { return modules[i].Number < modules[j].Number })
	return modules, nil
}

// nextModuleNumber 返回下一个可用模块号。
func nextModuleNumber(modules []ModuleDef) int {
	used := make(map[int]bool)
	for _, m := range modules {
		if m.Number > 0 {
			used[m.Number] = true
		}
	}
	for n := 1; ; n++ {
		if !used[n] {
			return n
		}
	}
}

// collectAllRoutes 返回所有已占用的路由。
func collectAllRoutes() ([]RouteDef, error) {
	var routes []RouteDef

	entries, err := scanProtoEnumEntries()
	if err != nil {
		return nil, err
	}

	for path, enums := range entries {
		modName := fileBase(path)
		for enumType, vals := range enums {
			if !isRouteEnum(enumType) {
				continue
			}
			for _, e := range vals {
				if e.Value > 0 {
					routes = append(routes, RouteDef{Name: e.Name, Code: e.Value, Module: modName})
				}
			}
		}
	}

	sort.Slice(routes, func(i, j int) bool { return routes[i].Code < routes[j].Code })
	return routes, nil
}

// nextRouteCode 返回模块区间内下一个可用路由号。
func nextRouteCode(routes []RouteDef, start int32) (int32, error) {
	used := make(map[int32]bool)
	for _, r := range routes {
		used[r.Code] = true
	}
	for code := start + 1; code < start+1000; code++ {
		if !used[code] {
			return code, nil
		}
	}
	return 0, fmt.Errorf("模块路由号段 [%d-%d] 已满", start+1, start+999)
}

// routeCodeExists 检查路由号是否已被占用。
func routeCodeExists(routes []RouteDef, code int32) (string, bool) {
	for _, r := range routes {
		if r.Code == code {
			return r.Name, true
		}
	}
	return "", false
}

// collectAllErrors 返回所有已占用的错误码。
func collectAllErrors() ([]ErrorDef, error) {
	var errs []ErrorDef

	entries, err := scanProtoEnumEntries()
	if err != nil {
		return nil, err
	}

	for path, enums := range entries {
		modName := "system"
		if dir := filepath.Dir(path); dir != "." {
			modName = filepath.Base(dir)
		}
		for enumType, vals := range enums {
			if !isErrorEnum(enumType) {
				continue
			}
			for _, e := range vals {
				if e.Value > 0 {
					errs = append(errs, ErrorDef{Name: e.Name, Code: e.Value, Module: modName})
				}
			}
		}
	}

	sort.Slice(errs, func(i, j int) bool { return errs[i].Code < errs[j].Code })
	return errs, nil
}

// nextErrorCode 返回模块区间内下一个可用错误码。
func nextErrorCode(defs []ErrorDef, module string, modNum int) (int32, error) {
	used := make(map[int32]bool)
	for _, d := range defs {
		used[d.Code] = true
	}

	var start, end int32
	if module == "" {
		start, end = 0, 999
	} else {
		start = int32(modNum * 1000)
		end = start + 999
	}

	for code := start; code <= end; code++ {
		if !used[code] && code > 0 {
			return code, nil
		}
	}
	return 0, fmt.Errorf("错误码号段 [%d-%d] 已满", start+1, end)
}

// codeExists 检查错误码是否已被占用。
func codeExists(defs []ErrorDef, code int32) (string, bool) {
	for _, d := range defs {
		if d.Code == code {
			return d.Name, true
		}
	}
	return "", false
}
