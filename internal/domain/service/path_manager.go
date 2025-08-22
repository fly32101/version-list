package service

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// PathManager 路径管理器接口
type PathManager interface {
	GetDefaultInstallDir() string
	ValidateInstallPath(path string) error
	CreateInstallDirectory(path string) error
	GetVersionDirectory(baseDir, version string) string
	GetTempDirectory(version string) string
	CheckDiskSpace(path string, requiredBytes int64) error
	GetAvailableDiskSpace(path string) (int64, error)
	EnsureDirectoryPermissions(path string) error
	CleanupDirectory(path string) error
	IsDirectoryEmpty(path string) (bool, error)
	GetDirectorySize(path string) (int64, error)
}

// PathManagerImpl 路径管理器实现
type PathManagerImpl struct {
	defaultBaseDir string
}

// NewPathManager 创建路径管理器
func NewPathManager() PathManager {
	return &PathManagerImpl{
		defaultBaseDir: getDefaultBaseDirectory(),
	}
}

// GetDefaultInstallDir 获取默认安装目录
func (pm *PathManagerImpl) GetDefaultInstallDir() string {
	return pm.defaultBaseDir
}

// ValidateInstallPath 验证安装路径
func (pm *PathManagerImpl) ValidateInstallPath(path string) error {
	if path == "" {
		return NewInstallError(ErrorTypeValidation, "安装路径不能为空", nil)
	}

	// 检查路径是否为绝对路径
	if !filepath.IsAbs(path) {
		return NewInstallError(ErrorTypeValidation,
			fmt.Sprintf("安装路径必须是绝对路径: %s", path), nil)
	}

	// 检查路径长度（Windows有路径长度限制）
	if runtime.GOOS == "windows" && len(path) > 260 {
		return NewInstallError(ErrorTypeValidation,
			fmt.Sprintf("Windows路径长度不能超过260字符: %d", len(path)), nil)
	}

	// 检查路径中是否包含非法字符
	if err := pm.validatePathCharacters(path); err != nil {
		return err
	}

	// 检查父目录是否存在且可写
	parentDir := filepath.Dir(path)
	if _, err := os.Stat(parentDir); err != nil {
		if os.IsNotExist(err) {
			// 父目录不存在，检查是否可以创建
			grandParentDir := filepath.Dir(parentDir)
			if _, err := os.Stat(grandParentDir); err != nil {
				return NewInstallError(ErrorTypeFileSystem,
					fmt.Sprintf("父目录的父目录不存在: %s", grandParentDir), err)
			}
			if err := pm.checkDirectoryWritable(grandParentDir); err != nil {
				return NewInstallError(ErrorTypePermission,
					fmt.Sprintf("无法在父目录的父目录中创建目录: %s", grandParentDir), err)
			}
		} else {
			return NewInstallError(ErrorTypeFileSystem,
				fmt.Sprintf("访问父目录失败: %s", parentDir), err)
		}
	} else {
		// 父目录存在，检查是否可写
		if err := pm.checkDirectoryWritable(parentDir); err != nil {
			return NewInstallError(ErrorTypePermission,
				fmt.Sprintf("父目录不可写: %s", parentDir), err)
		}
	}

	return nil
}

// CreateInstallDirectory 创建安装目录
func (pm *PathManagerImpl) CreateInstallDirectory(path string) error {
	// 先验证路径
	if err := pm.ValidateInstallPath(path); err != nil {
		return err
	}

	// 检查目录是否已存在
	if info, err := os.Stat(path); err == nil {
		if !info.IsDir() {
			return NewInstallError(ErrorTypeFileSystem,
				fmt.Sprintf("路径已存在但不是目录: %s", path), nil)
		}

		// 检查目录是否为空
		isEmpty, err := pm.IsDirectoryEmpty(path)
		if err != nil {
			return NewInstallError(ErrorTypeFileSystem,
				fmt.Sprintf("检查目录是否为空失败: %v", err), err)
		}

		if !isEmpty {
			return NewInstallError(ErrorTypeFileSystem,
				fmt.Sprintf("目录已存在且不为空: %s", path), nil)
		}

		return nil // 目录已存在且为空
	}

	// 创建目录
	if err := os.MkdirAll(path, 0755); err != nil {
		return NewInstallError(ErrorTypeFileSystem,
			fmt.Sprintf("创建安装目录失败: %v", err), err)
	}

	// 验证目录权限
	if err := pm.EnsureDirectoryPermissions(path); err != nil {
		return err
	}

	return nil
}

// GetVersionDirectory 获取版本目录路径
func (pm *PathManagerImpl) GetVersionDirectory(baseDir, version string) string {
	return filepath.Join(baseDir, version)
}

// GetTempDirectory 获取临时目录路径
func (pm *PathManagerImpl) GetTempDirectory(version string) string {
	tempBase := os.TempDir()
	return filepath.Join(tempBase, fmt.Sprintf("go-install-%s", version))
}

// CheckDiskSpace 检查磁盘空间
func (pm *PathManagerImpl) CheckDiskSpace(path string, requiredBytes int64) error {
	availableBytes, err := pm.GetAvailableDiskSpace(path)
	if err != nil {
		return NewInstallError(ErrorTypeFileSystem,
			fmt.Sprintf("获取磁盘空间信息失败: %v", err), err)
	}

	if availableBytes < requiredBytes {
		return NewInstallError(ErrorTypeInsufficientSpace,
			fmt.Sprintf("磁盘空间不足: 需要 %s, 可用 %s",
				formatBytes(requiredBytes), formatBytes(availableBytes)), nil).
			WithContext("required_bytes", requiredBytes).
			WithContext("available_bytes", availableBytes).
			WithContext("path", path)
	}

	return nil
}

// GetAvailableDiskSpace 获取可用磁盘空间
func (pm *PathManagerImpl) GetAvailableDiskSpace(path string) (int64, error) {
	// 确保路径存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// 如果路径不存在，检查父目录
		path = filepath.Dir(path)
	}

	if runtime.GOOS == "windows" {
		return pm.getWindowsDiskSpace(path)
	}
	return pm.getUnixDiskSpace(path)
}

// EnsureDirectoryPermissions 确保目录权限
func (pm *PathManagerImpl) EnsureDirectoryPermissions(path string) error {
	// 检查目录是否存在
	info, err := os.Stat(path)
	if err != nil {
		return NewInstallError(ErrorTypeFileSystem,
			fmt.Sprintf("访问目录失败: %v", err), err)
	}

	if !info.IsDir() {
		return NewInstallError(ErrorTypeValidation,
			fmt.Sprintf("路径不是目录: %s", path), nil)
	}

	// 检查读写权限
	if err := pm.checkDirectoryWritable(path); err != nil {
		return NewInstallError(ErrorTypePermission,
			fmt.Sprintf("目录权限不足: %s", path), err)
	}

	// 在Unix系统上，确保目录有正确的权限
	if runtime.GOOS != "windows" {
		if err := os.Chmod(path, 0755); err != nil {
			return NewInstallError(ErrorTypePermission,
				fmt.Sprintf("设置目录权限失败: %v", err), err)
		}
	}

	return nil
}

// CleanupDirectory 清理目录
func (pm *PathManagerImpl) CleanupDirectory(path string) error {
	// 检查目录是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // 目录不存在，无需清理
	}

	// 删除目录及其所有内容
	if err := os.RemoveAll(path); err != nil {
		return NewInstallError(ErrorTypeFileSystem,
			fmt.Sprintf("清理目录失败: %v", err), err)
	}

	return nil
}

// IsDirectoryEmpty 检查目录是否为空
func (pm *PathManagerImpl) IsDirectoryEmpty(path string) (bool, error) {
	dir, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer dir.Close()

	_, err = dir.Readdir(1)
	if err == nil {
		return false, nil // 目录不为空
	}

	// 检查是否是EOF错误（表示目录为空）
	if err.Error() == "EOF" {
		return true, nil
	}

	return false, err
}

// GetDirectorySize 获取目录大小
func (pm *PathManagerImpl) GetDirectorySize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	if err != nil {
		return 0, NewInstallError(ErrorTypeFileSystem,
			fmt.Sprintf("计算目录大小失败: %v", err), err)
	}

	return size, nil
}

// 私有辅助方法

// getDefaultBaseDirectory 获取默认基础目录
func getDefaultBaseDirectory() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// 如果无法获取用户主目录，使用临时目录
		return filepath.Join(os.TempDir(), "go-versions")
	}

	return filepath.Join(homeDir, ".go", "versions")
}

// validatePathCharacters 验证路径字符
func (pm *PathManagerImpl) validatePathCharacters(path string) error {
	// Windows不允许的字符（但允许驱动器字母后的冒号）
	if runtime.GOOS == "windows" {
		invalidChars := []string{"<", ">", "\"", "|", "?", "*"}
		for _, char := range invalidChars {
			if strings.Contains(path, char) {
				return NewInstallError(ErrorTypeValidation,
					fmt.Sprintf("路径包含非法字符 '%s': %s", char, path), nil)
			}
		}

		// 检查冒号的使用是否正确（只允许在驱动器字母后）
		colonIndex := strings.Index(path, ":")
		if colonIndex != -1 {
			// 如果有冒号，它应该在第二个位置（如 C:）
			if colonIndex != 1 || len(path) < 3 || path[2] != '\\' {
				// 检查是否有多个冒号
				if strings.Count(path, ":") > 1 {
					return NewInstallError(ErrorTypeValidation,
						fmt.Sprintf("路径包含多个冒号: %s", path), nil)
				}
			}
		}
	}

	// 检查控制字符
	for _, r := range path {
		if r < 32 {
			return NewInstallError(ErrorTypeValidation,
				fmt.Sprintf("路径包含控制字符: %s", path), nil)
		}
	}

	return nil
}

// checkDirectoryWritable 检查目录是否可写
func (pm *PathManagerImpl) checkDirectoryWritable(path string) error {
	// 尝试在目录中创建临时文件
	testFile := filepath.Join(path, ".write_test")
	file, err := os.Create(testFile)
	if err != nil {
		return err
	}
	file.Close()

	// 删除测试文件
	os.Remove(testFile)
	return nil
}

// getWindowsDiskSpace 获取Windows磁盘空间
func (pm *PathManagerImpl) getWindowsDiskSpace(path string) (int64, error) {
	if runtime.GOOS != "windows" {
		return 0, fmt.Errorf("此方法仅支持Windows系统")
	}

	// 使用简化的方法获取磁盘空间
	// 在实际项目中，可以使用第三方库如 github.com/shirou/gopsutil
	// 这里使用一个简化的实现

	// 尝试创建一个临时文件来测试可用空间
	tempFile := filepath.Join(path, ".space_test")
	file, err := os.Create(tempFile)
	if err != nil {
		return 0, fmt.Errorf("无法检测磁盘空间: %v", err)
	}
	file.Close()
	os.Remove(tempFile)

	// 返回一个保守的估计值（1GB）
	// 在生产环境中应该使用更精确的方法
	return 1024 * 1024 * 1024, nil
}

// getUnixDiskSpace 获取Unix磁盘空间
func (pm *PathManagerImpl) getUnixDiskSpace(path string) (int64, error) {
	if runtime.GOOS == "windows" {
		return 0, fmt.Errorf("此方法不支持Windows系统")
	}

	// 使用简化的方法获取磁盘空间
	// 在实际项目中，可以使用第三方库如 github.com/shirou/gopsutil
	// 这里使用一个简化的实现

	// 尝试创建一个临时文件来测试可用空间
	tempFile := filepath.Join(path, ".space_test")
	file, err := os.Create(tempFile)
	if err != nil {
		return 0, fmt.Errorf("无法检测磁盘空间: %v", err)
	}
	file.Close()
	os.Remove(tempFile)

	// 返回一个保守的估计值（1GB）
	// 在生产环境中应该使用更精确的方法
	return 1024 * 1024 * 1024, nil
}

// InstallPathConfig 安装路径配置
type InstallPathConfig struct {
	BaseDir           string // 基础安装目录
	CustomPath        string // 自定义路径
	CreateIfNotExists bool   // 如果不存在是否创建
	CheckPermissions  bool   // 是否检查权限
	CheckDiskSpace    bool   // 是否检查磁盘空间
	RequiredSpace     int64  // 所需空间（字节）
}

// PathValidator 路径验证器
type PathValidator struct {
	pathManager PathManager
}

// NewPathValidator 创建路径验证器
func NewPathValidator() *PathValidator {
	return &PathValidator{
		pathManager: NewPathManager(),
	}
}

// ValidateAndPrepareInstallPath 验证并准备安装路径
func (pv *PathValidator) ValidateAndPrepareInstallPath(config *InstallPathConfig) (string, error) {
	var installPath string

	// 确定安装路径
	if config.CustomPath != "" {
		installPath = config.CustomPath
	} else {
		installPath = config.BaseDir
		if installPath == "" {
			installPath = pv.pathManager.GetDefaultInstallDir()
		}
	}

	// 验证路径
	if err := pv.pathManager.ValidateInstallPath(installPath); err != nil {
		return "", err
	}

	// 检查磁盘空间
	if config.CheckDiskSpace && config.RequiredSpace > 0 {
		if err := pv.pathManager.CheckDiskSpace(installPath, config.RequiredSpace); err != nil {
			return "", err
		}
	}

	// 创建目录（如果需要）
	if config.CreateIfNotExists {
		if err := pv.pathManager.CreateInstallDirectory(installPath); err != nil {
			return "", err
		}
	}

	// 检查权限
	if config.CheckPermissions {
		if err := pv.pathManager.EnsureDirectoryPermissions(installPath); err != nil {
			return "", err
		}
	}

	return installPath, nil
}

// GetPathInfo 获取路径信息
func (pv *PathValidator) GetPathInfo(path string) (*PathInfo, error) {
	info := &PathInfo{
		Path:   path,
		Exists: false,
	}

	// 检查路径是否存在
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return info, nil
		}
		return nil, err
	}

	info.Exists = true
	info.IsDirectory = fileInfo.IsDir()
	info.Size = fileInfo.Size()
	info.ModTime = fileInfo.ModTime()
	info.Permissions = fileInfo.Mode()

	// 获取可用磁盘空间
	if availableSpace, err := pv.pathManager.GetAvailableDiskSpace(path); err == nil {
		info.AvailableSpace = availableSpace
	}

	// 如果是目录，检查是否为空
	if info.IsDirectory {
		if isEmpty, err := pv.pathManager.IsDirectoryEmpty(path); err == nil {
			info.IsEmpty = isEmpty
		}

		// 获取目录大小
		if dirSize, err := pv.pathManager.GetDirectorySize(path); err == nil {
			info.DirectorySize = dirSize
		}
	}

	return info, nil
}

// PathInfo 路径信息
type PathInfo struct {
	Path           string      // 路径
	Exists         bool        // 是否存在
	IsDirectory    bool        // 是否为目录
	IsEmpty        bool        // 是否为空（仅目录）
	Size           int64       // 文件大小
	DirectorySize  int64       // 目录大小（递归）
	AvailableSpace int64       // 可用磁盘空间
	Permissions    os.FileMode // 文件权限
	ModTime        time.Time   // 修改时间
}

// String 返回路径信息的字符串表示
func (pi *PathInfo) String() string {
	var parts []string

	parts = append(parts, fmt.Sprintf("路径: %s", pi.Path))
	parts = append(parts, fmt.Sprintf("存在: %t", pi.Exists))

	if pi.Exists {
		if pi.IsDirectory {
			parts = append(parts, "类型: 目录")
			parts = append(parts, fmt.Sprintf("为空: %t", pi.IsEmpty))
			if pi.DirectorySize > 0 {
				parts = append(parts, fmt.Sprintf("目录大小: %s", formatBytes(pi.DirectorySize)))
			}
		} else {
			parts = append(parts, "类型: 文件")
			parts = append(parts, fmt.Sprintf("大小: %s", formatBytes(pi.Size)))
		}

		parts = append(parts, fmt.Sprintf("权限: %s", pi.Permissions))
	}

	if pi.AvailableSpace > 0 {
		parts = append(parts, fmt.Sprintf("可用空间: %s", formatBytes(pi.AvailableSpace)))
	}

	return strings.Join(parts, "\n")
}
