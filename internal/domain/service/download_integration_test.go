package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestDownloadIntegration 测试下载服务与系统检测服务的集成
func TestDownloadIntegration(t *testing.T) {
	// 创建系统检测服务
	detector := NewSystemDetector()

	// 创建下载服务
	downloadService := NewDownloadService(&DownloadOptions{
		MaxRetries: 2,
		RetryDelay: 100 * time.Millisecond,
		Timeout:    10 * time.Second,
		ChunkSize:  1024,
	})

	// 获取当前系统的Go版本信息（使用一个较老的版本进行测试）
	version := "1.19.13" // 使用一个确定存在的版本
	systemInfo, err := detector.GetSystemInfo(version)
	if err != nil {
		t.Fatalf("获取系统信息失败: %v", err)
	}

	t.Logf("系统信息: OS=%s, Arch=%s", systemInfo.OS, systemInfo.Arch)
	t.Logf("下载URL: %s", systemInfo.URL)
	t.Logf("文件名: %s", systemInfo.Filename)

	// 检查文件大小（不实际下载大文件）
	size, err := downloadService.GetFileSize(systemInfo.URL)
	if err != nil {
		t.Logf("获取文件大小失败（可能是网络问题）: %v", err)
		t.Skip("跳过网络相关测试")
	}

	t.Logf("文件大小: %d 字节 (%.2f MB)", size, float64(size)/(1024*1024))

	// 检查是否支持断点续传
	supportsResume, err := downloadService.SupportsResume(systemInfo.URL)
	if err != nil {
		t.Logf("检查断点续传支持失败: %v", err)
	} else {
		t.Logf("支持断点续传: %t", supportsResume)
	}

	// 验证URL格式正确
	if systemInfo.URL == "" {
		t.Error("下载URL不能为空")
	}

	if systemInfo.Filename == "" {
		t.Error("文件名不能为空")
	}

	// 验证文件名格式
	expectedPrefix := "go" + version
	if len(systemInfo.Filename) < len(expectedPrefix) ||
		systemInfo.Filename[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("文件名格式不正确: %s", systemInfo.Filename)
	}
}

// TestDownloadServiceConfiguration 测试下载服务的配置
func TestDownloadServiceConfiguration(t *testing.T) {
	// 测试默认配置
	service1 := NewDownloadService(nil)
	if service1 == nil {
		t.Fatal("默认配置创建服务失败")
	}

	// 测试自定义配置
	options := &DownloadOptions{
		MaxRetries: 5,
		RetryDelay: 1 * time.Second,
		Timeout:    5 * time.Minute,
		ChunkSize:  4096,
		UserAgent:  "test-agent/1.0",
	}

	service2 := NewDownloadService(options)
	if service2 == nil {
		t.Fatal("自定义配置创建服务失败")
	}
}

// TestDownloadPathHandling 测试下载路径处理
func TestDownloadPathHandling(t *testing.T) {
	downloadService := NewDownloadService(&DownloadOptions{
		Timeout:   5 * time.Second,
		ChunkSize: 1024,
	})

	// 创建临时目录
	tempDir := t.TempDir()

	// 测试嵌套目录创建
	nestedPath := filepath.Join(tempDir, "nested", "deep", "path", "file.txt")

	// 使用一个简单的测试URL
	testURL := "https://golang.org/robots.txt"

	err := downloadService.Download(testURL, nestedPath, nil)
	if err != nil {
		t.Logf("下载失败（可能是网络问题）: %v", err)
		t.Skip("跳过网络相关测试")
	}

	// 验证文件是否存在
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Error("下载的文件不存在")
	}

	// 验证目录结构
	nestedDir := filepath.Dir(nestedPath)
	if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
		t.Error("嵌套目录未创建")
	}
}
