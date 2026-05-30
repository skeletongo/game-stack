package tool

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FindProjectRoot 从当前目录向上查找包含 game-stack 模块的 go.mod，返回项目根目录的绝对路径。
func FindProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取当前目录失败: %w", err)
	}
	dir := cwd
	for {
		modPath := filepath.Join(dir, "go.mod")
		data, err := os.ReadFile(modPath)
		if err == nil {
			if strings.Contains(string(data), "github.com/skeletongo/game-stack") {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("未找到 game-stack 项目根目录（从 %s 向上未找到包含 game-stack 的 go.mod）", cwd)
		}
		dir = parent
	}
}

// StackDir 返回项目根目录下的 stack/ 绝对路径。
func StackDir(root string) string { return filepath.Join(root, "stack") }

// ModuleDir 返回指定模块的绝对路径。
func ModuleDir(root, name string) string { return filepath.Join(root, "module", name) }

// ProtocolDir 返回指定模块协议包的绝对路径。
func ProtocolDir(root, name string) string { return filepath.Join(root, "protocol", name) }

// CmdNodeMain 返回 cmd/node/main.go 的绝对路径。
func CmdNodeMain(root string) string { return filepath.Join(root, "cmd", "node", "main.go") }

// RouteGo 返回 stack/route.go 的绝对路径。
func RouteGo(root string) string { return filepath.Join(root, "stack", "route.go") }

// ErrCodeGo 返回 stack/errcode.go 的绝对路径。
func ErrCodeGo(root string) string { return filepath.Join(root, "stack", "errcode.go") }
