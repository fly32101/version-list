package cli

import (
	"fmt"
	"os"

	"version-list/internal/application"

	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use [version]",
	Short: "切换到指定版本的Go",
	Long:  `切换到指定版本的Go，例如: go-version use 1.21.0`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		version := args[0]

		appService, err := application.NewVersionAppService()
		if err != nil {
			PrintError(fmt.Sprintf("初始化应用服务失败: %s", err))
			os.Exit(1)
		}

		PrintInfo(fmt.Sprintf("正在切换到Go %s...", version))
		err = appService.Use(version)
		if err != nil {
			PrintError(fmt.Sprintf("切换失败: %s", err))
			os.Exit(1)
		}

		PrintSuccess(fmt.Sprintf("已成功切换到Go %s", version))
		PrintInfo("符号链接已创建，无需重启终端")
	},
}
