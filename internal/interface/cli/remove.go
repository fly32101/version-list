package cli

import (
	"fmt"
	"os"

	"version-list/internal/application"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [version]",
	Short: "移除指定版本的Go",
	Long:  `移除指定版本的Go，例如: go-version remove 1.21.0`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		version := args[0]

		appService, err := application.NewVersionAppService()
		if err != nil {
			fmt.Fprintf(os.Stderr, "初始化应用服务失败: %s", err)
			os.Exit(1)
		}

		fmt.Printf("正在移除Go %s...", version)
		err = appService.Remove(version)
		if err != nil {
			fmt.Fprintf(os.Stderr, "移除失败: %s", err)
			os.Exit(1)
		}

		fmt.Printf("Go %s 已成功移除", version)
	},
}
