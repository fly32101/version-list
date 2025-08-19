package cli

import (
	"fmt"
	"os"

	"version-list/internal/application"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install [version]",
	Short: "安装指定版本的Go",
	Long:  `安装指定版本的Go，例如: go-version install 1.21.0`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		version := args[0]

		appService, err := application.NewVersionAppService()
		if err != nil {
			PrintError(fmt.Sprintf("初始化应用服务失败: %s", err))
			os.Exit(1)
		}

		PrintInfo(fmt.Sprintf("正在安装Go %s...", version))
		err = appService.Install(version)
		if err != nil {
			PrintError(fmt.Sprintf("安装失败: %s", err))
			os.Exit(1)
		}

		PrintSuccess(fmt.Sprintf("Go %s 安装成功", version))
	},
}
