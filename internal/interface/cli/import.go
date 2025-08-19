package cli

import (
	"fmt"
	"os"

	"version-list/internal/application"

	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import [path]",
	Short: "导入本地已安装的Go版本",
	Long:  `导入本地已安装的Go版本，例如: go-version import "C:\Go"`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]

		appService, err := application.NewVersionAppService()
		if err != nil {
			PrintError(fmt.Sprintf("初始化应用服务失败: %s", err))
			os.Exit(1)
		}

		PrintInfo(fmt.Sprintf("正在导入本地Go版本从路径: %s", path))
		version, err := appService.ImportLocal(path)
		if err != nil {
			PrintError(fmt.Sprintf("导入失败: %s", err))
			os.Exit(1)
		}

		PrintSuccess(fmt.Sprintf("成功导入Go版本: %s", version))
	},
}
