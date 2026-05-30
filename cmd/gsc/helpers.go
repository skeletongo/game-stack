package main

import (
	"path/filepath"
	"strings"
)

// fileBase 返回路径去掉目录和扩展名的文件名。
func fileBase(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// containsLower 不区分大小写检查 s 是否包含 sub。
func containsLower(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}
