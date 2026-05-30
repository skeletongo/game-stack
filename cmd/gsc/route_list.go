package main

import "fmt"

// routeList 列出各模块路由及编号。
func routeList(args []string) {
	routes, err := collectAllRoutes()
	if err != nil {
		fatalf("扫描路由失败: %v\n", err)
	}

	filter := ""
	if len(args) > 0 {
		filter = args[0]
	}

	moduleRoutes := make(map[string][]RouteDef)
	for _, r := range routes {
		if filter != "" && r.Module != filter {
			continue
		}
		moduleRoutes[r.Module] = append(moduleRoutes[r.Module], r)
	}

	for modName, modRoutes := range moduleRoutes {
		if modName == "" {
			continue
		}
		fmt.Printf("=== %s (%d 个路由) ===\n", modName, len(modRoutes))
		for _, r := range modRoutes {
			fmt.Printf("  %-45s %d\n", r.Name, r.Code)
		}
		fmt.Println()
	}

	if filter != "" && len(moduleRoutes) == 0 {
		fatalf("未找到模块: %s\n", filter)
	}
}
