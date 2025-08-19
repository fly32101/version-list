package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"version-list/internal/application"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "显示当前使用的Go版本",
	Long:  `显示当前使用的Go版本及其路径`,
	Run: func(cmd *cobra.Command, args []string) {
		appService, err := application.NewVersionAppService()
		if err != nil {
			PrintError(fmt.Sprintf("初始化应用服务失败: %s", err))
			os.Exit(1)
		}

		version, err := appService.Current()
		if err != nil {
			PrintError(fmt.Sprintf("获取当前版本失败: %s", err))
			os.Exit(1)
		}

		PrintInfo(fmt.Sprintf("当前使用的Go版本: %s", version.Version))
		if version.Path != "" {
			fmt.Printf("安装路径: %s\n", version.Path)
		}
	},
}
