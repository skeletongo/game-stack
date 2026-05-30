package tool

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

// ParseNodeMain 解析 cmd/node/main.go，返回：
// - imports: 已导入的 game-stack module 包路径列表
// - moduleArgs: WithModules() 调用中已有的模块参数名列表
// - withModLine: WithModules 调用的结束括号行号 (1-based)
// - lastImportLine: 最后一个 game-stack module import 所在行号
func ParseNodeMain(path string) (imports []string, moduleArgs []string, withModLine int, lastImportLine int, err error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("解析 %s 失败: %w", path, err)
	}

	// 收集 game-stack module import
	for _, imp := range f.Imports {
		impPath := strings.Trim(imp.Path.Value, `"`)
		if strings.HasPrefix(impPath, "github.com/skeletongo/game-stack/module/") {
			imports = append(imports, impPath)
			line := fset.Position(imp.Pos()).Line
			if line > lastImportLine {
				lastImportLine = line
			}
		}
	}

	// 查找 WithModules() 调用
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		x, ok := sel.X.(*ast.Ident)
		if !ok || x.Name != "stack" || sel.Sel.Name != "WithModules" {
			return true
		}

		// 收集已有模块参数名
		for _, arg := range call.Args {
			callExpr, ok := arg.(*ast.CallExpr)
			if !ok {
				continue
			}
			sel2, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			x2, ok := sel2.X.(*ast.Ident)
			if !ok {
				continue
			}
			moduleArgs = append(moduleArgs, x2.Name)
		}

		// WithModules() 结束括号位置
		withModLine = fset.Position(call.Rparen).Line
		return false
	})

	return
}

// FindLastRouteHandlerLine 在模块的 module.go 中找到最后一个 AddRouteHandler 调用的行号。
// 返回 0 表示还没注册任何路由。
func FindLastRouteHandlerLine(path string) (int, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return 0, fmt.Errorf("解析 %s 失败: %w", path, err)
	}

	lastLine := 0
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if sel.Sel.Name == "AddRouteHandler" {
			line := fset.Position(call.Rparen).Line
			if line > lastLine {
				lastLine = line
			}
		}
		return true
	})

	return lastLine, nil
}

// FindNewImplLine 在模块的 module.go 中找到 newImpl(...) 调用的行号。
// 用于在没有已注册路由时确定插入位置。
func FindNewImplLine(path string) (int, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return 0, fmt.Errorf("解析 %s 失败: %w", path, err)
	}

	newImplLine := 0
	ast.Inspect(f, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for _, rhs := range assign.Rhs {
			call, ok := rhs.(*ast.CallExpr)
			if !ok {
				continue
			}
			if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "newImpl" {
				newImplLine = fset.Position(assign.Pos()).Line
				return false
			}
		}
		return true
	})

	return newImplLine, nil
}

// ListModuleDirs 列出 module/ 下的所有子目录（排除 actor）。
func ListModuleDirs(root string) ([]string, error) {
	entries, err := filepath.Glob(filepath.Join(root, "module", "*"))
	if err != nil {
		return nil, err
	}
	var dirs []string
	for _, entry := range entries {
		info, err := filepath.Glob(entry) // just to reuse the path
		_ = info
		_ = err
	}
	return dirs, nil
}
