package tool

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strconv"
	"strings"
)

// 模块名（小写）→ 路由常量前缀中的首字母大写形式。
// 新模块默认使用 strings.Title 规则，此处仅处理缩写等特殊情况。
var moduleTitleForm = map[string]string{
	"inventory": "Inv",
}

// TitleName 返回模块名在路由常量中的首字母大写形式（处理缩写等特殊情况）。
func TitleName(name string) string {
	if t, ok := moduleTitleForm[name]; ok {
		return t
	}
	if len(name) == 0 {
		return name
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

// 段头部注释正则：// <Name> 模块 (<range>)
// 例如: "// Auth 模块 (1000-1999)"
var sectionHeaderRe = regexp.MustCompile(`^(\S+)\s*模块\s*\((\d+)-(\d+)\)\s*$`)

// ParseRouteFile 解析 stack/route.go，返回所有模块段及其路由。
func ParseRouteFile(path string) ([]ModuleSection, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("解析 %s 失败: %w", path, err)
	}

	// 找到 const() 声明块
	var constDecl *ast.GenDecl
	for _, decl := range f.Decls {
		if gd, ok := decl.(*ast.GenDecl); ok && gd.Tok == token.CONST {
			constDecl = gd
			break
		}
	}
	if constDecl == nil {
		return nil, fmt.Errorf("%s 中未找到 const() 声明块", path)
	}

	specs := constDecl.Specs
	// 先遍历一遍，为每个 Spec 收集其前导 Doc 注释内容
	type specInfo struct {
		spec      *ast.ValueSpec
		docText   string // Doc 注释文本（非空表示新模块段头部）
		specIndex int
	}
	var infos []specInfo
	for i, spec := range specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		var docText string
		if vs.Doc != nil && len(vs.Doc.List) > 0 {
			docText = vs.Doc.Text()
		}
		infos = append(infos, specInfo{spec: vs, docText: docText, specIndex: i})
	}

	// 按模块段分组
	var sections []ModuleSection
	var curSection *ModuleSection

	for _, info := range infos {
		vs := info.spec

		// 检查是否是模块段头部注释
		if info.docText != "" {
			matches := sectionHeaderRe.FindStringSubmatch(info.docText)
			if len(matches) == 4 {
				// 结束当前段，开始新段
				if curSection != nil {
					curSection.EndLine = lastRouteEndLine(curSection)
					sections = append(sections, *curSection)
				}
				baseNum, _ := strconv.Atoi(matches[2])
				curSection = &ModuleSection{
					ModuleName:   strings.ToLower(matches[1]),
					ModuleNumber: baseNum / 1000,
					BaseNumber:   int32(baseNum),
					HeaderLine:   fset.Position(vs.Pos()).Line,
				}
			}
		}

		if curSection == nil {
			continue // 第一个模块段头部之前的注释/声明跳过
		}

		// 提取路由常量
		if len(vs.Names) == 1 && len(vs.Values) == 1 {
			name := vs.Names[0].Name
			if !strings.HasPrefix(name, "Route") {
				continue
			}
			val, err := parseConstValue(vs.Values[0])
			if err != nil {
				continue // 非常量值跳过
			}

			route := RouteInfo{
				Name:       name,
				Number:     int32(val),
				ModuleName: curSection.ModuleName,
				Action:     extractAction(name, curSection.ModuleName),
				Line:       fset.Position(vs.Pos()).Line,
			}
			curSection.Routes = append(curSection.Routes, route)
		}
	}

	// 结束最后一个段
	if curSection != nil {
		curSection.EndLine = lastRouteEndLine(curSection)
		sections = append(sections, *curSection)
	}

	return sections, nil
}

// parseConstValue 解析 AST 常量值表达式为整数。
func parseConstValue(expr ast.Expr) (int, error) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind == token.INT {
			return strconv.Atoi(e.Value)
		}
	}
	return 0, fmt.Errorf("非常量表达式")
}

// extractAction 从路由常量名中提取操作名。
// 例如: "RouteAuthLogin" + module "auth" → "Login"
func extractAction(routeName, moduleName string) string {
	prefix := "Route" + TitleName(moduleName)
	if strings.HasPrefix(routeName, prefix) {
		return routeName[len(prefix):]
	}
	return ""
}

// lastRouteEndLine 返回模块段中最后一个路由常量所在的行号。
func lastRouteEndLine(sec *ModuleSection) int {
	if len(sec.Routes) > 0 {
		return sec.Routes[len(sec.Routes)-1].Line
	}
	return sec.HeaderLine
}
