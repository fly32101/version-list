package model

import (
	"testing"
	"time"
)

func TestGoVersion_IsOnlineInstalled(t *testing.T) {
	// 测试在线安装的版本
	onlineVersion := &GoVersion{
		Version: "1.21.0",
		Source:  SourceOnline,
		DownloadInfo: &DownloadInfo{
			URL:      "https://golang.org/dl/go1.21.0.windows-amd64.zip",
			Filename: "go1.21.0.windows-amd64.zip",
			Size:     100000,
		},
	}

	if !onlineVersion.IsOnlineInstalled() {
		t.Error("期望在线安装版本返回 true")
	}

	// 测试本地导入的版本
	localVersion := &GoVersion{
		Version: "1.20.0",
		Source:  SourceLocal,
	}

	if localVersion.IsOnlineInstalled() {
		t.Error("期望本地版本返回 false")
	}

	// 测试没有下载信息的在线版本
	incompleteOnlineVersion := &GoVersion{
		Version: "1.19.0",
		Source:  SourceOnline,
	}

	if incompleteOnlineVersion.IsOnlineInstalled() {
		t.Error("期望没有下载信息的在线版本返回 false")
	}
}

func TestGoVersion_IsValid(t *testing.T) {
	// 测试有效版本
	validVersion := &GoVersion{
		Version: "1.21.0",
		ValidationInfo: &ValidationInfo{
			ChecksumValid:   true,
			ExecutableValid: true,
			VersionValid:    true,
		},
	}

	if !validVersion.IsValid() {
		t.Error("期望有效版本返回 true")
	}

	// 测试无效版本
	invalidVersion := &GoVersion{
		Version: "1.20.0",
		ValidationInfo: &ValidationInfo{
			ChecksumValid:   false,
			ExecutableValid: true,
			VersionValid:    true,
		},
	}

	if invalidVersion.IsValid() {
		t.Error("期望无效版本返回 false")
	}

	// 测试没有验证信息的版本
	noValidationVersion := &GoVersion{
		Version: "1.19.0",
	}

	if noValidationVersion.IsValid() {
		t.Error("期望没有验证信息的版本返回 false")
	}
}

func TestGoVersion_GetDisplayName(t *testing.T) {
	// 测试有标签的版本
	taggedVersion := &GoVersion{
		Version: "1.21.0",
		Tags:    []string{"stable", "lts"},
	}

	expected := "1.21.0 (stable)"
	if taggedVersion.GetDisplayName() != expected {
		t.Errorf("GetDisplayName() = %s, 期望 %s", taggedVersion.GetDisplayName(), expected)
	}

	// 测试没有标签的版本
	untaggedVersion := &GoVersion{
		Version: "1.20.0",
	}

	if untaggedVersion.GetDisplayName() != "1.20.0" {
		t.Errorf("GetDisplayName() = %s, 期望 %s", untaggedVersion.GetDisplayName(), "1.20.0")
	}
}

func TestGoVersion_GetSizeInfo(t *testing.T) {
	version := &GoVersion{
		Version: "1.21.0",
		DownloadInfo: &DownloadInfo{
			Size: 100000,
		},
		ExtractInfo: &ExtractInfo{
			ExtractedSize: 200000,
		},
	}

	downloadSize, extractedSize := version.GetSizeInfo()

	if downloadSize != 100000 {
		t.Errorf("下载大小 = %d, 期望 %d", downloadSize, 100000)
	}

	if extractedSize != 200000 {
		t.Errorf("解压大小 = %d, 期望 %d", extractedSize, 200000)
	}
}

func TestGoVersion_MarkAsUsed(t *testing.T) {
	version := &GoVersion{
		Version:   "1.21.0",
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now().Add(-time.Hour),
	}

	oldUpdatedAt := version.UpdatedAt

	// 等待一小段时间确保时间戳不同
	time.Sleep(time.Millisecond)

	version.MarkAsUsed()

	if version.LastUsedAt == nil {
		t.Error("LastUsedAt 不应该为 nil")
	}

	if !version.UpdatedAt.After(oldUpdatedAt) {
		t.Error("UpdatedAt 应该被更新")
	}
}

func TestGoVersion_TagManagement(t *testing.T) {
	version := &GoVersion{
		Version: "1.21.0",
		Tags:    []string{},
	}

	// 测试添加标签
	version.AddTag("stable")
	if !version.HasTag("stable") {
		t.Error("期望版本有 'stable' 标签")
	}

	// 测试添加重复标签
	version.AddTag("stable")
	count := 0
	for _, tag := range version.Tags {
		if tag == "stable" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("期望只有一个 'stable' 标签，实际有 %d 个", count)
	}

	// 测试添加多个标签
	version.AddTag("lts")
	version.AddTag("recommended")

	if len(version.Tags) != 3 {
		t.Errorf("期望有 3 个标签，实际有 %d 个", len(version.Tags))
	}

	// 测试移除标签
	version.RemoveTag("lts")
	if version.HasTag("lts") {
		t.Error("期望 'lts' 标签被移除")
	}

	if len(version.Tags) != 2 {
		t.Errorf("期望有 2 个标签，实际有 %d 个", len(version.Tags))
	}

	// 测试移除不存在的标签
	version.RemoveTag("nonexistent")
	if len(version.Tags) != 2 {
		t.Errorf("移除不存在的标签后，期望仍有 2 个标签，实际有 %d 个", len(version.Tags))
	}
}

func TestInstallStatus_String(t *testing.T) {
	testCases := []struct {
		status   InstallStatus
		expected string
	}{
		{StatusPending, "pending"},
		{StatusDownloading, "downloading"},
		{StatusExtracting, "extracting"},
		{StatusConfiguring, "configuring"},
		{StatusCompleted, "completed"},
		{StatusFailed, "failed"},
		{StatusCancelled, "cancelled"},
		{InstallStatus(999), "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.status.String()
			if result != tc.expected {
				t.Errorf("InstallStatus(%d).String() = %s, 期望 %s", tc.status, result, tc.expected)
			}
		})
	}
}
