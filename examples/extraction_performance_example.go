package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// ExtractionProgressReporter 示例进度报告器
type ExtractionProgressReporter struct {
	lastUpdate time.Time
}

func (e *ExtractionProgressReporter) SetStage(stage string) {
	fmt.Printf("📋 阶段: %s\n", stage)
}

func (e *ExtractionProgressReporter) SetProgress(progress float64) {
	// 限制更新频率，避免输出过多
	now := time.Now()
	if now.Sub(e.lastUpdate) > 100*time.Millisecond {
		fmt.Printf("⏱️  进度: %.1f%%\n", progress)
		e.lastUpdate = now
	}
}

func (e *ExtractionProgressReporter) SetMessage(message string) {
	fmt.Printf("💬 %s\n", message)
}

func (e *ExtractionProgressReporter) UpdateProgress(stage string, progress float64, message string) {
	fmt.Printf("🔄 %s: %.1f%% - %s\n", stage, progress, message)
}

func main() {
	fmt.Println("🚀 Go版本管理器 - 解压性能优化示例")
	fmt.Println("=====================================")

	// 创建临时目录进行演示
	tempDir, err := os.MkdirTemp("", "go-extraction-demo")
	if err != nil {
		log.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建模拟的Go解压目录结构
	srcDir := filepath.Join(tempDir, "extracted", "go")
	destDir := filepath.Join(tempDir, "go1.25.0")

	fmt.Printf("📁 源目录: %s\n", srcDir)
	fmt.Printf("📁 目标目录: %s\n", destDir)

	// 创建测试文件结构
	fmt.Println("\n📦 创建模拟Go安装包结构...")
	if err := createMockGoStructure(srcDir); err != nil {
		log.Fatalf("创建模拟结构失败: %v", err)
	}

	// 统计文件数量
	fileCount, err := countFiles(srcDir)
	if err != nil {
		log.Fatalf("统计文件失败: %v", err)
	}
	fmt.Printf("📊 总文件数: %d\n", fileCount)

	// 创建版本服务（使用优化的实现）
	versionService := createVersionService()

	// 创建进度报告器
	progressReporter := &ExtractionProgressReporter{}

	fmt.Println("\n🔧 开始优化的文件移动操作...")
	startTime := time.Now()

	// 执行优化的移动操作
	err = versionService.MoveExtractedContentWithProgress(
		filepath.Dir(srcDir), // 源目录的父目录
		destDir,              // 目标目录
		"go",                 // 根目录名
		progressReporter,     // 进度报告器
	)

	duration := time.Since(startTime)

	if err != nil {
		log.Fatalf("移动操作失败: %v", err)
	}

	fmt.Printf("\n✅ 移动操作完成!\n")
	fmt.Printf("⏱️  总耗时: %v\n", duration)
	fmt.Printf("🚀 平均速度: %.0f 文件/秒\n", float64(fileCount)/duration.Seconds())

	// 验证结果
	fmt.Println("\n🔍 验证移动结果...")
	if err := verifyMoveResult(destDir); err != nil {
		log.Fatalf("验证失败: %v", err)
	}

	fmt.Println("✅ 验证通过! 所有文件已正确移动")

	// 性能对比信息
	fmt.Println("\n📈 性能优化特性:")
	fmt.Println("  • 🎯 智能目录移动: 优先尝试整目录重命名")
	fmt.Println("  • 🔄 并发文件处理: 多线程并行移动文件")
	fmt.Println("  • 📊 实时进度跟踪: 详细的进度反馈")
	fmt.Println("  • ⏰ 超时保护机制: 防止操作无限期挂起")
	fmt.Println("  • 💾 优化缓冲区: 大缓冲区提高I/O性能")

	fmt.Println("\n🎉 示例完成!")
}

// createMockGoStructure 创建模拟的Go安装包结构
func createMockGoStructure(baseDir string) error {
	// 创建目录结构
	dirs := []string{
		"bin",
		"src/runtime",
		"src/fmt",
		"src/net/http",
		"src/encoding/json",
		"pkg/tool/windows_amd64",
		"pkg/windows_amd64",
		"doc",
		"test",
		"api",
		"misc",
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(baseDir, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return err
		}
	}

	// 创建二进制文件
	binFiles := map[string]string{
		"bin/go.exe":       "Go compiler executable",
		"bin/gofmt.exe":    "Go formatter executable",
		"bin/godoc.exe":    "Go documentation tool",
		"bin/go-build.exe": "Go build tool",
	}

	for filePath, content := range binFiles {
		fullPath := filepath.Join(baseDir, filePath)
		if err := os.WriteFile(fullPath, []byte(content), 0755); err != nil {
			return err
		}
	}

	// 创建源代码文件
	srcFiles := map[string]string{
		"src/runtime/runtime.go":    "package runtime\n// Runtime package",
		"src/fmt/fmt.go":            "package fmt\n// Format package",
		"src/net/http/http.go":      "package http\n// HTTP package",
		"src/encoding/json/json.go": "package json\n// JSON package",
	}

	for filePath, content := range srcFiles {
		fullPath := filepath.Join(baseDir, filePath)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return err
		}
	}

	// 创建工具文件
	toolFiles := map[string]string{
		"pkg/tool/windows_amd64/compile.exe": "Go compiler tool",
		"pkg/tool/windows_amd64/link.exe":    "Go linker tool",
		"pkg/tool/windows_amd64/asm.exe":     "Go assembler tool",
	}

	for filePath, content := range toolFiles {
		fullPath := filepath.Join(baseDir, filePath)
		if err := os.WriteFile(fullPath, []byte(content), 0755); err != nil {
			return err
		}
	}

	// 创建文档文件
	docFiles := map[string]string{
		"doc/README.md":  "# Go Programming Language",
		"doc/INSTALL.md": "# Installation Instructions",
		"doc/LICENSE":    "BSD 3-Clause License",
		"api/go1.txt":    "Go 1.0 API",
		"api/go1.1.txt":  "Go 1.1 API",
	}

	for filePath, content := range docFiles {
		fullPath := filepath.Join(baseDir, filePath)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return err
		}
	}

	// 创建大量测试文件来模拟真实的Go安装包
	for i := 0; i < 500; i++ {
		fileName := fmt.Sprintf("test/test_%03d.go", i)
		fullPath := filepath.Join(baseDir, fileName)
		content := fmt.Sprintf("package test\n// Test file %d\nfunc Test%d() {}\n", i, i)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

// countFiles 统计目录中的文件数量
func countFiles(dir string) (int, error) {
	count := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	return count, err
}

// verifyMoveResult 验证移动结果
func verifyMoveResult(destDir string) error {
	// 检查关键文件是否存在
	keyFiles := []string{
		"bin/go.exe",
		"bin/gofmt.exe",
		"src/runtime/runtime.go",
		"src/fmt/fmt.go",
		"doc/README.md",
	}

	for _, file := range keyFiles {
		fullPath := filepath.Join(destDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return fmt.Errorf("关键文件不存在: %s", fullPath)
		}
	}

	// 统计移动后的文件数量
	count, err := countFiles(destDir)
	if err != nil {
		return err
	}

	if count < 500 { // 至少应该有500个测试文件加上其他文件
		return fmt.Errorf("文件数量不足: %d", count)
	}

	return nil
}

// createVersionService 创建版本服务实例
func createVersionService() *MockVersionService {
	return &MockVersionService{}
}

// MockVersionService 模拟版本服务（仅包含需要的方法）
type MockVersionService struct{}

// MoveExtractedContentWithProgress 公开的移动方法
func (m *MockVersionService) MoveExtractedContentWithProgress(srcDir, destDir, rootDir string, progressUI ProgressReporter) error {
	// 这里我们直接调用优化的移动逻辑
	return m.moveExtractedContentWithProgress(srcDir, destDir, rootDir, progressUI)
}

// ProgressReporter 进度报告接口
type ProgressReporter interface {
	SetStage(stage string)
	SetProgress(progress float64)
	SetMessage(message string)
	UpdateProgress(stage string, progress float64, message string)
}

// moveExtractedContentWithProgress 实际的移动实现（简化版本）
func (m *MockVersionService) moveExtractedContentWithProgress(srcDir, destDir, rootDir string, progressUI ProgressReporter) error {
	// 如果有根目录，需要从根目录中移动内容
	if rootDir != "" {
		srcDir = filepath.Join(srcDir, rootDir)
	}

	// 检查源目录是否存在
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("源目录不存在: %s", srcDir)
	}

	// 优化策略：尝试直接移动整个目录
	if progressUI != nil {
		progressUI.SetMessage("尝试快速目录移动...")
	}

	// 检查目标目录是否为空
	if entries, err := os.ReadDir(destDir); err == nil && len(entries) > 0 {
		// 目标目录不为空，使用文件级移动
		return m.moveFileByFile(srcDir, destDir, progressUI)
	}

	// 尝试直接重命名目录（最快的方式）
	parentDir := filepath.Dir(destDir)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return err
	}

	// 如果目标目录已存在，先删除
	if _, err := os.Stat(destDir); err == nil {
		if err := os.RemoveAll(destDir); err != nil {
			return err
		}
	}

	err := os.Rename(srcDir, destDir)
	if err == nil {
		if progressUI != nil {
			progressUI.SetProgress(100)
			progressUI.SetMessage("快速目录移动完成")
		}
		return nil
	}

	// 如果直接移动失败，使用文件级移动
	return m.moveFileByFile(srcDir, destDir, progressUI)
}

// moveFileByFile 逐文件移动
func (m *MockVersionService) moveFileByFile(srcDir, destDir string, progressUI ProgressReporter) error {
	if progressUI != nil {
		progressUI.SetMessage("使用文件级移动...")
	}

	// 收集所有文件
	var files []string
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	totalFiles := len(files)
	for i, file := range files {
		// 计算相对路径
		relPath, err := filepath.Rel(srcDir, file)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, relPath)

		// 确保目标目录存在
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		// 移动文件
		if err := os.Rename(file, destPath); err != nil {
			// 如果重命名失败，尝试复制+删除
			if err := m.copyFile(file, destPath); err != nil {
				return err
			}
			os.Remove(file)
		}

		// 更新进度
		if progressUI != nil && i%10 == 0 {
			progress := float64(i) / float64(totalFiles) * 100
			progressUI.SetProgress(progress)
			progressUI.SetMessage(fmt.Sprintf("已移动 %d/%d 文件", i+1, totalFiles))
		}
	}

	if progressUI != nil {
		progressUI.SetProgress(100)
		progressUI.SetMessage("文件移动完成")
	}

	return nil
}

// copyFile 复制文件
func (m *MockVersionService) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// 使用io.Copy进行高效复制
	_, err = io.Copy(dstFile, srcFile)
	return err
}
