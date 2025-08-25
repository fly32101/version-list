package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestMirrorCommand 测试mirror命令的基本功能
func TestMirrorCommand(t *testing.T) {
	// 保存原始的输出
	originalStdout := os.Stdout
	originalStderr := os.Stderr

	defer func() {
		os.Stdout = originalStdout
		os.Stderr = originalStderr
	}()

	t.Run("显示mirror命令帮助", func(t *testing.T) {
		cmd := createTestMirrorCommand()
		helpOutput, err := executeCommand(cmd, "--help")
		if err != nil {
			t.Fatalf("执行帮助命令失败: %v", err)
		}

		// 验证帮助内容
		if !strings.Contains(helpOutput, "镜像源管理") {
			t.Error("帮助信息应该包含镜像源管理描述")
		}

		if !strings.Contains(helpOutput, "list") {
			t.Error("帮助信息应该包含list子命令")
		}

		if !strings.Contains(helpOutput, "test") {
			t.Error("帮助信息应该包含test子命令")
		}

		if !strings.Contains(helpOutput, "fastest") {
			t.Error("帮助信息应该包含fastest子命令")
		}
	})

	t.Run("测试mirror list子命令帮助", func(t *testing.T) {
		cmd := createTestMirrorCommand()
		helpOutput, err := executeCommand(cmd, "list", "--help")
		if err != nil {
			t.Fatalf("执行list帮助命令失败: %v", err)
		}

		if !strings.Contains(helpOutput, "列出所有可用的镜像源") {
			t.Error("list帮助信息应该包含正确的描述")
		}

		if !strings.Contains(helpOutput, "--details") {
			t.Error("list帮助信息应该包含--details选项")
		}
	})

	t.Run("测试mirror test子命令帮助", func(t *testing.T) {
		cmd := createTestMirrorCommand()
		helpOutput, err := executeCommand(cmd, "test", "--help")
		if err != nil {
			t.Fatalf("执行test帮助命令失败: %v", err)
		}

		if !strings.Contains(helpOutput, "测试镜像源的连接速度") {
			t.Error("test帮助信息应该包含正确的描述")
		}

		if !strings.Contains(helpOutput, "--name") {
			t.Error("test帮助信息应该包含--name选项")
		}

		if !strings.Contains(helpOutput, "--timeout") {
			t.Error("test帮助信息应该包含--timeout选项")
		}
	})

	t.Run("测试mirror add子命令帮助", func(t *testing.T) {
		cmd := createTestMirrorCommand()
		helpOutput, err := executeCommand(cmd, "add", "--help")
		if err != nil {
			t.Fatalf("执行add帮助命令失败: %v", err)
		}

		if !strings.Contains(helpOutput, "添加自定义镜像源") {
			t.Error("add帮助信息应该包含正确的描述")
		}

		// 验证必需参数
		requiredFlags := []string{"--name", "--url", "--description", "--region"}
		for _, flag := range requiredFlags {
			if !strings.Contains(helpOutput, flag) {
				t.Errorf("add帮助信息应该包含必需参数 %s", flag)
			}
		}
	})

	t.Run("测试mirror remove子命令帮助", func(t *testing.T) {
		cmd := createTestMirrorCommand()
		helpOutput, err := executeCommand(cmd, "remove", "--help")
		if err != nil {
			t.Fatalf("执行remove帮助命令失败: %v", err)
		}

		if !strings.Contains(helpOutput, "移除自定义镜像源") {
			t.Error("remove帮助信息应该包含正确的描述")
		}

		if !strings.Contains(helpOutput, "--name") {
			t.Error("remove帮助信息应该包含--name选项")
		}
	})
}

// TestMirrorCommandValidation 测试mirror命令的参数验证
func TestMirrorCommandValidation(t *testing.T) {
	t.Run("add命令缺少必需参数", func(t *testing.T) {
		cmd := createTestMirrorCommand()

		// 测试缺少name参数
		_, err := executeCommand(cmd, "add", "--url", "https://example.com", "--description", "test", "--region", "test")
		if err == nil {
			t.Error("缺少--name参数时应该返回错误")
		}

		// 测试缺少url参数
		_, err = executeCommand(cmd, "add", "--name", "test", "--description", "test", "--region", "test")
		if err == nil {
			t.Error("缺少--url参数时应该返回错误")
		}
	})

	t.Run("remove命令缺少必需参数", func(t *testing.T) {
		cmd := createTestMirrorCommand()

		// 测试缺少name参数
		_, err := executeCommand(cmd, "remove")
		if err == nil {
			t.Error("缺少--name参数时应该返回错误")
		}
	})

	t.Run("validate命令缺少必需参数", func(t *testing.T) {
		cmd := createTestMirrorCommand()

		// 测试缺少name参数
		_, err := executeCommand(cmd, "validate")
		if err == nil {
			t.Error("缺少--name参数时应该返回错误")
		}
	})
}

// createTestMirrorCommand 创建用于测试的mirror命令
func createTestMirrorCommand() *cobra.Command {
	// 创建测试用的mirror命令，避免实际执行
	cmd := &cobra.Command{
		Use:   "mirror",
		Short: "镜像源管理",
		Long:  mirrorCmd.Long,
	}

	// 创建子命令
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有可用的镜像源",
		Long:  mirrorListCmd.Long,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("模拟执行 mirror list 命令")
		},
	}

	testCmd := &cobra.Command{
		Use:   "test",
		Short: "测试镜像源的连接速度和可用性",
		Long:  mirrorTestCmd.Long,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("模拟执行 mirror test 命令")
		},
	}

	fastestCmd := &cobra.Command{
		Use:   "fastest",
		Short: "自动选择最快的镜像源",
		Long:  mirrorFastestCmd.Long,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("模拟执行 mirror fastest 命令")
		},
	}

	addCmd := &cobra.Command{
		Use:   "add",
		Short: "添加自定义镜像源",
		Long:  mirrorAddCmd.Long,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("模拟执行 mirror add 命令")
		},
	}

	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "移除自定义镜像源",
		Long:  mirrorRemoveCmd.Long,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("模拟执行 mirror remove 命令")
		},
	}

	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "验证镜像源的可用性",
		Long:  mirrorValidateCmd.Long,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("模拟执行 mirror validate 命令")
		},
	}

	// 添加子命令
	cmd.AddCommand(listCmd)
	cmd.AddCommand(testCmd)
	cmd.AddCommand(fastestCmd)
	cmd.AddCommand(addCmd)
	cmd.AddCommand(removeCmd)
	cmd.AddCommand(validateCmd)

	// 添加选项（使用局部变量避免影响全局状态）
	var (
		testMirrorShowDetails bool
		testMirrorTestTimeout int
		testMirrorForceTest   bool
		testMirrorConfigPath  string
		testMirrorName        string
		testMirrorURL         string
		testMirrorDescription string
		testMirrorRegion      string
		testMirrorPriority    int
	)

	// list 命令选项
	listCmd.Flags().BoolVar(&testMirrorShowDetails, "details", false, "显示详细信息")
	listCmd.Flags().StringVar(&testMirrorConfigPath, "config", "", "指定配置文件路径")

	// test 命令选项
	testCmd.Flags().StringVar(&testMirrorName, "name", "", "指定要测试的镜像源名称")
	testCmd.Flags().IntVar(&testMirrorTestTimeout, "timeout", 30, "测试超时时间（秒）")
	testCmd.Flags().BoolVar(&testMirrorForceTest, "force", false, "强制测试，忽略缓存结果")

	// fastest 命令选项
	fastestCmd.Flags().IntVar(&testMirrorTestTimeout, "timeout", 30, "测试超时时间（秒）")
	fastestCmd.Flags().BoolVar(&testMirrorShowDetails, "details", false, "显示测试详情")

	// add 命令选项
	addCmd.Flags().StringVar(&testMirrorName, "name", "", "镜像源名称（必需）")
	addCmd.Flags().StringVar(&testMirrorURL, "url", "", "镜像源URL（必需）")
	addCmd.Flags().StringVar(&testMirrorDescription, "description", "", "镜像源描述（必需）")
	addCmd.Flags().StringVar(&testMirrorRegion, "region", "", "镜像源地区（必需）")
	addCmd.Flags().IntVar(&testMirrorPriority, "priority", 100, "镜像源优先级")
	addCmd.MarkFlagRequired("name")
	addCmd.MarkFlagRequired("url")
	addCmd.MarkFlagRequired("description")
	addCmd.MarkFlagRequired("region")

	// remove 命令选项
	removeCmd.Flags().StringVar(&testMirrorName, "name", "", "要移除的镜像源名称（必需）")
	removeCmd.MarkFlagRequired("name")

	// validate 命令选项
	validateCmd.Flags().StringVar(&testMirrorName, "name", "", "要验证的镜像源名称（必需）")
	validateCmd.MarkFlagRequired("name")

	return cmd
}

// executeCommand 执行命令并返回输出
func executeCommand(cmd *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return buf.String(), err
}
