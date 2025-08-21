package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestFullIntegration 测试系统检测、下载和解压服务的完整集成
func TestFullIntegration(t *testing.T) {
	// 创建系统检测服务
	detector := NewSystemDetector()

	// 创建下载服务
	downloadService := NewDownloadService(&DownloadOptions{
		MaxRetries: 2,
		RetryDelay: 100 * time.Millisecond,
		Timeout:    30 * time.Second,
		ChunkSize:  8192,
	})

	// 创建解压服务
	extractor := NewArchiveExtractor(&ArchiveExtractorOptions{
		PreservePermissions: true,
		OverwriteExisting:   true,
	})

	// 获取系统信息
	version := "1.19.13" // 使用一个较老但确定存在的版本
	systemInfo, err := detector.GetSystemInfo(version)
	if err != nil {
		t.Fatalf("获取系统信息失败: %v", err)
	}

	t.Logf("检测到系统: %s-%s", systemInfo.OS, systemInfo.Arch)
	t.Logf("下载URL: %s", systemInfo.URL)
	t.Logf("文件名: %s", systemInfo.Filename)

	// 验证压缩包类型检测
	expectedType := ArchiveTypeZip
	if systemInfo.OS != "windows" {
		expectedType = ArchiveTypeTarGz
	}

	actualType := extractor.GetArchiveType(systemInfo.Filename)
	if actualType != expectedType {
		t.Errorf("压缩包类型检测错误: 得到 %v, 期望 %v", actualType, expectedType)
	}

	// 创建临时目录
	tempDir := t.TempDir()
	extractDir := filepath.Join(tempDir, "extracted")

	// 检查文件大小（不实际下载大文件以节省时间）
	t.Log("检查远程文件大小...")
	size, err := downloadService.GetFileSize(systemInfo.URL)
	if err != nil {
		t.Logf("获取文件大小失败（可能是网络问题）: %v", err)
		t.Skip("跳过需要网络的测试")
	}

	t.Logf("远程文件大小: %d 字节 (%.2f MB)", size, float64(size)/(1024*1024))

	// 检查断点续传支持
	supportsResume, err := downloadService.SupportsResume(systemInfo.URL)
	if err != nil {
		t.Logf("检查断点续传支持失败: %v", err)
	} else {
		t.Logf("支持断点续传: %t", supportsResume)
	}

	// 由于Go安装包很大，我们不实际下载，而是创建一个模拟的压缩包进行测试
	t.Log("创建模拟压缩包进行解压测试...")

	var mockArchivePath string
	if expectedType == ArchiveTypeZip {
		mockArchivePath = filepath.Join(tempDir, "mock.zip")
		if err := createTestZip(mockArchivePath); err != nil {
			t.Fatalf("创建模拟ZIP文件失败: %v", err)
		}
	} else {
		mockArchivePath = filepath.Join(tempDir, "mock.tar.gz")
		if err := createTestTarGz(mockArchivePath); err != nil {
			t.Fatalf("创建模拟tar.gz文件失败: %v", err)
		}
	}

	// 验证模拟压缩包
	if err := extractor.ValidateArchive(mockArchivePath); err != nil {
		t.Fatalf("验证模拟压缩包失败: %v", err)
	}

	// 获取压缩包信息
	info, err := extractor.GetArchiveInfo(mockArchivePath)
	if err != nil {
		t.Fatalf("获取压缩包信息失败: %v", err)
	}

	t.Logf("压缩包信息: 类型=%v, 文件数=%d, 大小=%d字节",
		info.Type, info.FileCount, info.TotalSize)

	// 执行解压
	t.Log("开始解压模拟压缩包...")

	var progressCount int
	progressCallback := func(progress ExtractionProgress) {
		progressCount++
		if progressCount%2 == 0 { // 减少日志输出
			t.Logf("解压进度: %.1f%% (%d/%d)",
				progress.Percentage, progress.ProcessedFiles, progress.TotalFiles)
		}
	}

	err = extractor.Extract(mockArchivePath, extractDir, progressCallback)
	if err != nil {
		t.Fatalf("解压失败: %v", err)
	}

	t.Log("解压完成")

	// 验证解压结果
	expectedFiles := []string{
		"test_dir/file1.txt",
		"test_dir/file2.txt",
		"root_file.txt",
	}

	for _, expectedFile := range expectedFiles {
		filePath := filepath.Join(extractDir, expectedFile)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("期望的文件不存在: %s", expectedFile)
		}
	}

	// 验证进度回调被调用
	if progressCount == 0 {
		t.Error("解压进度回调未被调用")
	}

	t.Log("完整集成测试通过")
}

// TestServiceConfiguration 测试服务配置的兼容性
func TestServiceConfiguration(t *testing.T) {
	// 测试所有服务都能正常创建
	detector := NewSystemDetector()
	if detector == nil {
		t.Fatal("系统检测服务创建失败")
	}

	downloadService := NewDownloadService(nil)
	if downloadService == nil {
		t.Fatal("下载服务创建失败")
	}

	extractor := NewArchiveExtractor(nil)
	if extractor == nil {
		t.Fatal("解压服务创建失败")
	}

	// 测试基本功能
	version := "1.21.0"
	systemInfo, err := detector.GetSystemInfo(version)
	if err != nil {
		t.Fatalf("获取系统信息失败: %v", err)
	}

	archiveType := extractor.GetArchiveType(systemInfo.Filename)
	if archiveType == ArchiveTypeUnknown {
		t.Errorf("无法识别压缩包类型: %s", systemInfo.Filename)
	}

	t.Logf("服务配置测试通过: %s -> %s -> %v",
		version, systemInfo.Filename, archiveType)
}

// TestErrorHandling 测试错误处理
func TestErrorHandling(t *testing.T) {
	downloadService := NewDownloadService(&DownloadOptions{
		MaxRetries: 1,
		Timeout:    1 * time.Second,
	})
	extractor := NewArchiveExtractor(nil)

	// 测试无效URL
	_, err := downloadService.GetFileSize("http://invalid-url-that-does-not-exist.com")
	if err == nil {
		t.Error("期望无效URL返回错误")
	}

	// 测试无效压缩包
	tempDir := t.TempDir()
	invalidFile := filepath.Join(tempDir, "invalid.zip")
	if err := os.WriteFile(invalidFile, []byte("invalid content"), 0644); err != nil {
		t.Fatalf("创建无效文件失败: %v", err)
	}

	err = extractor.ValidateArchive(invalidFile)
	if err == nil {
		t.Error("期望验证无效压缩包时返回错误")
	}

	// 测试不存在的文件
	err = extractor.Extract("/path/that/does/not/exist.zip", tempDir, nil)
	if err == nil {
		t.Error("期望解压不存在的文件时返回错误")
	}

	t.Log("错误处理测试通过")
}
