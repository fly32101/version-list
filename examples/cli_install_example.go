package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("=== CLI Install命令示例 ===")

	// 模拟命令行参数
	testCases := []struct {
		name string
		args []string
	}{
		{
			name: "显示帮助信息",
			args: []string{"install", "--help"},
		},
		{
			name: "基本安装命令格式",
			args: []string{"install", "1.21.0"},
		},
		{
			name: "带自定义路径的安装",
			args: []string{"install", "1.21.0", "--path", "/custom/path"},
		},
		{
			name: "强制重新安装",
			args: []string{"install", "1.21.0", "--force"},
		},
		{
			name: "跳过验证的安装",
			args: []string{"install", "1.21.0", "--skip-verification"},
		},
		{
			name: "不显示进度的安装",
			args: []string{"install", "1.21.0", "--no-progress"},
		},
		{
			name: "自定义超时和重试",
			args: []string{"install", "1.21.0", "--timeout", "600", "--max-retries", "5"},
		},
	}

	for _, tc := range testCases {
		fmt.Printf("\n--- %s ---\n", tc.name)
		fmt.Printf("命令: go-version %s\n", joinArgs(tc.args))

		if tc.name == "显示帮助信息" {
			// 对于帮助命令，我们可以实际执行
			fmt.Println("执行帮助命令...")
			// 这里可以调用实际的CLI命令
		} else {
			// 对于其他命令，只显示格式
			fmt.Println("（示例命令格式，实际执行需要网络连接）")
		}
	}

	// 演示版本验证
	fmt.Println("\n=== 版本号验证示例 ===")
	testVersions := []string{
		"1.21.0",     // 有效
		"1.20.5",     // 有效
		"1.21.0-rc1", // 有效
		"invalid@",   // 无效
		"",           // 无效
	}

	for _, version := range testVersions {
		valid := isValidVersion(version)
		status := "✅ 有效"
		if !valid {
			status = "❌ 无效"
		}
		fmt.Printf("版本 '%s': %s\n", version, status)
	}

	// 演示字节格式化
	fmt.Println("\n=== 文件大小格式化示例 ===")
	sizes := []int64{
		1024,       // 1 KB
		1048576,    // 1 MB
		163553657,  // Go安装包大小
		1073741824, // 1 GB
	}

	for _, size := range sizes {
		formatted := formatBytes(size)
		fmt.Printf("%d 字节 = %s\n", size, formatted)
	}

	fmt.Println("\n=== 命令行选项说明 ===")
	options := []struct {
		flag        string
		description string
		example     string
	}{
		{"--path", "指定自定义安装路径", "--path /usr/local/go"},
		{"--force", "强制重新安装已存在的版本", "--force"},
		{"--skip-verification", "跳过文件完整性验证", "--skip-verification"},
		{"--timeout", "设置安装超时时间（秒）", "--timeout 600"},
		{"--max-retries", "设置最大重试次数", "--max-retries 5"},
		{"--no-progress", "不显示进度条", "--no-progress"},
		{"--online", "在线安装模式（默认）", "--online"},
	}

	for _, opt := range options {
		fmt.Printf("%-20s %s\n", opt.flag, opt.description)
		fmt.Printf("%-20s 示例: %s\n", "", opt.example)
		fmt.Println()
	}

	fmt.Println("=== 示例完成 ===")
}

// joinArgs 连接命令行参数
func joinArgs(args []string) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += " "
		}
		result += arg
	}
	return result
}

// isValidVersion 版本验证函数（从CLI包复制）
func isValidVersion(version string) bool {
	if len(version) == 0 {
		return false
	}

	for _, char := range version {
		if !((char >= '0' && char <= '9') || char == '.' || char == '-' || (char >= 'a' && char <= 'z')) {
			return false
		}
	}

	return true
}

// formatBytes 字节格式化函数（从CLI包复制）
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

func init() {
	// 设置示例环境
	if len(os.Args) > 1 && os.Args[1] == "demo" {
		// 演示模式，显示更多信息
		fmt.Println("演示模式已启用")
	}
}
