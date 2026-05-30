package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// insertIntoProto 在指定 proto 文件的 enum 块中插入一行。
// enumName 如 "AuthRoute"，entry 如 "\tLOGIN = 1001;"
func insertIntoProto(path, enumName, entry string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取 proto 失败: %w", err)
	}
	content := string(data)

	// 找到 enum 块的结束 '}'，在前面插入新行
	marker := "enum " + enumName
	idx := strings.Index(content, marker)
	if idx < 0 {
		return fmt.Errorf("未找到枚举 %s", enumName)
	}
	// 找到这个 enum 块最后的 '}'
	closeIdx := strings.Index(content[idx:], "}") + idx
	// 在 '}' 前一行插入（找到最近的换行）
	insertAt := closeIdx
	// 向前找最后一个非空行末尾
	for insertAt > idx && content[insertAt-1] == '\n' {
		insertAt--
	}

	// 在 enum 块最后一条条目和 } 之间插入
	newContent := content[:closeIdx] + entry + "\n" + content[closeIdx:]
	return os.WriteFile(path, []byte(newContent), 0644)
}

// addGoImport 在 Go 文件的 import 块中添加新的 import 行。
func addGoImport(path, importLine string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)

	// 找到 import ( 并在最后一个 ) 之前插入
	importStart := strings.Index(content, "import (")
	if importStart < 0 {
		return fmt.Errorf("未找到 import 块: %s", path)
	}
	importEnd := strings.Index(content[importStart:], ")") + importStart

	indent := "\t"
	newContent := content[:importEnd] + indent + importLine + "\n" + content[importEnd:]
	return os.WriteFile(path, []byte(newContent), 0644)
}

// appendGoBlock 在 Go 文件中指定标记位置后面追加代码块。
// marker 如 "// Auth 模块错误" ，block 是完整代码块。
func appendGoBlock(path, marker, block string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)

	idx := strings.Index(content, marker)
	if idx < 0 {
		return fmt.Errorf("未找到标记: %s", marker)
	}
	// 找到标记行末尾
	lineEnd := strings.Index(content[idx:], "\n") + idx
	newContent := content[:lineEnd+1] + block + "\n" + content[lineEnd+1:]
	return os.WriteFile(path, []byte(newContent), 0644)
}

// appendGoConst 在 Go 文件的 const 块中追加常量。
func appendGoConst(path, blockMarker, newConst string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)

	idx := strings.Index(content, blockMarker)
	if idx < 0 {
		// 在 const 块最后插入
		constIdx := strings.LastIndex(content, "const (")
		if constIdx < 0 {
			return fmt.Errorf("未找到 const 块")
		}
		closeIdx := strings.Index(content[constIdx:], ")") + constIdx
		newContent := content[:closeIdx] + "\n\n" + newConst + "\n" + content[closeIdx:]
		return os.WriteFile(path, []byte(newContent), 0644)
	}

	// 在标记块结束的 ) 之前插入
	end := strings.Index(content[idx:], ")") + idx
	newContent := content[:end] + "\t" + newConst + "\n" + content[end:]
	return os.WriteFile(path, []byte(newContent), 0644)
}

// appendGoModuleRun 在 cmd/node/main.go 的 WithModules() 中追加新模块。
func appendGoModuleRun(path, moduleName string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)

	// 在最后一个 ) 之前插入
	withModules := strings.Index(content, "stack.WithModules(")
	if withModules < 0 {
		return fmt.Errorf("未找到 WithModules 调用")
	}
	closeIdx := strings.Index(content[withModules:], "),") + withModules
	if closeIdx < withModules {
		// 无尾部逗号的情况
		closeIdx = strings.Index(content[withModules:], ")") + withModules
	}

	indent := "\t\t"
	moduleCall := fmt.Sprintf("%s%s.Module(),", indent, moduleName)
	newContent := content[:closeIdx] + "\n" + moduleCall + "\n\t" + content[closeIdx:]
	return os.WriteFile(path, []byte(newContent), 0644)
}

// addGoImportToMain 在 cmd/node/main.go 的 import 块中追加模块 import。
func addGoImportToMain(modulePath string) error {
	return addGoImport("cmd/node/main.go", fmt.Sprintf(`"%s"`, modulePath))
}

// runProtoc 运行 protoc 生成 proto 代码。
func runProtoc() error {
	cmd := exec.Command("bash", "gen_proto.sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// toSnakeCase 将驼峰名转为 proto 风格的下划线大写名。
// 例：GetInfo → GET_INFO, TokenRefresh → TOKEN_REFRESH
func toSnakeCase(s string) string {
	var result []byte
	for i, c := range s {
		if i > 0 && c >= 'A' && c <= 'Z' {
			if s[i-1] >= 'a' && s[i-1] <= 'z' {
				result = append(result, '_')
			} else if i+1 < len(s) && s[i+1] >= 'a' && s[i+1] <= 'z' {
				result = append(result, '_')
			}
		}
		result = append(result, byte(c))
	}
	return strings.ToUpper(string(result))
}

// goModulePath 返回模块的 Go import 路径。
func goModulePath(name string) string {
	return fmt.Sprintf("github.com/skeletongo/game-stack/module/%s", name)
}

// protoModulePath 返回模块 proto 的 Go import 路径。
func protoModulePath(name string) string {
	return fmt.Sprintf("github.com/skeletongo/game-stack/proto/%s", name)
}

// writeFileIfNotExist 仅在文件不存在时写入。
func writeFileIfNotExist(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("文件已存在: %s", path)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// protoEnumEntry 生成 proto 枚举条目。
func protoEnumEntry(name string, code int32, comment string) string {
	if comment != "" {
		return fmt.Sprintf("\t%-30s = %d;  // %s", name, code, comment)
	}
	return fmt.Sprintf("\t%-30s = %d;", name, code)
}

// goConstEntry 生成 Go 常量条目。
func goConstEntry(name string, value string) string {
	return fmt.Sprintf("%s int32 = %s", name, value)
}

// goErrEntry 生成 Go 错误码变量条目。
func goErrEntry(name string, protoRef string, message string) string {
	return fmt.Sprintf(`%s = &Code{Code: int32(%s), Message: "%s"}`, name, protoRef, message)
}
