package cli

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"version-list/internal/domain/service"

	"github.com/spf13/cobra"
)

// 镜像命令选项变量
var (
	mirrorTestTimeout int
	mirrorShowDetails bool
	mirrorForceTest   bool
	mirrorConfigPath  string
	// mirrorName 已在 install.go 中声明，此处重用
	mirrorURL         string
	mirrorDescription string
	mirrorRegion      string
	mirrorPriority    int
)

var mirrorCmd = &cobra.Command{
	Use:   "mirror",
	Short: "镜像源管理",
	Long: `管理Go下载镜像源。

支持的功能：
- 列出所有可用的镜像源
- 测试镜像源的连接速度和可用性
- 自动选择最快的镜像源
- 添加和管理自定义镜像源
- 验证镜像源的可用性

示例：
  go-version mirror list                    # 列出所有镜像源
  go-version mirror test                    # 测试所有镜像源速度
  go-version mirror test --name goproxy-cn # 测试指定镜像源
  go-version mirror fastest                # 选择最快的镜像源
  go-version mirror add                     # 添加自定义镜像源
  go-version mirror remove --name custom   # 移除自定义镜像源`,
}

var mirrorListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有可用的镜像源",
	Long: `列出所有可用的镜像源，包括预设镜像和自定义镜像。

选项：
  --details    显示详细信息
  --config     指定配置文件路径`,
	Run: runMirrorListCommand,
}

var mirrorTestCmd = &cobra.Command{
	Use:   "test",
	Short: "测试镜像源的连接速度和可用性",
	Long: `测试镜像源的连接速度和可用性。

可以测试所有镜像源或指定的镜像源。

选项：
  --name       指定要测试的镜像源名称
  --timeout    测试超时时间（秒）
  --force      强制测试，忽略缓存结果`,
	Run: runMirrorTestCommand,
}

var mirrorFastestCmd = &cobra.Command{
	Use:   "fastest",
	Short: "自动选择最快的镜像源",
	Long: `自动测试所有镜像源并选择响应最快的一个。

选项：
  --timeout    测试超时时间（秒）
  --details    显示测试详情`,
	Run: runMirrorFastestCommand,
}

var mirrorAddCmd = &cobra.Command{
	Use:   "add",
	Short: "添加自定义镜像源",
	Long: `添加自定义镜像源到配置中。

必需参数：
  --name        镜像源名称（唯一标识）
  --url         镜像源URL
  --description 镜像源描述
  --region      镜像源地区
  --priority    镜像源优先级（数字越小优先级越高）

示例：
  go-version mirror add --name mycompany --url https://mirrors.mycompany.com/golang/ --description "公司内部镜像" --region "内网" --priority 1`,
	Run: runMirrorAddCommand,
}

var mirrorRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "移除自定义镜像源",
	Long: `移除指定的自定义镜像源。

注意：只能移除自定义添加的镜像源，不能移除预设的镜像源。

必需参数：
  --name    要移除的镜像源名称`,
	Run: runMirrorRemoveCommand,
}

var mirrorValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "验证镜像源的可用性",
	Long: `验证指定镜像源是否可用。

必需参数：
  --name    要验证的镜像源名称`,
	Run: runMirrorValidateCommand,
}

func init() {
	// 添加子命令
	mirrorCmd.AddCommand(mirrorListCmd)
	mirrorCmd.AddCommand(mirrorTestCmd)
	mirrorCmd.AddCommand(mirrorFastestCmd)
	mirrorCmd.AddCommand(mirrorAddCmd)
	mirrorCmd.AddCommand(mirrorRemoveCmd)
	mirrorCmd.AddCommand(mirrorValidateCmd)

	// mirror list 命令选项
	mirrorListCmd.Flags().BoolVar(&mirrorShowDetails, "details", false, "显示详细信息")
	mirrorListCmd.Flags().StringVar(&mirrorConfigPath, "config", "", "指定配置文件路径")

	// mirror test 命令选项
	mirrorTestCmd.Flags().StringVar(&mirrorName, "name", "", "指定要测试的镜像源名称")
	mirrorTestCmd.Flags().IntVar(&mirrorTestTimeout, "timeout", 30, "测试超时时间（秒）")
	mirrorTestCmd.Flags().BoolVar(&mirrorForceTest, "force", false, "强制测试，忽略缓存结果")

	// mirror fastest 命令选项
	mirrorFastestCmd.Flags().IntVar(&mirrorTestTimeout, "timeout", 30, "测试超时时间（秒）")
	mirrorFastestCmd.Flags().BoolVar(&mirrorShowDetails, "details", false, "显示测试详情")

	// mirror add 命令选项
	mirrorAddCmd.Flags().StringVar(&mirrorName, "name", "", "镜像源名称（必需）")
	mirrorAddCmd.Flags().StringVar(&mirrorURL, "url", "", "镜像源URL（必需）")
	mirrorAddCmd.Flags().StringVar(&mirrorDescription, "description", "", "镜像源描述（必需）")
	mirrorAddCmd.Flags().StringVar(&mirrorRegion, "region", "", "镜像源地区（必需）")
	mirrorAddCmd.Flags().IntVar(&mirrorPriority, "priority", 100, "镜像源优先级")
	mirrorAddCmd.MarkFlagRequired("name")
	mirrorAddCmd.MarkFlagRequired("url")
	mirrorAddCmd.MarkFlagRequired("description")
	mirrorAddCmd.MarkFlagRequired("region")

	// mirror remove 命令选项
	mirrorRemoveCmd.Flags().StringVar(&mirrorName, "name", "", "要移除的镜像源名称（必需）")
	mirrorRemoveCmd.MarkFlagRequired("name")

	// mirror validate 命令选项
	mirrorValidateCmd.Flags().StringVar(&mirrorName, "name", "", "要验证的镜像源名称（必需）")
	mirrorValidateCmd.MarkFlagRequired("name")
}

func runMirrorListCommand(cmd *cobra.Command, args []string) {
	// 创建镜像服务
	var mirrorService service.MirrorService
	var err error

	if mirrorConfigPath != "" {
		mirrorService, err = service.NewMirrorServiceWithConfig(mirrorConfigPath)
		if err != nil {
			PrintError(fmt.Sprintf("加载镜像配置失败: %s", err))
			os.Exit(1)
		}
	} else {
		mirrorService = service.NewMirrorService()
	}

	// 获取所有镜像
	mirrors := mirrorService.GetAvailableMirrors()

	if len(mirrors) == 0 {
		PrintInfo("没有可用的镜像源")
		return
	}

	PrintInfo("可用的镜像源:")
	PrintInfo("")

	for i, mirror := range mirrors {
		if mirrorShowDetails {
			PrintInfo(fmt.Sprintf("%d. %s", i+1, mirror.Name))
			PrintInfo(fmt.Sprintf("   描述: %s", mirror.Description))
			PrintInfo(fmt.Sprintf("   地区: %s", mirror.Region))
			PrintInfo(fmt.Sprintf("   URL: %s", mirror.BaseURL))
			PrintInfo(fmt.Sprintf("   优先级: %d", mirror.Priority))
			PrintInfo("")
		} else {
			PrintInfo(fmt.Sprintf("  %-12s %s (%s)", mirror.Name, mirror.Description, mirror.Region))
			PrintInfo(fmt.Sprintf("               %s", mirror.BaseURL))
			PrintInfo("")
		}
	}

	PrintInfo("使用示例:")
	PrintInfo("  go-version install 1.21.0 --mirror goproxy-cn    # 使用指定镜像安装")
	PrintInfo("  go-version install 1.21.0 --auto-mirror          # 自动选择最快镜像")
	PrintInfo("  go-version mirror test --name goproxy-cn         # 测试指定镜像")
	PrintInfo("  go-version mirror fastest                        # 选择最快镜像")
}

func runMirrorTestCommand(cmd *cobra.Command, args []string) {
	// 创建镜像服务
	mirrorService := service.NewMirrorService()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(mirrorTestTimeout)*time.Second)
	defer cancel()

	// 获取要测试的镜像
	var mirrors []service.Mirror
	if mirrorName != "" {
		// 测试指定镜像
		mirror, err := mirrorService.GetMirrorByName(mirrorName)
		if err != nil {
			PrintError(fmt.Sprintf("找不到镜像源 '%s': %s", mirrorName, err))
			os.Exit(1)
		}
		mirrors = []service.Mirror{*mirror}
	} else {
		// 测试所有镜像
		mirrors = mirrorService.GetAvailableMirrors()
	}

	if len(mirrors) == 0 {
		PrintInfo("没有可用的镜像源进行测试")
		return
	}

	PrintInfo("正在测试镜像源...")
	PrintInfo("")

	// 测试镜像
	results := make([]*service.MirrorTestResult, 0, len(mirrors))
	for _, mirror := range mirrors {
		PrintInfo(fmt.Sprintf("测试 %s (%s)...", mirror.Name, mirror.Description))

		result, err := mirrorService.TestMirrorSpeed(ctx, mirror)
		if err != nil {
			PrintError(fmt.Sprintf("  测试失败: %s", err))
			continue
		}

		results = append(results, result)

		if result.Available {
			PrintSuccess(fmt.Sprintf("  ✅ 可用 (响应时间: %v)", result.ResponseTime))
		} else {
			errorMsg := "未知错误"
			if result.Error != nil {
				errorMsg = result.Error.Error()
			}
			PrintError(fmt.Sprintf("  ❌ 不可用 (%s)", errorMsg))
		}
	}

	// 显示测试总结
	if len(results) > 1 {
		PrintInfo("")
		PrintInfo("测试结果总结:")

		// 按响应时间排序
		sort.Slice(results, func(i, j int) bool {
			if !results[i].Available && results[j].Available {
				return false
			}
			if results[i].Available && !results[j].Available {
				return true
			}
			if results[i].Available && results[j].Available {
				return results[i].ResponseTime < results[j].ResponseTime
			}
			return false
		})

		for i, result := range results {
			if result.Available {
				PrintInfo(fmt.Sprintf("  %d. %s - %v", i+1, result.Mirror.Name, result.ResponseTime))
			} else {
				PrintInfo(fmt.Sprintf("  %d. %s - 不可用", i+1, result.Mirror.Name))
			}
		}
	}
}

func runMirrorFastestCommand(cmd *cobra.Command, args []string) {
	// 创建镜像服务
	mirrorService := service.NewMirrorService()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(mirrorTestTimeout)*time.Second)
	defer cancel()

	// 获取所有镜像
	mirrors := mirrorService.GetAvailableMirrors()

	if len(mirrors) == 0 {
		PrintInfo("没有可用的镜像源")
		return
	}

	PrintInfo("正在测试所有镜像源以选择最快的...")
	PrintInfo("")

	if mirrorShowDetails {
		// 显示详细测试过程
		for _, mirror := range mirrors {
			PrintInfo(fmt.Sprintf("测试 %s...", mirror.Name))
		}
		PrintInfo("")
	}

	// 选择最快镜像
	fastest, err := mirrorService.SelectFastestMirror(ctx, mirrors)
	if err != nil {
		PrintError(fmt.Sprintf("选择最快镜像失败: %s", err))
		os.Exit(1)
	}

	PrintSuccess(fmt.Sprintf("最快的镜像源: %s", fastest.Name))
	PrintInfo(fmt.Sprintf("描述: %s", fastest.Description))
	PrintInfo(fmt.Sprintf("地区: %s", fastest.Region))
	PrintInfo(fmt.Sprintf("URL: %s", fastest.BaseURL))
	PrintInfo("")
	PrintInfo("使用此镜像的示例:")
	PrintInfo(fmt.Sprintf("  go-version install 1.21.0 --mirror %s", fastest.Name))
}

func runMirrorAddCommand(cmd *cobra.Command, args []string) {
	// 验证URL格式
	if mirrorURL == "" || (!startsWith(mirrorURL, "http://") && !startsWith(mirrorURL, "https://")) {
		PrintError("URL必须以 http:// 或 https:// 开头")
		os.Exit(1)
	}

	// 创建镜像服务（需要配置路径来保存自定义镜像）
	configPath := mirrorConfigPath
	if configPath == "" {
		// 使用默认配置路径
		configPath = getDefaultMirrorConfigPath()
	}

	mirrorService, err := service.NewMirrorServiceWithConfig(configPath)
	if err != nil {
		// 如果配置文件不存在，创建新的服务
		mirrorService = service.NewMirrorService()
	}

	// 检查镜像名称是否已存在
	existing, err := mirrorService.GetMirrorByName(mirrorName)
	if err == nil && existing != nil {
		PrintError(fmt.Sprintf("镜像源 '%s' 已存在", mirrorName))
		os.Exit(1)
	}

	// 创建新镜像
	newMirror := service.Mirror{
		Name:        mirrorName,
		BaseURL:     mirrorURL,
		Description: mirrorDescription,
		Region:      mirrorRegion,
		Priority:    mirrorPriority,
	}

	// 验证镜像可用性
	PrintInfo(fmt.Sprintf("验证镜像源 '%s' 的可用性...", mirrorName))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = mirrorService.ValidateMirror(ctx, newMirror)
	if err != nil {
		PrintError(fmt.Sprintf("镜像源验证失败: %s", err))
		PrintInfo("是否仍要添加此镜像源？输入 'y' 继续，其他任意键取消:")

		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			PrintInfo("已取消添加镜像源")
			return
		}
	} else {
		PrintSuccess("镜像源验证通过")
	}

	// TODO: 实际保存逻辑需要在MirrorService中实现
	// 添加到服务中
	err = mirrorService.AddCustomMirror(newMirror)
	if err != nil {
		PrintError(fmt.Sprintf("添加镜像源失败: %s", err))
		os.Exit(1)
	}

	// 保存配置
	err = mirrorService.SaveConfig(configPath)
	if err != nil {
		PrintError(fmt.Sprintf("保存配置失败: %s", err))
		os.Exit(1)
	}

	PrintSuccess(fmt.Sprintf("成功添加自定义镜像源: %s", mirrorName))
	PrintInfo(fmt.Sprintf("名称: %s", newMirror.Name))
	PrintInfo(fmt.Sprintf("URL: %s", newMirror.BaseURL))
	PrintInfo(fmt.Sprintf("描述: %s", newMirror.Description))
	PrintInfo(fmt.Sprintf("地区: %s", newMirror.Region))
	PrintInfo(fmt.Sprintf("优先级: %d", newMirror.Priority))
}

func runMirrorRemoveCommand(cmd *cobra.Command, args []string) {
	// 创建镜像服务
	configPath := mirrorConfigPath
	if configPath == "" {
		configPath = getDefaultMirrorConfigPath()
	}

	mirrorService, err := service.NewMirrorServiceWithConfig(configPath)
	if err != nil {
		PrintError(fmt.Sprintf("加载镜像配置失败: %s", err))
		os.Exit(1)
	}

	// 检查镜像是否存在
	mirror, err := mirrorService.GetMirrorByName(mirrorName)
	if err != nil {
		PrintError(fmt.Sprintf("找不到镜像源 '%s'", mirrorName))
		os.Exit(1)
	}

	// 检查是否为预设镜像（预设镜像不能删除）
	defaultMirrors := []string{"official", "goproxy-cn", "aliyun", "tencent", "huawei"}
	for _, defaultName := range defaultMirrors {
		if mirrorName == defaultName {
			PrintError(fmt.Sprintf("不能移除预设镜像源 '%s'", mirrorName))
			os.Exit(1)
		}
	}

	PrintInfo(fmt.Sprintf("即将移除镜像源: %s (%s)", mirror.Name, mirror.Description))
	PrintInfo("确认移除？输入 'y' 继续，其他任意键取消:")

	var response string
	fmt.Scanln(&response)
	if response != "y" && response != "Y" {
		PrintInfo("已取消移除操作")
		return
	}

	// TODO: 实际移除逻辑需要在MirrorService中实现
	// 从服务中移除
	err = mirrorService.RemoveCustomMirror(mirrorName)
	if err != nil {
		PrintError(fmt.Sprintf("移除镜像源失败: %s", err))
		os.Exit(1)
	}

	// 保存配置
	err = mirrorService.SaveConfig(configPath)
	if err != nil {
		PrintError(fmt.Sprintf("保存配置失败: %s", err))
		os.Exit(1)
	}

	PrintSuccess(fmt.Sprintf("成功移除镜像源: %s", mirrorName))
}

func runMirrorValidateCommand(cmd *cobra.Command, args []string) {
	// 创建镜像服务
	mirrorService := service.NewMirrorService()

	// 获取指定镜像
	mirror, err := mirrorService.GetMirrorByName(mirrorName)
	if err != nil {
		PrintError(fmt.Sprintf("找不到镜像源 '%s': %s", mirrorName, err))
		os.Exit(1)
	}

	PrintInfo(fmt.Sprintf("验证镜像源 '%s' (%s)...", mirror.Name, mirror.Description))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = mirrorService.ValidateMirror(ctx, *mirror)
	if err != nil {
		PrintError(fmt.Sprintf("验证失败: %s", err))
		os.Exit(1)
	}

	PrintSuccess(fmt.Sprintf("镜像源 '%s' 验证通过，可正常使用", mirror.Name))
	PrintInfo(fmt.Sprintf("URL: %s", mirror.BaseURL))
}

// 辅助函数
func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func getDefaultMirrorConfigPath() string {
	// 返回默认的镜像配置文件路径
	// 这里简化处理，实际应该根据操作系统确定合适的配置目录
	return "mirrors.json"
}
