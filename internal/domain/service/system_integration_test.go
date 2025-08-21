package service

import (
	"testing"
	"time"
	"version-list/internal/domain/model"
)

// TestSystemDetectorIntegration 测试系统检测服务与其他组件的集成
func TestSystemDetectorIntegration(t *testing.T) {
	detector := NewSystemDetector()

	// 测试获取当前系统的Go 1.21.0版本信息
	version := "1.21.0"
	systemInfo, err := detector.GetSystemInfo(version)
	if err != nil {
		t.Fatalf("获取系统信息失败: %v", err)
	}

	// 验证系统信息可以用于创建GoVersion
	goVersion := &model.GoVersion{
		Version:  systemInfo.Version,
		Path:     "", // 安装路径稍后设置
		IsActive: false,
		Source:   model.SourceOnline,
		DownloadInfo: &model.DownloadInfo{
			URL:          systemInfo.URL,
			Filename:     systemInfo.Filename,
			Size:         0,           // 下载时设置
			Checksum:     "",          // 下载时设置
			DownloadedAt: time.Time{}, // 下载时设置
		},
	}

	// 验证GoVersion结构体的字段
	if goVersion.Version != version {
		t.Errorf("GoVersion.Version = %s, 期望 %s", goVersion.Version, version)
	}

	if goVersion.Source != model.SourceOnline {
		t.Errorf("GoVersion.Source = %v, 期望 %v", goVersion.Source, model.SourceOnline)
	}

	if goVersion.DownloadInfo == nil {
		t.Error("GoVersion.DownloadInfo 不能为 nil")
	} else {
		if goVersion.DownloadInfo.URL != systemInfo.URL {
			t.Errorf("DownloadInfo.URL = %s, 期望 %s", goVersion.DownloadInfo.URL, systemInfo.URL)
		}

		if goVersion.DownloadInfo.Filename != systemInfo.Filename {
			t.Errorf("DownloadInfo.Filename = %s, 期望 %s", goVersion.DownloadInfo.Filename, systemInfo.Filename)
		}
	}
}

// TestInstallationContextCreation 测试安装上下文的创建
func TestInstallationContextCreation(t *testing.T) {
	detector := NewSystemDetector()
	version := "1.21.0"

	systemInfo, err := detector.GetSystemInfo(version)
	if err != nil {
		t.Fatalf("获取系统信息失败: %v", err)
	}

	// 创建安装上下文
	installPaths := &model.InstallPaths{
		BaseDir:     "/home/user/.go/versions",
		VersionDir:  "/home/user/.go/versions/1.21.0",
		TempDir:     "/tmp/go-install",
		ArchiveFile: "/tmp/go-install/" + systemInfo.Filename,
	}

	context := &model.InstallationContext{
		Version:    version,
		SystemInfo: systemInfo,
		Paths:      installPaths,
		TempDir:    "/tmp/go-install",
	}

	// 验证安装上下文
	if context.Version != version {
		t.Errorf("InstallationContext.Version = %s, 期望 %s", context.Version, version)
	}

	if context.SystemInfo == nil {
		t.Error("InstallationContext.SystemInfo 不能为 nil")
	}

	if context.Paths == nil {
		t.Error("InstallationContext.Paths 不能为 nil")
	}

	if context.TempDir == "" {
		t.Error("InstallationContext.TempDir 不能为空")
	}
}
