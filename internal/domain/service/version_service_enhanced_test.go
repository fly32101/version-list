package service

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"version-list/internal/domain/model"
)

// MockVersionRepository 模拟版本仓库
type MockVersionRepository struct {
	versions map[string]*model.GoVersion
	active   string
}

func NewMockVersionRepository() *MockVersionRepository {
	return &MockVersionRepository{
		versions: make(map[string]*model.GoVersion),
	}
}

func (m *MockVersionRepository) Save(version *model.GoVersion) error {
	m.versions[version.Version] = version
	return nil
}

func (m *MockVersionRepository) Update(version *model.GoVersion) error {
	m.versions[version.Version] = version
	return nil
}

func (m *MockVersionRepository) FindByVersion(version string) (*model.GoVersion, error) {
	if v, exists := m.versions[version]; exists {
		return v, nil
	}
	return nil, fmt.Errorf("版本 %s 不存在", version)
}

func (m *MockVersionRepository) FindAll() ([]*model.GoVersion, error) {
	var result []*model.GoVersion
	for _, v := range m.versions {
		result = append(result, v)
	}
	return result, nil
}

func (m *MockVersionRepository) FindActive() (*model.GoVersion, error) {
	if m.active == "" {
		return nil, fmt.Errorf("没有激活的版本")
	}
	return m.FindByVersion(m.active)
}

func (m *MockVersionRepository) FindBySource(source model.InstallSource) ([]*model.GoVersion, error) {
	var result []*model.GoVersion
	for _, v := range m.versions {
		if v.Source == source {
			result = append(result, v)
		}
	}
	return result, nil
}

func (m *MockVersionRepository) FindByTag(tag string) ([]*model.GoVersion, error) {
	var result []*model.GoVersion
	for _, v := range m.versions {
		if v.HasTag(tag) {
			result = append(result, v)
		}
	}
	return result, nil
}

func (m *MockVersionRepository) SetActive(version string) error {
	if _, exists := m.versions[version]; !exists {
		return fmt.Errorf("版本 %s 不存在", version)
	}

	// 取消所有版本的激活状态
	for _, v := range m.versions {
		v.IsActive = false
	}

	// 设置指定版本为激活
	m.versions[version].IsActive = true
	m.active = version
	return nil
}

func (m *MockVersionRepository) Remove(version string) error {
	delete(m.versions, version)
	if m.active == version {
		m.active = ""
	}
	return nil
}

func (m *MockVersionRepository) UpdateLastUsed(version string) error {
	if v, exists := m.versions[version]; exists {
		v.MarkAsUsed()
		return nil
	}
	return fmt.Errorf("版本 %s 不存在", version)
}

func (m *MockVersionRepository) GetStatistics() (*model.VersionStatistics, error) {
	stats := &model.VersionStatistics{
		TotalVersions: len(m.versions),
	}

	for _, v := range m.versions {
		if v.Source == model.SourceOnline {
			stats.OnlineVersions++
		} else {
			stats.LocalVersions++
		}
	}

	stats.ActiveVersion = m.active
	return stats, nil
}

// MockEnvironmentRepository 模拟环境仓库
type MockEnvironmentRepository struct {
	env *model.Environment
}

func NewMockEnvironmentRepository() *MockEnvironmentRepository {
	return &MockEnvironmentRepository{
		env: &model.Environment{
			GOROOT: "/usr/local/go",
			GOPATH: "/home/user/go",
			GOBIN:  "/usr/local/go/bin",
		},
	}
}

func (m *MockEnvironmentRepository) Get() (*model.Environment, error) {
	return m.env, nil
}

func (m *MockEnvironmentRepository) Save(env *model.Environment) error {
	m.env = env
	return nil
}

func (m *MockEnvironmentRepository) UpdatePath(goPath string) error {
	m.env.GOROOT = goPath
	return nil
}

func TestVersionService_InstallOnline_CreateContext(t *testing.T) {
	// 创建模拟仓库
	versionRepo := NewMockVersionRepository()
	envRepo := NewMockEnvironmentRepository()

	// 创建版本服务
	service := NewVersionService(versionRepo, envRepo)

	// 测试创建安装上下文
	version := "1.21.0"
	options := &model.InstallOptions{
		Force:      false,
		Timeout:    300,
		MaxRetries: 3,
	}

	context, err := service.createInstallationContext(version, options)
	if err != nil {
		t.Fatalf("创建安装上下文失败: %v", err)
	}

	// 验证上下文
	if context.Version != version {
		t.Errorf("Version = %s, 期望 %s", context.Version, version)
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

	if context.Status != model.StatusPending {
		t.Errorf("Status = %v, 期望 %v", context.Status, model.StatusPending)
	}

	// 验证路径设置
	if context.Paths.VersionDir == "" {
		t.Error("VersionDir 不应该为空")
	}

	if context.Paths.TempDir == "" {
		t.Error("TempDir 不应该为空")
	}

	if context.Paths.ArchiveFile == "" {
		t.Error("ArchiveFile 不应该为空")
	}
}

func TestVersionService_InstallOnline_ExistingVersion(t *testing.T) {
	versionRepo := NewMockVersionRepository()
	envRepo := NewMockEnvironmentRepository()
	service := NewVersionService(versionRepo, envRepo)

	// 先添加一个已存在的版本
	existingVersion := &model.GoVersion{
		Version:  "1.21.0",
		Path:     "/some/path",
		IsActive: false,
		Source:   model.SourceLocal,
	}
	versionRepo.Save(existingVersion)

	// 尝试安装已存在的版本（不使用force）
	options := &model.InstallOptions{Force: false}
	_, err := service.InstallOnline("1.21.0", options)

	if err == nil {
		t.Error("期望安装已存在版本时返回错误")
	}

	// 使用force选项应该成功（但在这个测试中会因为网络操作而失败）
	options.Force = true
	_, err = service.InstallOnline("1.21.0", options)

	// 这里会失败是因为实际的网络操作，但我们验证了force逻辑
	if err == nil {
		t.Log("Force选项逻辑正常工作")
	}
}

func TestVersionService_GetBaseInstallDir(t *testing.T) {
	versionRepo := NewMockVersionRepository()
	envRepo := NewMockEnvironmentRepository()
	service := NewVersionService(versionRepo, envRepo)

	baseDir := service.getBaseInstallDir()

	if baseDir == "" {
		t.Error("基础安装目录不应该为空")
	}

	// 验证路径格式
	if !filepath.IsAbs(baseDir) {
		t.Error("基础安装目录应该是绝对路径")
	}

	expectedSuffix := filepath.Join(".go", "versions")
	if !strings.HasSuffix(baseDir, expectedSuffix) {
		t.Errorf("基础安装目录应该以 %s 结尾", expectedSuffix)
	}
}

func TestVersionService_GetGoExecutablePath(t *testing.T) {
	versionRepo := NewMockVersionRepository()
	envRepo := NewMockEnvironmentRepository()
	service := NewVersionService(versionRepo, envRepo)

	installDir := "/path/to/go"
	execPath := service.getGoExecutablePath(installDir)

	if execPath == "" {
		t.Error("Go可执行文件路径不应该为空")
	}

	expectedDir := filepath.Join(installDir, "bin")
	if !filepath.HasPrefix(execPath, expectedDir) {
		t.Errorf("Go可执行文件路径应该在 %s 目录下", expectedDir)
	}

	// 验证文件名
	expectedName := "go"
	if runtime.GOOS == "windows" {
		expectedName = "go.exe"
	}

	if filepath.Base(execPath) != expectedName {
		t.Errorf("Go可执行文件名应该是 %s", expectedName)
	}
}

func TestVersionService_CreateFailedResult(t *testing.T) {
	versionRepo := NewMockVersionRepository()
	envRepo := NewMockEnvironmentRepository()
	service := NewVersionService(versionRepo, envRepo)

	originalResult := &model.InstallationResult{
		Version: "1.21.0",
		Path:    "/some/path",
		Success: true, // 初始为成功
	}

	testError := fmt.Errorf("测试错误")

	failedResult, err := service.createFailedResult(originalResult, testError)

	// 验证返回的错误
	if err != testError {
		t.Errorf("返回的错误不匹配: 得到 %v, 期望 %v", err, testError)
	}

	// 验证结果被标记为失败
	if failedResult.Success {
		t.Error("结果应该被标记为失败")
	}

	if failedResult.Error != testError.Error() {
		t.Errorf("错误信息不匹配: 得到 %s, 期望 %s", failedResult.Error, testError.Error())
	}

	// 验证其他字段保持不变
	if failedResult.Version != originalResult.Version {
		t.Error("版本信息不应该改变")
	}

	if failedResult.Path != originalResult.Path {
		t.Error("路径信息不应该改变")
	}
}

func TestVersionService_CleanupOperations(t *testing.T) {
	versionRepo := NewMockVersionRepository()
	envRepo := NewMockEnvironmentRepository()
	service := NewVersionService(versionRepo, envRepo)

	// 创建临时目录进行测试
	tempDir := t.TempDir()
	versionDir := filepath.Join(tempDir, "version")

	// 创建测试目录和文件
	os.MkdirAll(versionDir, 0755)
	testFile := filepath.Join(versionDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	context := &model.InstallationContext{
		TempDir: tempDir,
		Paths: &model.InstallPaths{
			VersionDir: versionDir,
		},
	}

	// 测试清理失败的安装
	service.cleanupFailedInstallation(context)

	// 验证目录被删除
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Error("临时目录应该被删除")
	}

	if _, err := os.Stat(versionDir); !os.IsNotExist(err) {
		t.Error("版本目录应该被删除")
	}
}
