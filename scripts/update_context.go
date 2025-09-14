package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// 批量替换 context.Background() 为 middleware.GetRequestContext(c) 的脚本
func main() {
	// 定义要处理的目录
	handlersDir := "../internal/handlers"

	// 遍历handlers目录下的所有.go文件
	err := filepath.WalkDir(handlersDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 只处理.go文件
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// 跳过测试文件
		if strings.Contains(path, "_test.go") {
			return nil
		}

		fmt.Printf("Processing file: %s\n", path)
		return processFile(path)
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Context replacement completed!")
}

func processFile(filePath string) error {
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	originalContent := string(content)
	modifiedContent := originalContent

	// 检查是否包含 gin.Context 参数的函数
	hasGinContext := strings.Contains(modifiedContent, "func (") && strings.Contains(modifiedContent, "*gin.Context")

	if !hasGinContext {
		fmt.Printf("  Skipping %s - no gin.Context handlers found\n", filePath)
		return nil
	}

	// 替换 context.Background() 为 middleware.GetRequestContext(c)
	contextBackgroundPattern := regexp.MustCompile(`context\.Background\(\)`)
	modifiedContent = contextBackgroundPattern.ReplaceAllString(modifiedContent, "middleware.GetRequestContext(c)")

	// 检查是否需要添加 context 包的导入（如果移除了 context.Background() 的话）
	if strings.Contains(originalContent, "context.Background()") && !strings.Contains(modifiedContent, "context.") {
		// 移除 context 包的导入
		lines := strings.Split(modifiedContent, "\n")
		var newLines []string
		inImportBlock := false

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)

			// 检测导入块
			if strings.Contains(trimmed, "import (") {
				inImportBlock = true
				newLines = append(newLines, line)
				continue
			}

			if inImportBlock && trimmed == ")" {
				inImportBlock = false
				newLines = append(newLines, line)
				continue
			}

			// 在导入块中，跳过单独的 context 导入
			if inImportBlock && trimmed == `"context"` {
				fmt.Printf("  Removing context import from %s\n", filePath)
				continue
			}

			newLines = append(newLines, line)
		}

		modifiedContent = strings.Join(newLines, "\n")
	}

	// 如果内容有变化，写回文件
	if modifiedContent != originalContent {
		err = os.WriteFile(filePath, []byte(modifiedContent), 0644)
		if err != nil {
			return fmt.Errorf("failed to write file %s: %w", filePath, err)
		}
		fmt.Printf("  Updated %s\n", filePath)
	} else {
		fmt.Printf("  No changes needed for %s\n", filePath)
	}

	return nil
}
