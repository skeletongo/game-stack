// gsc — Game-Stack 代码生成工具。
//
// 自动生成模块骨架、路由常量、错误码、proto 协议，
// 保证路由号和错误码唯一，自动分配编号。
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "module":
		moduleCmd(os.Args[2:])
	case "route":
		routeCmd(os.Args[2:])
	case "errcode":
		errcodeCmd(os.Args[2:])
	default:
		printUsage()
		os.Exit(1)
	}
}

// moduleCmd 模块子命令总控。路由到 new / list。
func moduleCmd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "用法: gsc module <new|list> [...]")
		os.Exit(1)
	}
	switch args[0] {
	case "new":
		moduleNew(args[1:])
	case "list":
		moduleList(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "未知的子命令: %s\n", args[0])
		os.Exit(1)
	}
}

// routeCmd 路由子命令总控。路由到 add / list / check。
func routeCmd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "用法: gsc route <add|list|check> [...]")
		os.Exit(1)
	}
	switch args[0] {
	case "add":
		routeAdd(args[1:])
	case "list":
		routeList(args[1:])
	case "check":
		routeCheck(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "未知的子命令: %s\n", args[0])
		os.Exit(1)
	}
}

// errcodeCmd 错误码子命令总控。路由到 add / list。
func errcodeCmd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "用法: gsc errcode <add|list> [...]")
		os.Exit(1)
	}
	switch args[0] {
	case "add":
		errcodeAdd(args[1:])
	case "list":
		errcodeList(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "未知的子命令: %s\n", args[0])
		os.Exit(1)
	}
}

// fatalf 输出错误消息到 stderr 并退出。
func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func printUsage() {
	fmt.Println(`gsc — Game-Stack 代码生成工具

用法:
  gsc module new <模块名> [-n 模块号]
      创建新模块（proto + 6 文件 + route/errcode 注册）

  gsc module list
      列出所有模块及其路由数、错误码范围

  gsc route add -m <模块名> <操作名> [路由号]
      为已有模块添加路由，省略路由号时自动分配

  gsc route list [模块名]
      列出模块路由及编号缺口

  gsc route check
      检查路由号冲突、越界、命名规范

  gsc errcode add [-m <模块名>] <名称> <描述> [错误码]
      添加错误码。不带 -m 为系统错误，带 -m 为模块错误
      省略错误码时自动分配`)
}
