package main

import (
	"fmt"
	"os"
	"path/filepath"

	"version-list/internal/domain/service"
)

func main() {
	fmt.Println("=== Go Version Manager - 路径管理示例 ===\n")

	// 创建路径管理器
	pathManager := service.NewPathManager()
	pathValidator := service.NewPathValidator()

	// 1. 获取默认安装目录
	fmt.Println("1. 默认安装目录:")
	defaultDir := pathManager.GetDefaultInstallDir()
	fmt.Printf("   默认目录: %s\n\n", defaultDir)

	// 2. 验证路径
	fmt.Println("2. 路径验证:")
	testPaths := []string{
		defaultDir,
		filepath.Join(os.TempDir(), "go-versions"),
		"",              // 空路径
		"relative/path", // 相对路径
	}

	for _, path := range testPaths {
		if path == "" {
			fmt.Printf("   路径: (空)\n")
		} else {
			fmt.Printf("   路径: %s\n", path)
		}

		err := pathManager.ValidateInstallPath(path)
		if err != nil {
			fmt.Printf("   验证结果: ❌ %v\n", err)
		} else {
			fmt.Printf("   验证结果: ✅ 有效\n")
		}
		fmt.Println()
	}

	// 3. 创建安装目录
	fmt.Println("3. 创建安装目录:")
	testInstallDir := filepath.Join(os.TempDir(), "go-version-example")
	fmt.Printf("   目标目录: %s\n", testInstallDir)

	err := pathManager.CreateInstallDirectory(testInstallDir)
	if err != nil {
		fmt.Printf("   创建结果: ❌ %v\n", err)
	} else {
		fmt.Printf("   创建结果: ✅ 成功\n")

		// 清理
		defer func() {
			os.RemoveAll(testInstallDir)
			fmt.Printf("   清理完成: %s\n", testInstallDir)
		}()
	}
	fmt.Println()

	// 4. 版本目录管理
	fmt.Println("4. 版本目录管理:")
	version := "1.21.0"
	versionDir := pathManager.GetVersionDirectory(testInstallDir, version)
	fmt.Printf("   版本 %s 的目录: %s\n", version, versionDir)

	tempDir := pathManager.GetTempDirectory(version)
	fmt.Printf("   版本 %s 的临时目录: %s\n", version, tempDir)
	fmt.Println()

	// 5. 磁盘空间检查
	fmt.Println("5. 磁盘空间检查:")
	requiredSpace := int64(100 * 1024 * 1024) // 100MB
	fmt.Printf("   检查目录: %s\n", os.TempDir())
	fmt.Printf("   所需空间: %s\n", formatBytes(requiredSpace))

	err = pathManager.CheckDiskSpace(os.TempDir(), requiredSpace)
	if err != nil {
		fmt.Printf("   检查结果: ❌ %v\n", err)
	} else {
		fmt.Printf("   检查结果: ✅ 空间充足\n")

		// 获取可用空间
		availableSpace, err := pathManager.GetAvailableDiskSpace(os.TempDir())
		if err != nil {
			fmt.Printf("   可用空间: 无法获取 (%v)\n", err)
		} else {
			fmt.Printf("   可用空间: %s\n", formatBytes(availableSpace))
		}
	}
	fmt.Println()

	// 6. 目录状态检查
	fmt.Println("6. 目录状态检查:")
	if _, err := os.Stat(testInstallDir); err == nil {
		isEmpty, err := pathManager.IsDirectoryEmpty(testInstallDir)
		if err != nil {
			fmt.Printf("   检查是否为空: ❌ %v\n", err)
		} else {
			fmt.Printf("   目录是否为空: %t\n", isEmpty)
		}

		dirSize, err := pathManager.GetDirectorySize(testInstallDir)
		if err != nil {
			fmt.Printf("   目录大小: 无法获取 (%v)\n", err)
		} else {
			fmt.Printf("   目录大小: %s\n", formatBytes(dirSize))
		}
	}
	fmt.Println()

	// 7. 路径验证器高级功能
	fmt.Println("7. 路径验证器高级功能:")
	config := &service.InstallPathConfig{
		BaseDir:           testInstallDir,
		CreateIfNotExists: true,
		CheckPermissions:  true,
		CheckDiskSpace:    true,
		RequiredSpace:     1024 * 1024, // 1MB
	}

	validatedPath, err := pathValidator.ValidateAndPrepareInstallPath(config)
	if err != nil {
		fmt.Printf("   验证和准备: ❌ %v\n", err)
	} else {
		fmt.Printf("   验证和准备: ✅ 成功\n")
		fmt.Printf("   最终路径: %s\n", validatedPath)
	}
	fmt.Println()

	// 8. 路径信息详情
	fmt.Println("8. 路径信息详情:")
	pathInfo, err := pathValidator.GetPathInfo(testInstallDir)
	if err != nil {
		fmt.Printf("   获取路径信息失败: %v\n", err)
	} else {
		fmt.Printf("   路径信息:\n")
		fmt.Printf("%s\n", indent(pathInfo.String(), "   "))
	}

	fmt.Println("\n=== 示例完成 ===")
}

// formatBytes 格式化字节数
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// indent 为每行添加缩进
func indent(text, prefix string) string {
	lines := []string{}
	for _, line := range splitLines(text) {
		lines = append(lines, prefix+line)
	}
	return joinLines(lines)
}

// splitLines 分割字符串为行
func splitLines(text string) []string {
	lines := []string{}
	current := ""
	for _, r := range text {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

// joinLines 连接行为字符串
func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}
