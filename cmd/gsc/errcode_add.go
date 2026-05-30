package main

import (
	"fmt"
	"os"
	"strings"
)

// errcodeAdd 添加错误码。
// 用法:
//
//	gsc errcode add <名称> <描述>               系统错误 → proto/common/common.proto
//	gsc errcode add -m <模块名> <名称> <描述>   模块错误 → proto/<name>/<name>.proto
func errcodeAdd(args []string) {
	var modName, name, msg string
	var codeNum int32

	// 解析参数：[-m 模块] 名称 描述 [错误码]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-m":
			i++
			if i >= len(args) {
				fatalf("-m 需要一个模块名\n")
			}
			modName = args[i]
		default:
			if name == "" {
				name = args[i]
			} else if msg == "" {
				msg = args[i]
			} else if codeNum == 0 {
				fmt.Sscanf(args[i], "%d", &codeNum)
			}
		}
	}

	if name == "" || msg == "" {
		fmt.Fprintln(os.Stderr, "用法:")
		fmt.Fprintln(os.Stderr, "  gsc errcode add <名称> <描述>              系统错误")
		fmt.Fprintln(os.Stderr, "  gsc errcode add -m <模块名> <名称> <描述>  模块错误")
		fmt.Fprintln(os.Stderr, "示例:")
		fmt.Fprintln(os.Stderr, "  gsc errcode add ERR_RATE_LIMIT \"rate limit exceeded\"")
		fmt.Fprintln(os.Stderr, "  gsc errcode add -m mail MAIL_FULL \"mailbox full\"")
		os.Exit(1)
	}

	// 收集已有错误码
	errDefs, err := collectAllErrors()
	if err != nil {
		fatalf("扫描错误码失败: %v\n", err)
	}

	// 确定 proto 文件和枚举名
	var protoPath, enumName, goSuffix string
	var modNum int

	if modName != "" {
		// 模块错误
		protoPath = fmt.Sprintf("proto/%s/%s.proto", modName, modName)
		if _, err := os.Stat(protoPath); os.IsNotExist(err) {
			fatalf("proto 文件不存在: %s（请先执行 gsc module new %s）\n", protoPath, modName)
		}
		enumName = fmt.Sprintf("%sError", strings.ToUpper(modName[:1])+modName[1:])

		modules, _ := scanModules()
		for _, m := range modules {
			if m.Name == modName {
				modNum = m.Number
				break
			}
		}

		goSuffix = fmt.Sprintf("%s.%s_%s", modName, enumName, name)
	} else {
		// 系统错误
		protoPath = "proto/common/common.proto"
		enumName = "SysError"
		goSuffix = fmt.Sprintf("common.SysError_%s", name)
	}

	// 分配或验证错误码
	if codeNum == 0 {
		codeNum, err = nextErrorCode(errDefs, modName, modNum)
		if err != nil {
			fatalf("%v\n", err)
		}
	} else {
		if existing, ok := codeExists(errDefs, codeNum); ok {
			fatalf("错误码 %d 已被 %s 使用\n", codeNum, existing)
		}
	}

	// 检查枚举名是否已存在
	for _, e := range errDefs {
		if e.Name == name {
			fatalf("枚举名 %s 已存在（错误码 %d）\n", name, e.Code)
		}
	}

	fmt.Printf("添加错误码: %s = %d (\"%s\")\n", name, codeNum, msg)

	// 1. 写入 proto 枚举
	entry := fmt.Sprintf("\t%-30s = %d;", name, codeNum)
	if err := insertIntoProto(protoPath, enumName, entry); err != nil {
		fatalf("写入 proto 失败: %v\n", err)
	}
	fmt.Printf("  更新 %s\n", protoPath)

	// 2. 写入 stack/errcode.go
	errPath := "stack/errcode.go"
	data, _ := os.ReadFile(errPath)
	content := string(data)
	lastParen := strings.LastIndex(content, ")")
	goLine := fmt.Sprintf("\t%-30s = &Code{Code: int32(%s), Message: \"%s\"}",
		"Err"+toGoName(name), goSuffix, msg)
	// 找模块错误段（如果有 modName）
	if modName != "" {
		marker := fmt.Sprintf("// %s 模块错误", strings.ToUpper(modName[:1])+modName[1:])
		idx := strings.Index(content, marker)
		if idx >= 0 {
			// 插入到该段之后
			insertAfter := idx
			for insertAfter < len(content) && !strings.HasPrefix(content[insertAfter:], "var (") {
				insertAfter++
			}
			// 找到 var ( 块结束的 )
			varEnd := strings.Index(content[insertAfter:], ")") + insertAfter
			newContent := content[:varEnd] + "\t" + goLine + "\n" + content[varEnd:]
			if err := os.WriteFile(errPath, []byte(newContent), 0644); err != nil {
				fatalf("写入 errcode.go 失败: %v\n", err)
			}
			fmt.Printf("  更新 %s\n", errPath)
			return
		}
	}
	// 系统错误，在 ) 之前插入
	newContent := content[:lastParen] + goLine + "\n" + content[lastParen:]
	if err := os.WriteFile(errPath, []byte(newContent), 0644); err != nil {
		fatalf("写入 errcode.go 失败: %v\n", err)
	}
	fmt.Printf("  更新 %s\n", errPath)
}

// toGoName 将 proto 枚举名转为 Go 导出变量名。
// 例: INVALID_TOKEN → InvalidToken, MAIL_FULL → MailFull
func toGoName(protoEnum string) string {
	parts := strings.Split(strings.ToLower(protoEnum), "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}

// errcodeList 列出所有错误码。
func errcodeList(_ []string) {
	errs, err := collectAllErrors()
	if err != nil {
		fatalf("扫描错误码失败: %v\n", err)
	}

	byModule := make(map[string][]ErrorDef)
	for _, e := range errs {
		byModule[e.Module] = append(byModule[e.Module], e)
	}

	for mod, list := range byModule {
		fmt.Printf("=== %s (%d 个) ===\n", mod, len(list))
		for _, e := range list {
			fmt.Printf("  %-35s %d  \"%s\"\n", e.Name, e.Code, e.Message)
		}
		fmt.Println()
	}
}
