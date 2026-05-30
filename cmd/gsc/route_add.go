package main

import (
	"fmt"
	"os"
	"strings"
)

// routeAdd 为指定模块添加路由常量。
// 用法: gsc route add -m <模块名> <操作名> [路由号]
func routeAdd(args []string) {
	var modName, action string
	var codeNum int32

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-m":
			i++
			if i >= len(args) {
				fatalf("-m 需要一个模块名\n")
			}
			modName = args[i]
		default:
			if action == "" {
				action = args[i]
			} else if codeNum == 0 {
				fmt.Sscanf(args[i], "%d", &codeNum)
			}
		}
	}

	if modName == "" || action == "" {
		fmt.Fprintln(os.Stderr, "用法: gsc route add -m <模块名> <操作名> [路由号]")
		fmt.Fprintln(os.Stderr, "示例: gsc route add -m auth Login")
		fmt.Fprintln(os.Stderr, "      gsc route add -m mail Send 10001")
		os.Exit(1)
	}

	// 获取模块号
	modules, _ := scanModules()
	var modNum int
	for _, m := range modules {
		if m.Name == modName {
			modNum = m.Number
			break
		}
	}
	if modNum == 0 {
		fatalf("未找到模块: %s（请先执行 gsc module new %s）\n", modName, modName)
	}

	// 获取已占用的路由号
	routes, err := collectAllRoutes()
	if err != nil {
		fatalf("扫描路由失败: %v\n", err)
	}

	if codeNum == 0 {
		// 自动分配
		codeNum, err = nextRouteCode(routes, int32(modNum*1000))
		if err != nil {
			fatalf("%v\n", err)
		}
	} else {
		// 手动指定，检查冲突
		if name, ok := routeCodeExists(routes, codeNum); ok {
			fatalf("路由号 %d 已被 %s 使用\n", codeNum, name)
		}
		// 检查范围
		base := int32(modNum * 1000)
		if codeNum < base || codeNum > base+999 {
			fatalf("路由号 %d 超出模块 %s 范围 (%d-%d)\n", codeNum, modName, base, base+999)
		}
	}

	// proto 枚举名
	protoName := toSnakeCase(action)
	routeConstName := fmt.Sprintf("Route%s%s", strings.ToUpper(modName[:1])+modName[1:], action)

	if name, ok := routeCodeExists(routes, codeNum); ok && name != protoName {
		fatalf("路由号 %d 已被 %s 使用\n", codeNum, name)
	}

	fmt.Printf("添加路由: %s = %d (%s)\n", routeConstName, codeNum, protoName)

	// 1. 写入 proto 枚举
	protoPath := fmt.Sprintf("proto/%s/%s.proto", modName, modName)
	routeEnum := fmt.Sprintf("%sRoute", strings.ToUpper(modName[:1])+modName[1:])
	entry := fmt.Sprintf("\t%-30s = %d;", protoName, codeNum)
	if err := insertIntoProto(protoPath, routeEnum, entry); err != nil {
		fatalf("写入 proto 失败: %v\n", err)
	}
	fmt.Printf("  更新 %s\n", protoPath)

	// 2. 写入 stack/route.go
	routePath := "stack/route.go"
	data, _ := os.ReadFile(routePath)
	content := string(data)
	// 在最后的 ) 之前插入
	lastParen := strings.LastIndex(content, ")")
	constLine := fmt.Sprintf("\t%-40s int32 = int32(%s.%s_%s)", routeConstName, modName, routeEnum, protoName)
	newContent := content[:lastParen] + constLine + "\n" + content[lastParen:]
	if err := os.WriteFile(routePath, []byte(newContent), 0644); err != nil {
		fatalf("写入 route.go 失败: %v\n", err)
	}
	fmt.Printf("  更新 %s\n", routePath)

	// 3. 注册模块路由 handler
	fmt.Println("完成。请在 module/<name>/module.go 中添加 AddRouteHandler 注册。")
}

