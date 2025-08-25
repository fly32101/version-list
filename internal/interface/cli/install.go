package cli

import (
	"fmt"
	"os"

	"version-list/internal/application"
	"version-list/internal/domain/model"
	"version-list/internal/interface/ui"

	"github.com/spf13/cobra"
)

// 命令行选项变量
var (
	installPath      string
	forceInstall     bool
	skipVerification bool
	installTimeout   int
	maxRetries       int
	noProgress       bool
	onlineInstall    bool
	mirrorName       string
	autoMirror       bool
	listMirrors      bool
)

var installCmd = &cobra.Command{
	Use:   "install [version]",
	Short: "安装指定版本的Go",
	Long: `安装指定版本的Go。

支持两种安装模式：
1. 在线安装（默认）：从Go官网或镜像源自动下载并安装指定版本
2. 本地安装：仅注册已存在的Go版本到版本管理器

镜像源支持：
- 使用 --mirror 指定镜像源（official, goproxy-cn, aliyun, tencent, huawei）
- 使用 --auto-mirror 自动选择最快的镜像源
- 使用 --list-mirrors 查看所有可用的镜像源

示例：
  go-version install 1.21.0                           # 在线安装Go 1.21.0（使用官方源）
  go-version install 1.21.0 --mirror goproxy-cn      # 使用七牛云镜像安装
  go-version install 1.21.0 --auto-mirror            # 自动选择最快镜像安装
  go-version install 1.21.0 --path /custom           # 安装到自定义路径
  go-version install 1.21.0 --force                  # 强制重新安装
  go-version install 1.21.0 --no-progress            # 不显示进度条
  go-version install --list-mirrors                  # 查看可用镜像源`,
	Args: cobra.RangeArgs(0, 1),
	Run:  runInstallCommand,
}

func init() {
	// 添加命令行选项
	installCmd.Flags().StringVar(&installPath, "path", "", "自定义安装路径")
	installCmd.Flags().BoolVar(&forceInstall, "force", false, "强制重新安装（即使版本已存在）")
	installCmd.Flags().BoolVar(&skipVerification, "skip-verification", false, "跳过文件完整性验证")
	installCmd.Flags().IntVar(&installTimeout, "timeout", 300, "安装超时时间（秒）")
	installCmd.Flags().IntVar(&maxRetries, "max-retries", 3, "最大重试次数")
	installCmd.Flags().BoolVar(&noProgress, "no-progress", false, "不显示进度条")
	installCmd.Flags().BoolVar(&onlineInstall, "online", true, "在线安装模式（默认）")

	// 镜像相关选项
	installCmd.Flags().StringVar(&mirrorName, "mirror", "", "指定镜像源 (official, goproxy-cn, aliyun, tencent, huawei)")
	installCmd.Flags().BoolVar(&autoMirror, "auto-mirror", false, "自动选择最快的镜像源")
	installCmd.Flags().BoolVar(&listMirrors, "list-mirrors", false, "显示所有可用的镜像源")
}

func runInstallCommand(cmd *cobra.Command, args []string) {
	// 处理 --list-mirrors 选项
	if listMirrors {
		runListMirrors()
		return
	}

	// 检查是否提供了版本号
	if len(args) == 0 {
		PrintError("请指定要安装的Go版本号")
		PrintInfo("使用示例: go-version install 1.21.0")
		PrintInfo("使用 --list-mirrors 查看可用镜像源")
		os.Exit(1)
	}

	version := args[0]

	// 验证版本号格式
	if !isValidVersion(version) {
		PrintError(fmt.Sprintf("无效的版本号格式: %s", version))
		PrintInfo("版本号格式示例: 1.21.0, 1.20.5")
		os.Exit(1)
	}

	// 验证镜像选项
	if err := validateMirrorOptions(); err != nil {
		PrintError(err.Error())
		os.Exit(1)
	}

	// 创建应用服务
	appService, err := application.NewVersionAppService()
	if err != nil {
		PrintError(fmt.Sprintf("初始化应用服务失败: %s", err))
		os.Exit(1)
	}

	if onlineInstall {
		runOnlineInstall(appService, version)
	} else {
		runLocalInstall(appService, version)
	}
}

func runOnlineInstall(appService *application.VersionAppService, version string) {
	PrintInfo(fmt.Sprintf("开始在线安装Go %s...", version))

	// 创建安装选项
	options := &model.InstallOptions{
		Force:            forceInstall,
		CustomPath:       installPath,
		SkipVerification: skipVerification,
		Timeout:          installTimeout,
		MaxRetries:       maxRetries,
		Mirror:           mirrorName,
		AutoMirror:       autoMirror,
	}

	// 创建进度UI
	var progressUI *ui.InstallProgressUI
	if !noProgress {
		progressUI = ui.NewInstallProgressUI()
		progressUI.Start()
		defer progressUI.Stop()
	}

	// 执行在线安装
	result, err := appService.InstallOnline(version, options, progressUI)
	if err != nil {
		if progressUI != nil {
			progressUI.PrintError(fmt.Sprintf("安装失败: %s", err))
		} else {
			PrintError(fmt.Sprintf("安装失败: %s", err))
		}
		os.Exit(1)
	}

	// 显示安装结果
	displayInstallResult(result, progressUI)
}

func runLocalInstall(appService *application.VersionAppService, version string) {
	PrintInfo(fmt.Sprintf("正在注册本地Go版本 %s...", version))

	if installPath == "" {
		PrintError("本地安装模式需要指定 --path 参数")
		os.Exit(1)
	}

	// 执行本地导入
	importedVersion, err := appService.ImportLocal(installPath)
	if err != nil {
		PrintError(fmt.Sprintf("导入本地版本失败: %s", err))
		os.Exit(1)
	}

	PrintSuccess(fmt.Sprintf("成功导入本地Go版本 %s", importedVersion))
	PrintInfo(fmt.Sprintf("安装路径: %s", installPath))
}

func displayInstallResult(result *model.InstallationResult, progressUI *ui.InstallProgressUI) {
	if result.Success {
		message := fmt.Sprintf("Go %s 安装成功!", result.Version)
		if progressUI != nil {
			progressUI.PrintSuccess(message)
		} else {
			PrintSuccess(message)
		}

		// 显示详细信息
		PrintInfo(fmt.Sprintf("安装路径: %s", result.Path))
		PrintInfo(fmt.Sprintf("安装耗时: %v", result.Duration))

		if result.DownloadInfo != nil {
			downloadSize := formatBytes(result.DownloadInfo.Size)
			downloadSpeed := formatBytes(int64(result.DownloadInfo.Speed)) + "/s"
			PrintInfo(fmt.Sprintf("下载大小: %s (平均速度: %s)", downloadSize, downloadSpeed))
		}

		if result.ExtractInfo != nil {
			PrintInfo(fmt.Sprintf("解压文件数: %d", result.ExtractInfo.FileCount))
		}

		PrintInfo("使用 'go-version use " + result.Version + "' 来切换到此版本")
	} else {
		message := fmt.Sprintf("Go %s 安装失败: %s", result.Version, result.Error)
		if progressUI != nil {
			progressUI.PrintError(message)
		} else {
			PrintError(message)
		}
	}
}

func isValidVersion(version string) bool {
	// 简单的版本号验证
	if len(version) == 0 {
		return false
	}

	// 检查是否包含基本的版本号格式（数字和点）
	for _, char := range version {
		if !((char >= '0' && char <= '9') || char == '.' || char == '-' || (char >= 'a' && char <= 'z')) {
			return false
		}
	}

	return true
}

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

// runListMirrors 显示所有可用的镜像源
func runListMirrors() {
	PrintInfo("可用的镜像源:")
	PrintInfo("")

	mirrors := []struct {
		Name        string
		Description string
		Region      string
		URL         string
	}{
		{"official", "Go官方下载源", "全球", "https://golang.org/dl/"},
		{"goproxy-cn", "七牛云Go代理镜像", "中国", "https://goproxy.cn/golang/"},
		{"aliyun", "阿里云镜像源", "中国", "https://mirrors.aliyun.com/golang/"},
		{"tencent", "腾讯云镜像源", "中国", "https://mirrors.cloud.tencent.com/golang/"},
		{"huawei", "华为云镜像源", "中国", "https://mirrors.huaweicloud.com/golang/"},
	}

	for _, mirror := range mirrors {
		PrintInfo(fmt.Sprintf("  %-12s %s (%s)", mirror.Name, mirror.Description, mirror.Region))
		PrintInfo(fmt.Sprintf("               %s", mirror.URL))
		PrintInfo("")
	}

	PrintInfo("使用示例:")
	PrintInfo("  go-version install 1.21.0 --mirror goproxy-cn    # 使用七牛云镜像")
	PrintInfo("  go-version install 1.21.0 --auto-mirror          # 自动选择最快镜像")
}

// validateMirrorOptions 验证镜像选项
func validateMirrorOptions() error {
	// 检查互斥选项
	if mirrorName != "" && autoMirror {
		return fmt.Errorf("--mirror 和 --auto-mirror 选项不能同时使用")
	}

	// 验证镜像名称
	if mirrorName != "" {
		validMirrors := []string{"official", "goproxy-cn", "aliyun", "tencent", "huawei"}
		isValid := false
		for _, valid := range validMirrors {
			if mirrorName == valid {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("无效的镜像名称: %s。可用镜像: %v", mirrorName, validMirrors)
		}
	}

	return nil
}
