package main

import (
	"fmt"
	"os"
	"text/tabwriter"
)

// moduleList 列出所有模块、路由数、错误码范围。
func moduleList(_ []string) {
	protoEnums, err := scanAllProtoEnums()
	if err != nil {
		fatalf("扫描 proto 失败: %v\n", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "模块名\t模块号\t路由数\t错误码数")
	fmt.Fprintln(w, "------\t------\t------\t--------")

	for path, enums := range protoEnums {
		name := fileBase(path)
		if name == "common" {
			continue
		}

		modNum := 0
		routeCount := 0
		errCount := 0

		for enumName, values := range enums {
			if isRouteEnum(enumName) {
				for _, v := range values {
					if v > 0 {
						routeCount++
					}
				}
				if len(values) > 0 {
					for _, v := range values {
						if v >= 1000 {
							modNum = int(v / 1000)
							break
						}
					}
				}
			}
			if isErrorEnum(enumName) {
				for _, v := range values {
					if v > 0 {
						errCount++
					}
				}
			}
		}

		fmt.Fprintf(w, "%s\t%d\t%d\t%d\n", name, modNum, routeCount, errCount)
	}

	w.Flush()
}

func isRouteEnum(name string) bool {
	return containsLower(name, "route")
}

func isErrorEnum(name string) bool {
	return containsLower(name, "error") || name == "SysError"
}
