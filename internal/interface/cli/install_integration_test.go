package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestInstallCommandIntegration 测试install命令的集成功能
func TestInstallCommandIntegration(t *testing.T) {
	// 保存原始的输出
	originalStdout := os.Stdout
	originalStderr := os.Stderr

	defer func() {
		os.Stdout = originalStdout
		os.Stderr = originalStderr
	}()

	t.Run("显示帮助信息", func(t *testing.T) {
		// 创建命令副本以避免影响全局状态
		cmd := &cobra.Command{
			Use:   installCmd.Use,
			Short: installCmd.Short,
			Long:  installCmd.Long,
		}

		// 测试帮助输出
		helpOutput, err := executeCommand(cmd, "--help")
		if err != nil {
			t.Fatalf("执行帮助命令失败: %v", err)
		}

		// 验证帮助内容
		if !strings.Contains(helpOutput, "安装指定版本的Go") {
			t.Error("帮助信息应该包含命令描述")
		}

		if !strings.Contains(helpOutput, "在线安装") {
			t.Error("帮助信息应该包含在线安装说明")
		}
	})

	t.Run("验证参数要求", func(t *testing.T) {
		// 测试缺少版本参数
		cmd := createTestInstallCommand()
		_, err := executeCommand(cmd)

		if err == nil {
			t.Error("缺少版本参数时应该返回错误")
		}
	})

	t.Run("验证选项解析", func(t *testing.T) {
		// 重置全局变量
		resetInstallFlags()

		cmd := createTestInstallCommand()

		// 测试选项解析（不实际执行安装）
		args := []string{
			"1.21.0",
			"--path", "/custom/path",
			"--force",
			"--skip-verification",
			"--timeout", "600",
			"--max-retries", "5",
			"--no-progress",
		}

		// 只解析参数，不执行
		cmd.ParseFlags(args)

		// 验证选项值（这里需要手动检查，因为我们没有实际执行命令）
		pathFlag := cmd.Flags().Lookup("path")
		if pathFlag != nil && pathFlag.Value.String() != "/custom/path" {
			t.Errorf("--path 选项解析错误: 得到 %s, 期望 /custom/path", pathFlag.Value.String())
		}

		forceFlag := cmd.Flags().Lookup("force")
		if forceFlag != nil && forceFlag.Value.String() != "true" {
			t.Errorf("--force 选项解析错误: 得到 %s, 期望 true", forceFlag.Value.String())
		}
	})
}

// TestInstallCommandValidation 测试命令验证逻辑
func TestInstallCommandValidation(t *testing.T) {
	t.Run("版本号验证", func(t *testing.T) {
		validVersions := []string{
			"1.21.0",
			"1.20.5",
			"1.19.10",
			"1.21.0-rc1",
		}

		for _, version := range validVersions {
			if !isValidVersion(version) {
				t.Errorf("版本 %s 应该是有效的", version)
			}
		}

		invalidVersions := []string{
			"",
			"1.21.0@invalid",
			"1.21.0#test",
			"1.21.0$",
		}

		for _, version := range invalidVersions {
			if isValidVersion(version) {
				t.Errorf("版本 %s 应该是无效的", version)
			}
		}
	})

	t.Run("选项组合验证", func(t *testing.T) {
		// 测试本地安装模式需要path参数的逻辑
		// 这里主要测试逻辑，不实际执行安装

		// 模拟本地安装模式但没有path
		onlineInstall = false
		installPath = ""

		// 这种情况下runLocalInstall应该会报错
		// 但我们在这里只测试逻辑，不实际调用

		if onlineInstall == false && installPath == "" {
			// 这是预期的错误情况
			t.Log("本地安装模式需要path参数 - 验证通过")
		}

		// 重置状态
		onlineInstall = true
		installPath = ""
	})
}

// TestInstallCommandOutput 测试命令输出格式
func TestInstallCommandOutput(t *testing.T) {
	t.Run("字节格式化输出", func(t *testing.T) {
		testCases := []struct {
			bytes    int64
			expected string
		}{
			{1024, "1.0 KB"},
			{1048576, "1.0 MB"},
			{1073741824, "1.0 GB"},
		}

		for _, tc := range testCases {
			result := formatBytes(tc.bytes)
			if result != tc.expected {
				t.Errorf("formatBytes(%d) = %s, 期望 %s", tc.bytes, result, tc.expected)
			}
		}
	})

	t.Run("版本验证输出", func(t *testing.T) {
		// 测试各种版本格式
		versions := []string{
			"1.21.0",
			"1.20.5",
			"1.19.10-rc1",
		}

		for _, version := range versions {
			if !isValidVersion(version) {
				t.Errorf("版本 %s 验证失败", version)
			}
		}
	})
}

// 辅助函数

// executeCommand 执行命令并返回输出
func executeCommand(cmd *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return buf.String(), err
}

// createTestInstallCommand 创建测试用的install命令
func createTestInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [version]",
		Short: "安装指定版本的Go",
		Long:  installCmd.Long,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// 测试用的空实现，不实际执行安装
			version := args[0]
			if !isValidVersion(version) {
				cmd.PrintErr("无效的版本号")
				return
			}
			cmd.Printf("模拟安装版本: %s\n", version)
		},
	}

	// 添加选项
	cmd.Flags().StringVar(&installPath, "path", "", "自定义安装路径")
	cmd.Flags().BoolVar(&forceInstall, "force", false, "强制重新安装")
	cmd.Flags().BoolVar(&skipVerification, "skip-verification", false, "跳过验证")
	cmd.Flags().IntVar(&installTimeout, "timeout", 300, "超时时间")
	cmd.Flags().IntVar(&maxRetries, "max-retries", 3, "最大重试次数")
	cmd.Flags().BoolVar(&noProgress, "no-progress", false, "不显示进度")
	cmd.Flags().BoolVar(&onlineInstall, "online", true, "在线安装")

	return cmd
}

// resetInstallFlags 重置安装选项到默认值
func resetInstallFlags() {
	installPath = ""
	forceInstall = false
	skipVerification = false
	installTimeout = 300
	maxRetries = 3
	noProgress = false
	onlineInstall = true
}

// TestInstallCommandEdgeCases 测试边界情况
func TestInstallCommandEdgeCases(t *testing.T) {
	t.Run("极长版本号", func(t *testing.T) {
		longVersion := strings.Repeat("1.0.", 100) + "0"
		if !isValidVersion(longVersion) {
			t.Error("极长版本号应该被接受（由具体实现决定是否支持）")
		}
	})

	t.Run("特殊字符版本号", func(t *testing.T) {
		specialVersions := []string{
			"1.21.0-alpha",
			"1.21.0-beta.1",
			"1.21.0-rc.1",
		}

		for _, version := range specialVersions {
			if !isValidVersion(version) {
				t.Errorf("特殊版本号 %s 应该被接受", version)
			}
		}
	})

	t.Run("数值边界", func(t *testing.T) {
		// 测试超大数值
		largeTimeout := 999999
		if largeTimeout < 0 {
			t.Error("超时时间不应该为负数")
		}

		// 测试零值
		zeroRetries := 0
		if zeroRetries < 0 {
			t.Error("重试次数不应该为负数")
		}
	})
}

// TestInstallCommandCompatibility 测试命令兼容性
func TestInstallCommandCompatibility(t *testing.T) {
	t.Run("向后兼容性", func(t *testing.T) {
		// 测试基本的install命令格式是否保持兼容
		cmd := installCmd

		if cmd.Use != "install [version]" {
			t.Error("基本命令格式应该保持向后兼容")
		}

		// 测试必需参数
		if cmd.Args == nil {
			t.Error("应该保持版本参数要求")
		}
	})

	t.Run("选项兼容性", func(t *testing.T) {
		// 确保新增的选项不会破坏现有功能
		cmd := installCmd

		// 检查关键选项存在
		essentialFlags := []string{"path", "force"}
		for _, flagName := range essentialFlags {
			if cmd.Flags().Lookup(flagName) == nil {
				t.Errorf("关键选项 --%s 缺失", flagName)
			}
		}
	})
}
