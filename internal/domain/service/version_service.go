package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"version-list/internal/domain/model"
	"version-list/internal/domain/repository"
)

// VersionInfo 版本信息视图模型
type VersionInfo struct {
	Version  string // 版本号
	Path     string // 安装路径
	IsActive bool   // 是否为当前激活版本
}

// VersionService 版本管理领域服务
type VersionService struct {
	versionRepo     repository.VersionRepository
	environmentRepo repository.EnvironmentRepository
}

// NewVersionService 创建版本服务实例
func NewVersionService(versionRepo repository.VersionRepository, environmentRepo repository.EnvironmentRepository) *VersionService {
	return &VersionService{
		versionRepo:     versionRepo,
		environmentRepo: environmentRepo,
	}
}

// Install 安装指定版本的Go
func (s *VersionService) Install(version string) error {
	// 检查版本是否已存在
	_, err := s.versionRepo.FindByVersion(version)
	if err == nil {
		return fmt.Errorf("Go版本 %s 已安装", version)
	}

	// 创建新版本记录
	newVersion := &model.GoVersion{
		Version:  version,
		IsActive: false,
	}

	// 保存版本记录
	return s.versionRepo.Save(newVersion)
}

// List 列出所有已安装的Go版本
func (s *VersionService) List() ([]*model.GoVersion, error) {
	return s.versionRepo.FindAll()
}

// Use 切换到指定版本的Go
func (s *VersionService) Use(version string) error {
	// 检查版本是否存在
	targetVersion, err := s.versionRepo.FindByVersion(version)
	if err != nil {
		return fmt.Errorf("Go版本 %s 未安装", version)
	}

	// 获取环境变量配置
	env, err := s.environmentRepo.Get()
	if err != nil {
		return fmt.Errorf("获取环境变量配置失败: %v", err)
	}

	// 创建符号链接目录
	symlinkDir := filepath.Join(os.Getenv("HOME"), ".go-version", "current")
	if runtime.GOOS == "windows" {
		// Windows使用不同的路径
		homeDir, _ := os.UserHomeDir()
		symlinkDir = filepath.Join(homeDir, ".go-version", "current")
	}

	// 确保符号链接目录的父目录存在
	if err := os.MkdirAll(filepath.Dir(symlinkDir), 0755); err != nil {
		return fmt.Errorf("创建符号链接目录失败: %v", err)
	}

	// 如果符号链接已存在，先删除
	if _, err := os.Lstat(symlinkDir); err == nil {
		if err := os.Remove(symlinkDir); err != nil {
			return fmt.Errorf("删除旧符号链接失败: %v", err)
		}
	}

	// 创建新的符号链接
	// 注意：这里需要确定实际的Go安装路径
	goInstallPath := targetVersion.Path
	if goInstallPath == "" {
		// 如果没有指定路径，使用默认路径
		homeDir, _ := os.UserHomeDir()
		goInstallPath = filepath.Join(homeDir, ".go", "versions", version)
	}

	// 检查目标路径是否存在
	if _, err := os.Stat(goInstallPath); os.IsNotExist(err) {
		return fmt.Errorf("Go版本 %s 的安装路径不存在: %s", version, goInstallPath)
	}

	// 创建符号链接
	if runtime.GOOS == "windows" {
		// Windows使用mklink命令
		cmd := exec.Command("cmd", "/C", "mklink", "/J", symlinkDir, goInstallPath)
		if err := cmd.Run(); err != nil {
			// 如果mklink失败，尝试使用powershell
			cmd = exec.Command("powershell", "-Command", "New-Item", "-ItemType", "SymbolicLink", "-Path", symlinkDir, "-Target", goInstallPath)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("创建符号链接失败: %v", err)
			}
		}
	} else {
		// Unix-like系统使用symlink
		if err := os.Symlink(goInstallPath, symlinkDir); err != nil {
			return fmt.Errorf("创建符号链接失败: %v", err)
		}
	}

	// 更新环境变量配置，使用符号链接路径
	env.GOROOT = symlinkDir
	env.GOBIN = filepath.Join(symlinkDir, "bin")

	// 保存环境变量配置
	if err := s.environmentRepo.Save(env); err != nil {
		return fmt.Errorf("保存环境变量配置失败: %v", err)
	}

	// 设置指定版本为激活状态
	if err := s.versionRepo.SetActive(version); err != nil {
		return fmt.Errorf("设置激活版本失败: %v", err)
	}

	return nil
}

// Current 获取当前使用的Go版本
func (s *VersionService) Current() (*model.GoVersion, error) {
	return s.versionRepo.FindActive()
}

// Remove 移除指定版本的Go
func (s *VersionService) Remove(version string) error {
	// 检查版本是否存在
	_, err := s.versionRepo.FindByVersion(version)
	if err != nil {
		return fmt.Errorf("Go版本 %s 未安装", version)
	}

	// 检查是否为当前激活版本
	activeVersion, err := s.versionRepo.FindActive()
	if err == nil && activeVersion.Version == version {
		return fmt.Errorf("不能移除当前使用的Go版本 %s", version)
	}

	// 移除版本
	return s.versionRepo.Remove(version)
}

// ImportLocal 导入本地已安装的Go版本
func (s *VersionService) ImportLocal(path string) (string, error) {
	// 检查路径是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("路径 %s 不存在", path)
	}

	// 尝试从路径中获取Go版本号
	version, err := s.extractVersionFromPath(path)
	if err != nil {
		return "", fmt.Errorf("无法从路径中提取Go版本: %v", err)
	}

	// 检查版本是否已存在
	_, err = s.versionRepo.FindByVersion(version)
	if err == nil {
		return "", fmt.Errorf("Go版本 %s 已存在", version)
	}

	// 创建新版本记录
	newVersion := &model.GoVersion{
		Version:  version,
		Path:     path,
		IsActive: false,
	}

	// 保存版本记录
	if err := s.versionRepo.Save(newVersion); err != nil {
		return "", fmt.Errorf("保存版本记录失败: %v", err)
	}

	return version, nil
}

// extractVersionFromPath 从Go安装路径中提取版本号
func (s *VersionService) extractVersionFromPath(path string) (string, error) {
	// 确定go可执行文件的路径
	var goExecPath string
	if runtime.GOOS == "windows" {
		goExecPath = filepath.Join(path, "bin", "go.exe")
	} else {
		goExecPath = filepath.Join(path, "bin", "go")
	}

	// 检查go可执行文件是否存在
	if _, err := os.Stat(goExecPath); os.IsNotExist(err) {
		return "", fmt.Errorf("在路径 %s 中未找到go可执行文件", path)
	}

	// 执行go version命令获取版本信息
	cmd := exec.Command(goExecPath, "version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("执行go version命令失败: %v", err)
	}

	// 解析输出获取版本号
	// go version命令的输出格式通常为: go version go1.21.0 windows/amd64
	versionStr := string(output)

	// 使用正则表达式提取版本号
	re := regexp.MustCompile(`go(\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(versionStr)
	if len(matches) < 2 {
		return "", fmt.Errorf("无法解析go version输出: %s", versionStr)
	}

	return matches[1], nil
}
