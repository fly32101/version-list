package model

import (
	"strings"
	"testing"
	"time"
)

func TestGoVersion_ToJSON(t *testing.T) {
	version := createTestGoVersion()

	jsonStr, err := version.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() 返回错误: %v", err)
	}

	if jsonStr == "" {
		t.Error("JSON字符串不应该为空")
	}

	// 验证JSON包含关键字段
	if !strings.Contains(jsonStr, "1.21.0") {
		t.Error("JSON应该包含版本号")
	}

	// 打印JSON以调试
	t.Logf("JSON输出: %s", jsonStr)

	// InstallSource是枚举类型，在JSON中会是数字1（SourceOnline）
	if !strings.Contains(jsonStr, `"Source": 1`) {
		t.Error("JSON应该包含安装来源")
	}
}

func TestGoVersionFromJSON(t *testing.T) {
	originalVersion := createTestGoVersion()

	// 转换为JSON
	jsonStr, err := originalVersion.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() 返回错误: %v", err)
	}

	// 从JSON恢复
	restoredVersion, err := GoVersionFromJSON(jsonStr)
	if err != nil {
		t.Fatalf("FromJSON() 返回错误: %v", err)
	}

	// 验证关键字段
	if restoredVersion.Version != originalVersion.Version {
		t.Errorf("Version = %s, 期望 %s", restoredVersion.Version, originalVersion.Version)
	}

	if restoredVersion.Source != originalVersion.Source {
		t.Errorf("Source = %v, 期望 %v", restoredVersion.Source, originalVersion.Source)
	}

	if restoredVersion.IsActive != originalVersion.IsActive {
		t.Errorf("IsActive = %v, 期望 %v", restoredVersion.IsActive, originalVersion.IsActive)
	}
}

func TestGoVersion_ToMap(t *testing.T) {
	version := createTestGoVersion()

	versionMap := version.ToMap()

	// 验证基本字段
	if versionMap["version"] != "1.21.0" {
		t.Errorf("version = %v, 期望 %s", versionMap["version"], "1.21.0")
	}

	if versionMap["is_active"] != true {
		t.Errorf("is_active = %v, 期望 %v", versionMap["is_active"], true)
	}

	// 验证下载信息
	downloadInfo, ok := versionMap["download_info"].(map[string]interface{})
	if !ok {
		t.Error("download_info 应该是 map[string]interface{}")
	} else {
		if downloadInfo["filename"] != "go1.21.0.windows-amd64.zip" {
			t.Errorf("filename = %v, 期望 %s", downloadInfo["filename"], "go1.21.0.windows-amd64.zip")
		}
	}

	// 验证标签
	tags, ok := versionMap["tags"].([]string)
	if !ok {
		t.Error("tags 应该是 []string")
	} else {
		if len(tags) != 2 || tags[0] != "stable" {
			t.Errorf("tags = %v, 期望包含 'stable'", tags)
		}
	}
}

func TestGoVersion_Clone(t *testing.T) {
	original := createTestGoVersion()

	clone := original.Clone()

	// 验证基本字段
	if clone.Version != original.Version {
		t.Errorf("Clone.Version = %s, 期望 %s", clone.Version, original.Version)
	}

	if clone.IsActive != original.IsActive {
		t.Errorf("Clone.IsActive = %v, 期望 %v", clone.IsActive, original.IsActive)
	}

	// 验证深拷贝 - 修改克隆不应影响原始对象
	clone.Version = "1.22.0"
	if original.Version == "1.22.0" {
		t.Error("修改克隆对象不应影响原始对象")
	}

	// 验证标签的深拷贝
	clone.Tags[0] = "modified"
	if original.Tags[0] == "modified" {
		t.Error("修改克隆对象的标签不应影响原始对象")
	}

	// 验证下载信息的深拷贝
	if clone.DownloadInfo == original.DownloadInfo {
		t.Error("DownloadInfo 应该是深拷贝，不应该是同一个对象")
	}

	clone.DownloadInfo.URL = "modified-url"
	if original.DownloadInfo.URL == "modified-url" {
		t.Error("修改克隆对象的下载信息不应影响原始对象")
	}
}

func TestGoVersion_Validate(t *testing.T) {
	// 测试有效版本
	validVersion := createTestGoVersion()
	if err := validVersion.Validate(); err != nil {
		t.Errorf("有效版本验证失败: %v", err)
	}

	// 测试无效版本 - 空版本号
	invalidVersion1 := &GoVersion{
		Version: "",
		Path:    "/some/path",
		Source:  SourceLocal,
	}
	if err := invalidVersion1.Validate(); err == nil {
		t.Error("期望空版本号验证失败")
	}

	// 测试无效版本 - 空路径
	invalidVersion2 := &GoVersion{
		Version: "1.21.0",
		Path:    "",
		Source:  SourceLocal,
	}
	if err := invalidVersion2.Validate(); err == nil {
		t.Error("期望空路径验证失败")
	}

	// 测试无效版本 - 在线安装但没有下载信息
	invalidVersion3 := &GoVersion{
		Version: "1.21.0",
		Path:    "/some/path",
		Source:  SourceOnline,
	}
	if err := invalidVersion3.Validate(); err == nil {
		t.Error("期望在线安装但没有下载信息时验证失败")
	}

	// 测试无效版本 - 下载信息不完整
	invalidVersion4 := &GoVersion{
		Version: "1.21.0",
		Path:    "/some/path",
		Source:  SourceOnline,
		DownloadInfo: &DownloadInfo{
			URL:      "", // 空URL
			Filename: "test.zip",
			Size:     1000,
		},
	}
	if err := invalidVersion4.Validate(); err == nil {
		t.Error("期望下载URL为空时验证失败")
	}
}

func TestGoVersion_Summary(t *testing.T) {
	// 测试在线安装的活跃版本
	onlineActiveVersion := &GoVersion{
		Version:  "1.21.0",
		IsActive: true,
		Source:   SourceOnline,
		Tags:     []string{"stable", "lts"},
	}

	summary := onlineActiveVersion.Summary()
	expectedParts := []string{"Go 1.21.0", "(当前)", "在线安装", "[stable]"}

	for _, part := range expectedParts {
		if !strings.Contains(summary, part) {
			t.Errorf("摘要应该包含 '%s': %s", part, summary)
		}
	}

	// 测试本地安装的非活跃版本
	localInactiveVersion := &GoVersion{
		Version:  "1.20.0",
		IsActive: false,
		Source:   SourceLocal,
	}

	summary2 := localInactiveVersion.Summary()
	if strings.Contains(summary2, "(当前)") {
		t.Error("非活跃版本的摘要不应包含 '(当前)'")
	}

	if !strings.Contains(summary2, "本地安装") {
		t.Error("本地版本的摘要应该包含 '本地安装'")
	}
}

// createTestGoVersion 创建测试用的GoVersion对象
func createTestGoVersion() *GoVersion {
	now := time.Now()
	return &GoVersion{
		Version:  "1.21.0",
		Path:     "/home/user/.go/versions/1.21.0",
		IsActive: true,
		Source:   SourceOnline,
		DownloadInfo: &DownloadInfo{
			URL:          "https://golang.org/dl/go1.21.0.windows-amd64.zip",
			Filename:     "go1.21.0.windows-amd64.zip",
			Size:         163553657,
			Checksum:     "abc123def456",
			ChecksumType: "sha256",
			DownloadedAt: now,
			Duration:     30000,
			Speed:        5451788.57,
		},
		ExtractInfo: &ExtractInfo{
			ArchiveSize:   163553657,
			ExtractedSize: 400000000,
			FileCount:     1500,
			Duration:      5 * time.Second,
			RootDir:       "go",
		},
		ValidationInfo: &ValidationInfo{
			ChecksumValid:   true,
			ExecutableValid: true,
			VersionValid:    true,
		},
		SystemInfo: &SystemInfo{
			OS:       "windows",
			Arch:     "amd64",
			Version:  "1.21.0",
			Filename: "go1.21.0.windows-amd64.zip",
			URL:      "https://golang.org/dl/go1.21.0.windows-amd64.zip",
		},
		CreatedAt:       now.Add(-time.Hour),
		UpdatedAt:       now,
		LastUsedAt:      &now,
		InstallDuration: 35 * time.Second,
		Tags:            []string{"stable", "lts"},
		Notes:           "测试版本",
	}
}
