package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// MockProgressReporter 模拟进度报告器
type MockProgressReporter struct {
	stage    string
	progress float64
	message  string
	updates  []string
}

func (m *MockProgressReporter) SetStage(stage string) {
	m.stage = stage
	m.updates = append(m.updates, "Stage: "+stage)
}

func (m *MockProgressReporter) SetProgress(progress float64) {
	m.progress = progress
	m.updates = append(m.updates, fmt.Sprintf("Progress: %.1f%%", progress))
}

func (m *MockProgressReporter) SetMessage(message string) {
	m.message = message
	m.updates = append(m.updates, "Message: "+message)
}

func (m *MockProgressReporter) UpdateProgress(stage string, progress float64, message string) {
	m.stage = stage
	m.progress = progress
	m.message = message
	m.updates = append(m.updates, fmt.Sprintf("Update: %s %.1f%% %s", stage, progress, message))
}

func TestVersionService_MoveExtractedContentPerformance(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "go-version-test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建模拟的解压目录结构
	srcDir := filepath.Join(tempDir, "src", "go")
	destDir := filepath.Join(tempDir, "dest")

	// 创建测试文件结构
	if err := createTestFileStructure(srcDir); err != nil {
		t.Fatalf("创建测试文件结构失败: %v", err)
	}

	// 创建版本服务
	versionRepo := &MockVersionRepository{}
	envRepo := &MockEnvironmentRepository{}
	service := NewVersionService(versionRepo, envRepo)

	// 创建进度报告器
	progressReporter := &MockProgressReporter{}

	// 测试移动操作
	startTime := time.Now()
	err = service.moveExtractedContentWithProgress(filepath.Dir(srcDir), destDir, "go", progressReporter)
	duration := time.Since(startTime)

	if err != nil {
		t.Fatalf("移动解压内容失败: %v", err)
	}

	// 验证文件是否正确移动
	if err := verifyMovedFiles(destDir); err != nil {
		t.Fatalf("验证移动文件失败: %v", err)
	}

	// 验证性能（应该在合理时间内完成）
	if duration > 30*time.Second {
		t.Errorf("移动操作耗时过长: %v", duration)
	}

	// 验证进度报告
	if len(progressReporter.updates) == 0 {
		t.Error("没有收到进度更新")
	}

	t.Logf("移动操作完成，耗时: %v", duration)
	t.Logf("进度更新次数: %d", len(progressReporter.updates))
}

func TestVersionService_MoveExtractedContentWithTimeout(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "go-version-timeout-test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	srcDir := filepath.Join(tempDir, "src", "go")
	destDir := filepath.Join(tempDir, "dest")

	// 创建测试文件结构
	if err := createTestFileStructure(srcDir); err != nil {
		t.Fatalf("创建测试文件结构失败: %v", err)
	}

	// 创建版本服务
	versionRepo := &MockVersionRepository{}
	envRepo := &MockEnvironmentRepository{}
	service := NewVersionService(versionRepo, envRepo)

	// 创建进度报告器
	progressReporter := &MockProgressReporter{}

	// 测试超时机制（使用很短的超时时间）
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	err = service.moveContentBatchWithTimeout(ctx, srcDir, destDir, progressReporter)

	// 应该返回超时错误
	if err == nil {
		t.Error("期望超时错误，但操作成功完成")
	}

	if err != nil && err.Error() != "文件移动操作超时: context deadline exceeded" {
		t.Logf("收到错误: %v", err) // 这是正常的，因为我们故意设置了很短的超时
	}
}

func TestVersionService_TryDirectoryMove(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "go-version-direct-move-test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	srcDir := filepath.Join(tempDir, "src")
	destDir := filepath.Join(tempDir, "dest")

	// 创建源目录和一些文件
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("创建源目录失败: %v", err)
	}

	testFile := filepath.Join(srcDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 创建版本服务
	versionRepo := &MockVersionRepository{}
	envRepo := &MockEnvironmentRepository{}
	service := NewVersionService(versionRepo, envRepo)

	// 测试直接目录移动
	err = service.tryDirectoryMove(srcDir, destDir)
	if err != nil {
		t.Fatalf("直接目录移动失败: %v", err)
	}

	// 验证文件是否存在于目标位置
	movedFile := filepath.Join(destDir, "test.txt")
	if _, err := os.Stat(movedFile); os.IsNotExist(err) {
		t.Error("文件未正确移动到目标位置")
	}

	// 验证源目录是否已删除
	if _, err := os.Stat(srcDir); !os.IsNotExist(err) {
		t.Error("源目录应该已被删除")
	}
}

// createTestFileStructure 创建测试文件结构
func createTestFileStructure(baseDir string) error {
	// 创建目录结构
	dirs := []string{
		"bin",
		"src/runtime",
		"src/fmt",
		"pkg/tool",
		"doc",
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(baseDir, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return err
		}
	}

	// 创建一些测试文件
	files := map[string]string{
		"bin/go":                 "#!/bin/bash\necho 'go'",
		"bin/gofmt":              "#!/bin/bash\necho 'gofmt'",
		"src/runtime/runtime.go": "package runtime",
		"src/fmt/fmt.go":         "package fmt",
		"pkg/tool/compile":       "compile tool",
		"doc/README.md":          "# Go Documentation",
	}

	for filePath, content := range files {
		fullPath := filepath.Join(baseDir, filePath)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return err
		}
	}

	// 创建更多文件来测试性能
	for i := 0; i < 100; i++ {
		fileName := fmt.Sprintf("src/test_file_%d.go", i)
		fullPath := filepath.Join(baseDir, fileName)
		content := fmt.Sprintf("package main\n// Test file %d\nfunc main() {}\n", i)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

// verifyMovedFiles 验证文件是否正确移动
func verifyMovedFiles(destDir string) error {
	expectedFiles := []string{
		"bin/go",
		"bin/gofmt",
		"src/runtime/runtime.go",
		"src/fmt/fmt.go",
		"pkg/tool/compile",
		"doc/README.md",
	}

	for _, file := range expectedFiles {
		fullPath := filepath.Join(destDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return fmt.Errorf("文件不存在: %s", fullPath)
		}
	}

	// 验证测试文件
	for i := 0; i < 100; i++ {
		fileName := fmt.Sprintf("src/test_file_%d.go", i)
		fullPath := filepath.Join(destDir, fileName)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return fmt.Errorf("测试文件不存在: %s", fullPath)
		}
	}

	return nil
}
