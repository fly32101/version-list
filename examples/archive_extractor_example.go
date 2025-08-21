package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"version-list/internal/domain/service"
)

func main() {
	fmt.Println("=== 压缩包解压服务示例 ===")

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "extractor_test")
	if err != nil {
		log.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建示例ZIP文件
	zipPath := filepath.Join(tempDir, "example.zip")
	if err := createExampleZip(zipPath); err != nil {
		log.Fatalf("创建示例ZIP文件失败: %v", err)
	}

	fmt.Printf("创建的ZIP文件: %s\n", zipPath)

	// 创建解压服务
	extractor := service.NewArchiveExtractor(&service.ArchiveExtractorOptions{
		PreservePermissions: true,
		OverwriteExisting:   true,
	})

	// 检测压缩包类型
	archiveType := extractor.GetArchiveType(zipPath)
	fmt.Printf("压缩包类型: %v\n", archiveType)

	// 获取压缩包信息
	fmt.Println("\n获取压缩包信息...")
	info, err := extractor.GetArchiveInfo(zipPath)
	if err != nil {
		log.Fatalf("获取压缩包信息失败: %v", err)
	}

	fmt.Printf("文件数量: %d\n", info.FileCount)
	fmt.Printf("解压后总大小: %d 字节\n", info.TotalSize)
	fmt.Printf("根目录: %s\n", info.RootDir)

	// 验证压缩包
	fmt.Println("\n验证压缩包完整性...")
	if err := extractor.ValidateArchive(zipPath); err != nil {
		log.Fatalf("压缩包验证失败: %v", err)
	}
	fmt.Println("压缩包验证通过")

	// 创建解压目录
	extractDir := filepath.Join(tempDir, "extracted")
	fmt.Printf("\n解压到目录: %s\n", extractDir)

	// 设置进度回调
	fmt.Println("\n开始解压...")
	progressCallback := func(progress service.ExtractionProgress) {
		fmt.Printf("\r进度: %.1f%% (%d/%d 文件) 当前: %s",
			progress.Percentage,
			progress.ProcessedFiles,
			progress.TotalFiles,
			progress.CurrentFile)
	}

	// 执行解压
	err = extractor.Extract(zipPath, extractDir, progressCallback)
	if err != nil {
		log.Fatalf("解压失败: %v", err)
	}

	fmt.Println("\n解压完成!")

	// 列出解压后的文件
	fmt.Println("\n解压后的文件:")
	err = filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(extractDir, path)
		if relPath == "." {
			return nil
		}

		if info.IsDir() {
			fmt.Printf("  📁 %s/\n", relPath)
		} else {
			fmt.Printf("  📄 %s (%d 字节)\n", relPath, info.Size())
		}

		return nil
	})

	if err != nil {
		log.Printf("遍历解压目录失败: %v", err)
	}

	// 验证解压后的文件内容
	fmt.Println("\n验证解压后的文件内容:")
	testFile := filepath.Join(extractDir, "example_project", "src", "main.go")
	if content, err := os.ReadFile(testFile); err == nil {
		fmt.Printf("main.go 内容:\n%s\n", string(content))
	}

	fmt.Println("\n=== 示例完成 ===")
}

// createExampleZip 创建示例ZIP文件
func createExampleZip(zipPath string) error {
	file, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// 创建项目结构
	files := map[string]string{
		"example_project/README.md": `# Example Project

This is an example Go project for testing archive extraction.

## Structure

- src/: Source code
- docs/: Documentation
- tests/: Test files
`,
		"example_project/src/main.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
	fmt.Println("This is an example Go application.")
}
`,
		"example_project/src/utils.go": `package main

import "strings"

func ToUpper(s string) string {
	return strings.ToUpper(s)
}

func ToLower(s string) string {
	return strings.ToLower(s)
}
`,
		"example_project/docs/api.md": `# API Documentation

## Functions

### ToUpper(s string) string
Converts a string to uppercase.

### ToLower(s string) string  
Converts a string to lowercase.
`,
		"example_project/tests/main_test.go": `package main

import "testing"

func TestToUpper(t *testing.T) {
	result := ToUpper("hello")
	if result != "HELLO" {
		t.Errorf("Expected HELLO, got %s", result)
	}
}
`,
		"example_project/go.mod": `module example_project

go 1.21
`,
	}

	// 添加目录
	dirs := []string{
		"example_project/",
		"example_project/src/",
		"example_project/docs/",
		"example_project/tests/",
	}

	for _, dir := range dirs {
		if _, err := zipWriter.Create(dir); err != nil {
			return err
		}
	}

	// 添加文件
	for filename, content := range files {
		writer, err := zipWriter.Create(filename)
		if err != nil {
			return err
		}

		if _, err := io.WriteString(writer, content); err != nil {
			return err
		}
	}

	return nil
}
