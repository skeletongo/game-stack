package tool

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// InsertLines 在文件的指定行号（1-based）后插入文本行。
// atLine 是插入位置之前的行号。若 atLine 为 0，则在文件头部插入。
func InsertLines(content []byte, atLine int, lines []string) []byte {
	if len(lines) == 0 {
		return content
	}

	text := string(content)
	existingLines := strings.Split(text, "\n")

	// 处理文件末尾无换行的情况
	hasTrailingNewline := strings.HasSuffix(text, "\n")

	var result []string
	for i, line := range existingLines {
		result = append(result, line)
		if i+1 == atLine {
			result = append(result, lines...)
		}
	}

	joined := strings.Join(result, "\n")
	if hasTrailingNewline && !strings.HasSuffix(joined, "\n") {
		joined += "\n"
	}
	return []byte(joined)
}

// AppendToEnd 在文件末尾追加文本行（确保换行）。
func AppendToEnd(content []byte, lines []string) []byte {
	text := string(content)
	// 确保文件以换行结尾
	for len(text) > 0 && text[len(text)-1] != '\n' {
		text += "\n"
	}
	for _, line := range lines {
		text += line + "\n"
	}
	return []byte(text)
}

// WriteFile 写入文件内容到磁盘，然后运行 gofmt 格式化。
func WriteFile(path string, content []byte) error {
	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("写入 %s 失败: %w", path, err)
	}
	return gofmt(path)
}

// InsertInFile 读取文件，在指定行后插入文本，写回并格式化。
func InsertInFile(path string, atLine int, lines []string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取 %s 失败: %w", path, err)
	}

	newData := InsertLines(data, atLine, lines)
	return WriteFile(path, newData)
}

// AppendToFile 在文件末尾追加文本，写回并格式化。
func AppendToFile(path string, lines []string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取 %s 失败: %w", path, err)
	}

	newData := AppendToEnd(data, lines)
	return WriteFile(path, newData)
}

// gofmt 对指定文件执行 gofmt -w。
func gofmt(path string) error {
	cmd := exec.Command("gofmt", "-w", path)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gofmt %s 失败: %s", path, stderr.String())
	}
	return nil
}

// SectionEndLine 返回文件内容中指定模块段（通过注释标记定位）的最后一行号。
// marker 是段头部注释的特征字符串，如 "Auth 模块"。
// 返回段末尾后空行之前最后一行的行号。
func SectionEndLine(content []byte, marker string) int {
	lines := strings.Split(string(content), "\n")
	inSection := false
	lastNonEmpty := 0

	for i, line := range lines {
		if strings.Contains(line, marker) && strings.HasPrefix(strings.TrimSpace(line), "//") {
			inSection = true
			lastNonEmpty = i + 1
			continue
		}
		if inSection {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				// 空行可能表示段之间分隔
				// 继续检查下一行是否是新段
				continue
			}
			if strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "// TODO") {
				// 可能是下一个段的开始，结束当前段
				break
			}
			lastNonEmpty = i + 1
		}
	}
	return lastNonEmpty
}
