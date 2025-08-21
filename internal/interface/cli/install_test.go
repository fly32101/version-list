package cli

import (
	"strings"
	"testing"
)

func TestIsValidVersion(t *testing.T) {
	testCases := []struct {
		version string
		valid   bool
	}{
		{"1.21.0", true},
		{"1.20.5", true},
		{"1.19.10", true},
		{"1.21.0-rc1", true},
		{"1.21.0-beta1", true},
		{"", false},
		{"invalid", true},  // 只包含字母，应该有效
		{"1.21.0@", false}, // 包含无效字符
		{"1.21.0#", false}, // 包含无效字符
		{"1.21.0$", false}, // 包含无效字符
	}

	for _, tc := range testCases {
		t.Run(tc.version, func(t *testing.T) {
			result := isValidVersion(tc.version)
			if result != tc.valid {
				t.Errorf("isValidVersion(%s) = %v, 期望 %v", tc.version, result, tc.valid)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	testCases := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{163553657, "156.0 MB"}, // 实际Go安装包大小
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := formatBytes(tc.bytes)
			if result != tc.expected {
				t.Errorf("formatBytes(%d) = %s, 期望 %s", tc.bytes, result, tc.expected)
			}
		})
	}
}

func TestInstallCommandFlags(t *testing.T) {
	// 测试命令行选项是否正确定义
	cmd := installCmd

	// 检查命令基本信息
	if cmd.Use != "install [version]" {
		t.Errorf("命令使用格式不正确: %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("命令简短描述不能为空")
	}

	if cmd.Long == "" {
		t.Error("命令详细描述不能为空")
	}

	// 检查必需参数数量
	if cmd.Args == nil {
		t.Error("命令应该定义参数验证")
	}

	// 检查是否定义了预期的选项
	expectedFlags := []string{
		"path",
		"force",
		"skip-verification",
		"timeout",
		"max-retries",
		"no-progress",
		"online",
	}

	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("缺少预期的选项: --%s", flagName)
		}
	}

	// 检查选项类型
	pathFlag := cmd.Flags().Lookup("path")
	if pathFlag != nil && pathFlag.Value.Type() != "string" {
		t.Error("--path 选项应该是字符串类型")
	}

	forceFlag := cmd.Flags().Lookup("force")
	if forceFlag != nil && forceFlag.Value.Type() != "bool" {
		t.Error("--force 选项应该是布尔类型")
	}

	timeoutFlag := cmd.Flags().Lookup("timeout")
	if timeoutFlag != nil && timeoutFlag.Value.Type() != "int" {
		t.Error("--timeout 选项应该是整数类型")
	}
}

func TestInstallCommandHelp(t *testing.T) {
	// 测试帮助文本内容
	helpText := installCmd.Long

	// 检查是否包含关键信息
	expectedContent := []string{
		"在线安装",
		"本地安装",
		"示例",
		"go-version install",
		"--path",
		"--force",
		"--no-progress",
	}

	for _, content := range expectedContent {
		if !strings.Contains(helpText, content) {
			t.Errorf("帮助文本应该包含: %s", content)
		}
	}
}

func TestInstallCommandExamples(t *testing.T) {
	// 验证帮助文本中的示例是否正确
	helpText := installCmd.Long

	// 检查示例命令格式
	examples := []string{
		"go-version install 1.21.0",
		"go-version install 1.21.0 --path /custom",
		"go-version install 1.21.0 --force",
		"go-version install 1.21.0 --no-progress",
	}

	for _, example := range examples {
		if !strings.Contains(helpText, example) {
			t.Errorf("帮助文本应该包含示例: %s", example)
		}
	}
}

// 测试命令行选项的默认值
func TestInstallCommandDefaults(t *testing.T) {
	// 重置全局变量到默认值
	installPath = ""
	forceInstall = false
	skipVerification = false
	installTimeout = 300
	maxRetries = 3
	noProgress = false
	onlineInstall = true

	// 验证默认值
	if installPath != "" {
		t.Errorf("installPath 默认值应该为空字符串, 得到: %s", installPath)
	}

	if forceInstall != false {
		t.Errorf("forceInstall 默认值应该为 false, 得到: %v", forceInstall)
	}

	if skipVerification != false {
		t.Errorf("skipVerification 默认值应该为 false, 得到: %v", skipVerification)
	}

	if installTimeout != 300 {
		t.Errorf("installTimeout 默认值应该为 300, 得到: %d", installTimeout)
	}

	if maxRetries != 3 {
		t.Errorf("maxRetries 默认值应该为 3, 得到: %d", maxRetries)
	}

	if noProgress != false {
		t.Errorf("noProgress 默认值应该为 false, 得到: %v", noProgress)
	}

	if onlineInstall != true {
		t.Errorf("onlineInstall 默认值应该为 true, 得到: %v", onlineInstall)
	}
}

// 测试版本号验证的边界情况
func TestVersionValidationEdgeCases(t *testing.T) {
	edgeCases := []struct {
		version     string
		valid       bool
		description string
	}{
		{"1", true, "单个数字"},
		{"1.0", true, "两段版本号"},
		{"1.0.0", true, "三段版本号"},
		{"1.0.0.0", true, "四段版本号"},
		{"1.21.0-rc.1", true, "带RC版本"},
		{"1.21.0-beta.1", true, "带Beta版本"},
		{"1.21.0-alpha", true, "带Alpha版本"},
		{"go1.21.0", true, "带go前缀"},
		{"v1.21.0", true, "带v前缀"},
		{"1.21.0+build.1", false, "带构建号（包含+）"},
		{"1.21.0 ", false, "带空格"},
		{" 1.21.0", false, "前导空格"},
		{"1..21.0", true, "双点（应该被允许）"},
	}

	for _, tc := range edgeCases {
		t.Run(tc.description, func(t *testing.T) {
			result := isValidVersion(tc.version)
			if result != tc.valid {
				t.Errorf("版本 '%s' (%s): isValidVersion() = %v, 期望 %v",
					tc.version, tc.description, result, tc.valid)
			}
		})
	}
}

// 测试字节格式化的边界情况
func TestFormatBytesEdgeCases(t *testing.T) {
	edgeCases := []struct {
		bytes       int64
		expected    string
		description string
	}{
		{0, "0 B", "零字节"},
		{1, "1 B", "1字节"},
		{1023, "1023 B", "1KB以下最大值"},
		{1025, "1.0 KB", "刚超过1KB"},
		{1048575, "1024.0 KB", "1MB以下最大值"},
		{1048577, "1.0 MB", "刚超过1MB"},
		{-1, "-1 B", "负数"},
		{9223372036854775807, "8.0 EB", "最大int64值"},
	}

	for _, tc := range edgeCases {
		t.Run(tc.description, func(t *testing.T) {
			result := formatBytes(tc.bytes)
			if result != tc.expected {
				t.Errorf("formatBytes(%d) = %s, 期望 %s (%s)",
					tc.bytes, result, tc.expected, tc.description)
			}
		})
	}
}
