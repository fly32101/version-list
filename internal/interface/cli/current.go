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
			fmt.Fprintf(os.Stderr, "初始化应用服务失败: %s\n", err)
			os.Exit(1)
		}

		version, err := appService.Current()
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取当前版本失败: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("当前使用的Go版本: %s\n", version.Version)
		if version.Path != "" {
			fmt.Printf("安装路径: %s\n", version.Path)
		}
	},
}
