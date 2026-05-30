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

// 错误码块头部注释正则：// <Name> 模块错误 (<range>)
var errRangeHeaderRe = regexp.MustCompile(`^(\S+)\s*模块错误\s*\((\d+)-(\d+)\)\s*$`)

// ParseErrCodeFile 解析 stack/errcode.go，返回所有模块的错误码块。
func ParseErrCodeFile(path string) ([]ErrRange, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("解析 %s 失败: %w", path, err)
	}

	var ranges []ErrRange
	var curRange *ErrRange

	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.VAR {
			continue
		}

		// 检查 Doc 注释是否是模块错误码头部
		docText := ""
		if gd.Doc != nil {
			docText = gd.Doc.Text()
		}

		matches := errRangeHeaderRe.FindStringSubmatch(docText)
		if len(matches) == 4 {
			if curRange != nil {
				curRange.EndLine = fset.Position(gd.Pos()).Line - 1
				ranges = append(ranges, *curRange)
			}
			base, _ := strconv.Atoi(matches[2])
			curRange = &ErrRange{
				Module: strings.ToLower(matches[1]),
				Base:   int32(base),
			}
		}

		if curRange == nil {
			continue // 系统级错误码块（无模块头部注释）跳过
		}

		// 提取错误码
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok || len(vs.Names) == 0 {
				continue
			}

			name := vs.Names[0].Name
			if !strings.HasPrefix(name, "Err") {
				continue
			}

			codeVal, msg := extractCodeLiteral(vs)
			if codeVal < 0 {
				continue
			}

			curRange.Errs = append(curRange.Errs, ErrInfo{
				Name:    name,
				Code:    codeVal,
				Message: msg,
			})
		}

		// 记录 var() 块的结束行
		curRange.EndLine = fset.Position(gd.End()).Line
	}

	if curRange != nil {
		ranges = append(ranges, *curRange)
	}

	return ranges, nil
}

// extractCodeLiteral 从 ValueSpec 中提取 &Code{Code: N, Message: "..."} 的字面量。
func extractCodeLiteral(vs *ast.ValueSpec) (int32, string) {
	if len(vs.Values) != 1 {
		return -1, ""
	}

	comp, ok := vs.Values[0].(*ast.CompositeLit)
	if !ok {
		return -1, ""
	}

	var code int32 = -1
	var msg string

	for _, elt := range comp.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		switch key.Name {
		case "Code":
			if lit, ok := kv.Value.(*ast.BasicLit); ok && lit.Kind == token.INT {
				v, err := strconv.Atoi(lit.Value)
				if err == nil {
					code = int32(v)
				}
			}
		case "Message":
			if lit, ok := kv.Value.(*ast.BasicLit); ok && lit.Kind == token.STRING {
				msg = strings.Trim(lit.Value, `"`)
			}
		}
	}

	return code, msg
}

// NextErrBase 根据已解析的错误码块，返回下一个模块可用的错误码基数。
func NextErrBase(ranges []ErrRange) int32 {
	maxBase := int32(0)
	for _, r := range ranges {
		if r.Base > maxBase {
			maxBase = r.Base
		}
	}
	if maxBase == 0 {
		return 1000
	}
	return maxBase + 100
}
