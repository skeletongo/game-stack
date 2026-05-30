package main

import (
	"fmt"
	"os"
)

// routeCheck 检查路由号冲突和越界。
func routeCheck(_ []string) {
	routes, err := collectAllRoutes()
	if err != nil {
		fatalf("扫描路由失败: %v\n", err)
	}

	protoEnums, _ := scanAllProtoEnums()
	errors := 0

	// 检查重复
	seen := make(map[int32]string)
	for _, r := range routes {
		if prev, ok := seen[r.Code]; ok {
			fmt.Fprintf(os.Stderr, "ERROR: 重复路由号 %d: %s 和 %s\n", r.Code, prev, r.Name)
			errors++
		}
		seen[r.Code] = r.Name
	}

	// 检查越界（每个模块的 Route 枚举值应在对应千位段）
	for path, enums := range protoEnums {
		modName := fileBase(path)
		if modName == "common" {
			continue
		}
		for enumName, values := range enums {
			if !isRouteEnum(enumName) {
				continue
			}
			for _, v := range values {
				if v == 0 {
					continue
				}
				// 找该模块的起始值
				base := (v / 1000) * 1000
				if v < base || v > base+999 {
					fmt.Fprintf(os.Stderr, "ERROR: 路由号 %d 超出模块 %s 范围 (%d-%d)\n",
						v, modName, base, base+999)
					errors++
				}
			}
		}
	}

	if errors == 0 {
		fmt.Println("路由检查通过，未发现问题。")
	} else {
		fmt.Printf("\n%d 个错误\n", errors)
		os.Exit(1)
	}
}
