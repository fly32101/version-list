package model

import (
	"testing"
	"time"
)

func TestInstallationContext_Creation(t *testing.T) {
	systemInfo := &SystemInfo{
		OS:       "windows",
		Arch:     "amd64",
		Version:  "1.21.0",
		Filename: "go1.21.0.windows-amd64.zip",
		URL:      "https://golang.org/dl/go1.21.0.windows-amd64.zip",
	}

	paths := &InstallPaths{
		BaseDir:     "/home/user/.go/versions",
		VersionDir:  "/home/user/.go/versions/1.21.0",
		TempDir:     "/tmp/go-install",
		ArchiveFile: "/tmp/go-install/go1.21.0.windows-amd64.zip",
	}

	options := &InstallOptions{
		Force:            false,
		CustomPath:       "",
		SkipVerification: false,
		Timeout:          300,
		MaxRetries:       3,
	}

	context := &InstallationContext{
		Version:    "1.21.0",
		SystemInfo: systemInfo,
		Paths:      paths,
		TempDir:    "/tmp/go-install",
		Options:    options,
		StartTime:  time.Now(),
		Status:     StatusPending,
	}

	// 验证上下文创建
	if context.Version != "1.21.0" {
		t.Errorf("Version = %s, 期望 %s", context.Version, "1.21.0")
	}

	if context.SystemInfo == nil {
		t.Error("SystemInfo 不应该为 nil")
	}

	if context.Paths == nil {
		t.Error("Paths 不应该为 nil")
	}

	if context.Options == nil {
		t.Error("Options 不应该为 nil")
	}

	if context.Status != StatusPending {
		t.Errorf("Status = %v, 期望 %v", context.Status, StatusPending)
	}
}

func TestDownloadInfo_Enhancement(t *testing.T) {
	downloadInfo := &DownloadInfo{
		URL:          "https://golang.org/dl/go1.21.0.windows-amd64.zip",
		Filename:     "go1.21.0.windows-amd64.zip",
		Size:         163553657,
		Checksum:     "abc123def456",
		ChecksumType: "sha256",
		DownloadedAt: time.Now(),
		Duration:     30000,      // 30秒
		Speed:        5451788.57, // 约5.2MB/s
	}

	// 验证下载信息
	if downloadInfo.URL == "" {
		t.Error("URL 不应该为空")
	}

	if downloadInfo.Size <= 0 {
		t.Error("Size 应该大于 0")
	}

	if downloadInfo.ChecksumType == "" {
		t.Error("ChecksumType 不应该为空")
	}

	if downloadInfo.Duration <= 0 {
		t.Error("Duration 应该大于 0")
	}

	if downloadInfo.Speed <= 0 {
		t.Error("Speed 应该大于 0")
	}
}

func TestExtractInfo_Creation(t *testing.T) {
	extractInfo := &ExtractInfo{
		ArchiveSize:   163553657,
		ExtractedSize: 400000000,
		FileCount:     1500,
		Duration:      5 * time.Second,
		RootDir:       "go",
	}

	// 验证解压信息
	if extractInfo.ArchiveSize <= 0 {
		t.Error("ArchiveSize 应该大于 0")
	}

	if extractInfo.ExtractedSize <= 0 {
		t.Error("ExtractedSize 应该大于 0")
	}

	if extractInfo.FileCount <= 0 {
		t.Error("FileCount 应该大于 0")
	}

	if extractInfo.Duration <= 0 {
		t.Error("Duration 应该大于 0")
	}

	if extractInfo.RootDir == "" {
		t.Error("RootDir 不应该为空")
	}
}

func TestValidationInfo_Validation(t *testing.T) {
	// 测试有效的验证信息
	validInfo := &ValidationInfo{
		ChecksumValid:   true,
		ExecutableValid: true,
		VersionValid:    true,
		Error:           "",
	}

	if !validInfo.ChecksumValid {
		t.Error("期望校验和有效")
	}

	if !validInfo.ExecutableValid {
		t.Error("期望可执行文件有效")
	}

	if !validInfo.VersionValid {
		t.Error("期望版本有效")
	}

	if validInfo.Error != "" {
		t.Error("期望没有错误信息")
	}

	// 测试无效的验证信息
	invalidInfo := &ValidationInfo{
		ChecksumValid:   false,
		ExecutableValid: true,
		VersionValid:    true,
		Error:           "校验和不匹配",
	}

	if invalidInfo.ChecksumValid {
		t.Error("期望校验和无效")
	}

	if invalidInfo.Error == "" {
		t.Error("期望有错误信息")
	}
}

func TestInstallationResult_Creation(t *testing.T) {
	downloadInfo := &DownloadInfo{
		URL:          "https://golang.org/dl/go1.21.0.windows-amd64.zip",
		Filename:     "go1.21.0.windows-amd64.zip",
		Size:         163553657,
		DownloadedAt: time.Now(),
		Duration:     30000,
		Speed:        5451788.57,
	}

	extractInfo := &ExtractInfo{
		ArchiveSize:   163553657,
		ExtractedSize: 400000000,
		FileCount:     1500,
		Duration:      5 * time.Second,
		RootDir:       "go",
	}

	result := &InstallationResult{
		Success:      true,
		Version:      "1.21.0",
		Path:         "/home/user/.go/versions/1.21.0",
		Duration:     35 * time.Second,
		Error:        "",
		DownloadInfo: downloadInfo,
		ExtractInfo:  extractInfo,
	}

	// 验证安装结果
	if !result.Success {
		t.Error("期望安装成功")
	}

	if result.Version != "1.21.0" {
		t.Errorf("Version = %s, 期望 %s", result.Version, "1.21.0")
	}

	if result.Path == "" {
		t.Error("Path 不应该为空")
	}

	if result.Duration <= 0 {
		t.Error("Duration 应该大于 0")
	}

	if result.DownloadInfo == nil {
		t.Error("DownloadInfo 不应该为 nil")
	}

	if result.ExtractInfo == nil {
		t.Error("ExtractInfo 不应该为 nil")
	}

	// 测试失败的安装结果
	failedResult := &InstallationResult{
		Success: false,
		Version: "1.20.0",
		Error:   "下载失败",
	}

	if failedResult.Success {
		t.Error("期望安装失败")
	}

	if failedResult.Error == "" {
		t.Error("期望有错误信息")
	}
}

func TestInstallOptions_DefaultValues(t *testing.T) {
	options := &InstallOptions{
		Force:            false,
		CustomPath:       "",
		SkipVerification: false,
		Timeout:          300,
		MaxRetries:       3,
	}

	// 验证默认值
	if options.Force {
		t.Error("期望 Force 默认为 false")
	}

	if options.CustomPath != "" {
		t.Error("期望 CustomPath 默认为空")
	}

	if options.SkipVerification {
		t.Error("期望 SkipVerification 默认为 false")
	}

	if options.Timeout != 300 {
		t.Errorf("Timeout = %d, 期望 %d", options.Timeout, 300)
	}

	if options.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, 期望 %d", options.MaxRetries, 3)
	}
}
