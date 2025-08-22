package service

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPathManager_GetDefaultInstallDir(t *testing.T) {
	pm := NewPathManager()
	defaultDir := pm.GetDefaultInstallDir()

	if defaultDir == "" {
		t.Error("默认安装目录不能为空")
	}

	// 检查路径是否包含预期的组件
	if !filepath.IsAbs(defaultDir) {
		t.Error("默认安装目录应该是绝对路径")
	}
}

func TestPathManager_ValidateInstallPath(t *testing.T) {
	pm := NewPathManager()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "空路径",
			path:    "",
			wantErr: true,
		},
		{
			name:    "相对路径",
			path:    "relative/path",
			wantErr: true,
		},
		{
			name:    "有效的绝对路径",
			path:    getValidAbsolutePath(),
			wantErr: false,
		},
	}

	// Windows特定测试
	if runtime.GOOS == "windows" {
		tests = append(tests, []struct {
			name    string
			path    string
			wantErr bool
		}{
			{
				name:    "包含非法字符的路径",
				path:    `C:\invalid<path`,
				wantErr: true,
			},
			{
				name:    "路径过长",
				path:    `C:\` + generateLongPath(300),
				wantErr: true,
			},
		}...)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pm.ValidateInstallPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInstallPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPathManager_CreateInstallDirectory(t *testing.T) {
	pm := NewPathManager()

	// 创建临时目录进行测试
	tempDir := filepath.Join(os.TempDir(), "go-version-test")
	defer os.RemoveAll(tempDir)

	testDir := filepath.Join(tempDir, "test-install")

	// 测试创建新目录
	err := pm.CreateInstallDirectory(testDir)
	if err != nil {
		t.Errorf("CreateInstallDirectory() error = %v", err)
	}

	// 验证目录是否创建成功
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Error("目录创建失败")
	}

	// 测试目录已存在的情况
	err = pm.CreateInstallDirectory(testDir)
	if err != nil {
		t.Errorf("CreateInstallDirectory() 对已存在的空目录应该成功, error = %v", err)
	}
}

func TestPathManager_GetVersionDirectory(t *testing.T) {
	pm := NewPathManager()

	baseDir := "/test/base"
	version := "1.21.0"
	expected := filepath.Join(baseDir, version)

	result := pm.GetVersionDirectory(baseDir, version)
	if result != expected {
		t.Errorf("GetVersionDirectory() = %v, want %v", result, expected)
	}
}

func TestPathManager_GetTempDirectory(t *testing.T) {
	pm := NewPathManager()

	version := "1.21.0"
	result := pm.GetTempDirectory(version)

	if result == "" {
		t.Error("临时目录路径不能为空")
	}

	if !filepath.IsAbs(result) {
		t.Error("临时目录路径应该是绝对路径")
	}

	// 检查路径是否包含版本信息
	if !strings.Contains(result, version) {
		t.Error("临时目录路径应该包含版本信息")
	}
}

func TestPathManager_CheckDiskSpace(t *testing.T) {
	pm := NewPathManager()

	// 使用临时目录进行测试
	tempDir := os.TempDir()

	// 测试正常情况（需要很少的空间）
	err := pm.CheckDiskSpace(tempDir, 1024) // 1KB
	if err != nil {
		t.Errorf("CheckDiskSpace() error = %v", err)
	}

	// 测试空间不足的情况（需要非常大的空间）
	err = pm.CheckDiskSpace(tempDir, 1024*1024*1024*1024*1024) // 1PB
	if err == nil {
		t.Error("CheckDiskSpace() 应该返回空间不足错误")
	}
}

func TestPathManager_IsDirectoryEmpty(t *testing.T) {
	pm := NewPathManager()

	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "go-version-empty-test")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	// 测试空目录
	isEmpty, err := pm.IsDirectoryEmpty(tempDir)
	if err != nil {
		t.Errorf("IsDirectoryEmpty() error = %v", err)
	}
	if !isEmpty {
		t.Error("空目录应该返回true")
	}

	// 在目录中创建文件
	testFile := filepath.Join(tempDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	// 测试非空目录
	isEmpty, err = pm.IsDirectoryEmpty(tempDir)
	if err != nil {
		t.Errorf("IsDirectoryEmpty() error = %v", err)
	}
	if isEmpty {
		t.Error("非空目录应该返回false")
	}
}

func TestPathValidator_ValidateAndPrepareInstallPath(t *testing.T) {
	pv := NewPathValidator()

	// 创建临时目录进行测试
	tempDir := filepath.Join(os.TempDir(), "go-version-validator-test")
	defer os.RemoveAll(tempDir)

	config := &InstallPathConfig{
		BaseDir:           tempDir,
		CreateIfNotExists: true,
		CheckPermissions:  true,
		CheckDiskSpace:    true,
		RequiredSpace:     1024, // 1KB
	}

	path, err := pv.ValidateAndPrepareInstallPath(config)
	if err != nil {
		t.Errorf("ValidateAndPrepareInstallPath() error = %v", err)
	}

	if path != tempDir {
		t.Errorf("ValidateAndPrepareInstallPath() = %v, want %v", path, tempDir)
	}

	// 验证目录是否创建
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("目录应该被创建")
	}
}

func TestPathInfo_String(t *testing.T) {
	info := &PathInfo{
		Path:           "/test/path",
		Exists:         true,
		IsDirectory:    true,
		IsEmpty:        true,
		AvailableSpace: 1024 * 1024 * 1024, // 1GB
	}

	result := info.String()
	if result == "" {
		t.Error("PathInfo.String() 不应该返回空字符串")
	}

	// 检查是否包含关键信息
	if !strings.Contains(result, info.Path) {
		t.Error("结果应该包含路径信息")
	}
}

// 辅助函数

func getValidAbsolutePath() string {
	// 使用临时目录作为父目录，这样肯定存在且可写
	tempDir := os.TempDir()
	if runtime.GOOS == "windows" {
		return filepath.Join(tempDir, "go-test-path")
	}
	return filepath.Join(tempDir, "go-test-path")
}

func generateLongPath(length int) string {
	path := ""
	for len(path) < length {
		path += "very-long-directory-name/"
	}
	return path
}
